// Package netx contains network hacks
package netx

import (
	"context"
	"fmt"
	"net"

	"go.uber.org/multierr"
	"golang.org/x/sync/errgroup"
)

// MergeListeners merges multiple net.Listeners into one. Trips on the first error seen.
func MergeListeners(listeners ...net.Listener) net.Listener {
	merged := &mergedListener{listeners: listeners}
	merged.start()
	return merged
}

type mergedListener struct {
	listeners []net.Listener
	ctx       context.Context
	conns     chan net.Conn
}

func (m *mergedListener) start() {
	group, ctx := errgroup.WithContext(context.Background())
	m.ctx = ctx
	m.conns = make(chan net.Conn)
	go func() {
		defer m.Close()
		<-ctx.Done()
	}()
	for _, listener := range m.listeners {
		listener_ := listener
		group.Go(func() error {
			for {
				accept, err := listener_.Accept()
				if err != nil {
					return err
				}
				select {
				case <-ctx.Done():
					return nil
				case m.conns <- accept:
				}
			}
		})
	}
}

func (m *mergedListener) Accept() (net.Conn, error) {
	select {
	case <-m.ctx.Done():
		return nil, net.ErrClosed
	case conn := <-m.conns:
		return conn, nil
	}
}

func (m *mergedListener) Close() error {
	errs := make([]error, 0, len(m.listeners))
	for _, listener := range m.listeners {
		if err := listener.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return multierr.Combine(errs...)
}

func (m *mergedListener) Addr() net.Addr {
	for _, listener := range m.listeners {
		if ta, ok := listener.Addr().(*net.TCPAddr); ok {
			if ta.IP.IsGlobalUnicast() {
				return ta
			}
		}
	}
	return m.listeners[0].Addr()
}

// ListenTCPInterface is like net.ListenTCP but can bind to one interface only.
func ListenTCPInterface(network string, ifaceName string, port uint16) (net.Listener, []net.TCPAddr, error) {
	var listenAddrs []net.TCPAddr
	if ifaceName != "" {
		iface, err := net.InterfaceByName(ifaceName)
		if err != nil {
			return nil, nil, err
		}
		ifaceAddrs, err := iface.Addrs()
		if err != nil {
			return nil, nil, err
		}
		for _, addr := range ifaceAddrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}
			listenAddrs = append(listenAddrs, net.TCPAddr{
				IP:   ip,
				Port: int(port),
				Zone: iface.Name,
			})
		}
	} else {
		listenAddrs = []net.TCPAddr{
			{
				IP:   net.IPv6zero, // dual-stack
				Port: int(port),
			},
		}
	}

	listeners := make([]net.Listener, len(listenAddrs))
	tcpAddrs := make([]net.TCPAddr, len(listenAddrs))
	for i, tcp := range listenAddrs {
		listen, err := net.ListenTCP(network, &tcp)
		if err != nil {
			return nil, nil, err
		}
		listeners[i] = listen
		tcpAddrs[i] = tcp
	}
	if len(listeners) == 0 {
		return nil, nil, fmt.Errorf("listen on %s: interface has no addresses", ifaceName)
	}
	return MergeListeners(listeners...), tcpAddrs, nil
}
