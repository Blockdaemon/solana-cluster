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
	"os"

	"go.blockdaemon.com/solana/cluster-manager/types"
	"gopkg.in/resty.v1"
)

// TODO rewrite this package to use OpenAPI code gen

// SidecarClient accesses the sidecar API.
type SidecarClient struct {
	resty         *resty.Client
	mkProxyReader func(name string, size int64, rd io.Reader) io.ReadCloser
}

func NewSidecarClient(sidecarURL string) *SidecarClient {
	return NewSidecarClientWithResty(resty.New().SetHostURL(sidecarURL))
}

func NewSidecarClientWithResty(client *resty.Client) *SidecarClient {
	return &SidecarClient{
		resty: client,
		mkProxyReader: func(_ string, _ int64, rd io.Reader) io.ReadCloser {
			return io.NopCloser(rd)
		},
	}
}

func (c *SidecarClient) SetProxyReaderFunc(m func(name string, size int64, rd io.Reader) io.ReadCloser) {
	c.mkProxyReader = m
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
	if err := expectOK(res.RawResponse); err != nil {
		return nil, err
	}
	return
}

func (c *SidecarClient) DownloadSnapshotFile(ctx context.Context, name string) error {
	// Open temporary file. (Consider using O_TMPFILE)
	f, err := os.Create(".tmp." + name)
	if err != nil {
		return err
	}
	defer f.Close()

	// Start new snapshot request.
	snapURL := c.resty.HostURL + "/" + name
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, snapURL, nil)
	if err != nil {
		return err
	}
	res, err := c.resty.GetClient().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if err := expectOK(res); err != nil {
		return err
	}
	if res.ContentLength < 0 {
		return fmt.Errorf("content length unknown")
	}

	// Download
	proxyRd := c.mkProxyReader(name, res.ContentLength, res.Body)
	_, err = io.Copy(f, proxyRd)
	if err != nil {
		_ = proxyRd.Close()
		return fmt.Errorf("download failed: %w", err)
	}
	_ = proxyRd.Close()

	// Promote temporary file.
	return os.Rename(f.Name(), name)
}

func expectOK(res *http.Response) error {
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("get snapshots: %s", res.Status)
	}
	return nil
}
