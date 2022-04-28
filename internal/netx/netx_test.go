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

package netx

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListenTCPInterface(t *testing.T) {
	// Find loopback interface.
	ifaces, err := net.Interfaces()
	require.NoError(t, err)
	var lo string
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			lo = iface.Name
			break
		}
	}
	if lo == "" {
		t.Skip("Could not find loopback interface")
	}

	// Bind to loopback.
	t.Logf("Binding to %v", lo)
	listener, addrs, err := ListenTCPInterface("tcp", lo, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, addrs)
	t.Log("Addresses", addrs)
	require.NoError(t, listener.Close())
}

// TestMergeListeners ensures an HTTP server can accept conns from multiple listeners.
func TestMergeListeners(t *testing.T) {
	l1, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	l2, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	merged := MergeListeners(l1, l2)
	assert.Equal(t, l1.Addr(), merged.Addr())

	server := http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
	}
	go server.Serve(merged)

	pingUntilAwake(t, "http://"+l1.Addr().String())

	_, err = http.DefaultClient.Get("http://" + l1.Addr().String())
	require.NoError(t, err)
	_, err = http.DefaultClient.Get("http://" + l2.Addr().String())
	require.NoError(t, err)

	require.NoError(t, merged.Close())
	time.Sleep(100 * time.Millisecond)
	assert.True(t, errors.Is(l1.Close(), net.ErrClosed))
	assert.True(t, errors.Is(l2.Close(), net.ErrClosed))
}

// TestMergeListeners_ExternalClose ensures a merged listener shuts down when any of its listeners fail.
func TestMergeListeners_ExternalClose(t *testing.T) {
	l1, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	l2, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	merged := MergeListeners(l1, l2)
	assert.Equal(t, l1.Addr(), merged.Addr())

	server := http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
	}
	go server.Serve(merged)

	pingUntilAwake(t, "http://"+l1.Addr().String())

	// Externally close one of the listeners.
	require.NoError(t, l1.Close())

	time.Sleep(100 * time.Millisecond)

	assert.True(t, errors.Is(l1.Close(), net.ErrClosed))
	assert.True(t, errors.Is(l2.Close(), net.ErrClosed))
	require.True(t, errors.Is(merged.Close(), net.ErrClosed))
}

// TestMergeListeners_CloseSlowAcceptor ensures a merged listener shuts down when net.Listener::Accept() has never returned.
func TestMergeListeners_CloseSlowAcceptor(t *testing.T) {
	l1, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	l2, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	merged := MergeListeners(l1, l2)
	assert.Equal(t, l1.Addr(), merged.Addr())

	time.Sleep(200 * time.Millisecond)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := net.Dial("tcp", l1.Addr().String())
		assert.NoError(t, err)
	}()

	time.Sleep(200 * time.Millisecond)

	require.NoError(t, merged.Close())
	wg.Wait()
	time.Sleep(200 * time.Millisecond)
	assert.True(t, errors.Is(l1.Close(), net.ErrClosed))
	assert.True(t, errors.Is(l2.Close(), net.ErrClosed))
}

func pingUntilAwake(t *testing.T, url string) {
	const timeout = 3 * time.Second
	const interval = 20 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		time.Sleep(interval)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		require.NoError(t, err)
		_, err = http.DefaultClient.Do(req)
		if errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		} else if err != nil {
			continue
		}
		return
	}
}
