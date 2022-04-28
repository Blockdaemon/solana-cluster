// Copyright 2022 Blockdaemon Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package netx contains network hacks
package netx

import (
	"context"
	"fmt"
	"net"

	"go.uber.org/atomic"
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
	cancel    context.CancelFunc
	group     *errgroup.Group
	ctx       context.Context
	conns     chan net.Conn
	err       atomic.Error
}

func (m *mergedListener) start() {
	m.ctx, m.cancel = context.WithCancel(context.Background())
	m.group, m.ctx = errgroup.WithContext(m.ctx)
	m.conns = make(chan net.Conn)
	for _, listener := range m.listeners {
		listener_ := listener
		go func() {
			<-m.ctx.Done()
			if err := listener_.Close(); err != nil {
				m.err.Store(err)
			}
		}()
		m.group.Go(func() error {
			for {
				accept, err := listener_.Accept()
				if err != nil {
					if m.ctx.Err() != nil {
						return nil
					} else {
						return err
					}
				}
				select {
				case <-m.ctx.Done():
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
		err := net.ErrClosed
		if storedErr := m.err.Load(); storedErr != nil {
			err = storedErr
		}
		return nil, err
	case conn := <-m.conns:
		return conn, nil
	}
}

func (m *mergedListener) Close() (err error) {
	m.cancel()
	if err2 := m.group.Wait(); err2 != nil {
		return err2
	}
	return err
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
