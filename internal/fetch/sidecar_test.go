package fetch

import (
	"bytes"
	"context"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"gopkg.in/resty.v1"
)

func TestConnectError(t *testing.T) {
	client := NewSidecarClient("invalid://e")

	_, err := client.ListSnapshots(context.TODO())
	assert.Error(t, err)

	err = client.DownloadSnapshotFile(context.TODO(), "/nonexistent9", "snap")
	assert.EqualError(t, err, "Get \"invalid://e/v1/snapshot/snap\": unsupported protocol scheme \"invalid\"")
}

func TestInternalServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewSidecarClientWithOpts(server.URL, SidecarClientOpts{Resty: resty.NewWithClient(server.Client())})

	_, err := client.ListSnapshots(context.TODO())
	assert.EqualError(t, err, "list snapshots: 500 Internal Server Error")

	err = client.DownloadSnapshotFile(context.TODO(), "/nonexistent3", "bla")
	assert.EqualError(t, err, "download snapshot: 500 Internal Server Error")
}

func TestSidecarClient_DownloadSnapshotFile(t *testing.T) {
	const snapshotName = "bla.tar.zst"
	const size = 100

	// Start server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/snapshot/bla.tar.zst", r.URL.Path)
		w.Header().Set("content-length", "100")
		w.Header().Set("last-modified", "Wed, 01 Jan 2020 01:01:01 GMT")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bytes.Repeat([]byte{'A'}, size))
	}))
	defer server.Close()

	// Create client
	var proxyReader atomic.Value
	client := NewSidecarClientWithOpts(server.URL, SidecarClientOpts{
		Resty: resty.NewWithClient(server.Client()),
		ProxyReaderFunc: func(name string, size_ int64, rd io.Reader) io.ReadCloser {
			assert.Equal(t, snapshotName, name)
			assert.Equal(t, int64(size), size_)
			proxy := &mockReadCloser{rd: rd}
			assert.True(t, proxyReader.CompareAndSwap(nil, proxy))
			return proxy
		},
	})

	// Temp dir
	tmpDir, err := os.MkdirTemp("", "download_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Download snapshot to temp dir
	err = client.DownloadSnapshotFile(context.TODO(), tmpDir, snapshotName)
	require.NoError(t, err)

	// Ensure proxy was closed exactly once.
	assert.Equal(t, proxyReader.Load().(*mockReadCloser).closes.Load(), int32(1))

	// Make sure file details check out
	stat, err := os.Stat(filepath.Join(tmpDir, snapshotName))
	require.NoError(t, err)
	assert.Equal(t, int64(size), stat.Size())
	var modTime = time.Date(2020, 1, 1, 1, 1, 1, 0, time.UTC)
	assert.Less(t, math.Abs(modTime.Sub(stat.ModTime()).Seconds()), float64(2), "different mod times")
}

type mockReadCloser struct {
	rd     io.Reader
	closes atomic.Int32
}

func (m *mockReadCloser) Read(p []byte) (int, error) {
	return m.rd.Read(p)
}

func (m *mockReadCloser) Close() error {
	m.closes.Add(1)
	return nil
}
