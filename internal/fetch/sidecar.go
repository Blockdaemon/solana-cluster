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

// Package fetch contains a sidecar API client.
package fetch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"go.blockdaemon.com/solana/cluster-manager/types"
	"go.uber.org/zap"
	"gopkg.in/resty.v1"
)

// TODO rewrite this package to use OpenAPI code gen

// SidecarClient accesses the sidecar API.
type SidecarClient struct {
	resty           *resty.Client
	log             *zap.Logger
	proxyReaderFunc ProxyReaderFunc
}

type SidecarClientOpts struct {
	Resty           *resty.Client
	Log             *zap.Logger
	ProxyReaderFunc ProxyReaderFunc
}

type ProxyReaderFunc func(name string, size int64, rd io.Reader) io.ReadCloser

func NewSidecarClient(sidecarURL string) *SidecarClient {
	return NewSidecarClientWithOpts(sidecarURL, SidecarClientOpts{})
}

func NewSidecarClientWithOpts(sidecarURL string, opts SidecarClientOpts) *SidecarClient {
	if opts.Resty == nil {
		opts.Resty = resty.New()
	}
	opts.Resty.SetHostURL(sidecarURL)
	if opts.ProxyReaderFunc == nil {
		opts.ProxyReaderFunc = func(_ string, _ int64, rd io.Reader) io.ReadCloser {
			return io.NopCloser(rd)
		}
	}
	if opts.Log == nil {
		opts.Log = zap.NewNop()
	}
	return &SidecarClient{
		resty:           opts.Resty,
		log:             opts.Log,
		proxyReaderFunc: opts.ProxyReaderFunc,
	}
}

func (c *SidecarClient) ListSnapshots(ctx context.Context) (infos []*types.SnapshotInfo, err error) {
	res, err := c.resty.R().
		SetContext(ctx).
		SetHeader("accept", "application/json").
		SetResult(&infos).
		Get("/v1/snapshots")
	if err != nil {
		return nil, err
	}
	if err := expectOK(res.RawResponse, "list snapshots"); err != nil {
		return nil, err
	}
	return
}

// StreamSnapshot starts a download of a snapshot file.
// The returned response is guaranteed to have a valid ContentLength.
// The caller has the responsibility to close the response body even if the error is not nil.
func (c *SidecarClient) StreamSnapshot(ctx context.Context, name string) (res *http.Response, err error) {
	snapURL := c.resty.HostURL + "/v1/snapshot/" + url.PathEscape(name)
	c.log.Debug("Downloading snapshot", zap.String("snapshot_url", snapURL))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, snapURL, nil)
	if err != nil {
		return nil, err
	}
	res, err = c.resty.GetClient().Do(req)
	if err != nil {
		return
	}
	if err = expectOK(res, "download snapshot"); err != nil {
		return
	}
	if res.ContentLength < 0 {
		err = fmt.Errorf("content length unknown")
	}
	return
}

// DownloadSnapshotFile downloads a snapshot to a file in the local file system.
func (c *SidecarClient) DownloadSnapshotFile(ctx context.Context, destDir string, name string) error {
	res, err := c.StreamSnapshot(ctx, name)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return err
	}

	// Open temporary file. (Consider using O_TMPFILE)
	f, err := os.Create(filepath.Join(destDir, ".tmp."+name))
	if err != nil {
		return err
	}
	defer f.Close()

	// Download
	proxyRd := c.proxyReaderFunc(name, res.ContentLength, res.Body)
	_, err = io.Copy(f, proxyRd)
	if err != nil {
		_ = proxyRd.Close()
		return fmt.Errorf("download failed: %w", err)
	}
	_ = proxyRd.Close()

	// Promote temporary file.
	destPath := filepath.Join(destDir, name)
	err = os.Rename(f.Name(), destPath)
	if err != nil {
		return err
	}

	// Change modification time to what server said.
	modTime, err := time.Parse(http.TimeFormat, res.Header.Get("last-modified"))
	if err == nil && !modTime.IsZero() {
		_ = os.Chtimes(destPath, time.Now(), modTime)
	}

	return nil
}

func expectOK(res *http.Response, op string) error {
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", op, res.Status)
	}
	return nil
}
