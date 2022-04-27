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
	"net/http"

	"go.blockdaemon.com/solana/cluster-manager/types"
	"gopkg.in/resty.v1"
)

// TODO rewrite this package to use OpenAPI code gen

// Client accesses the sidecar API.
type Client struct {
	resty *resty.Client
}

func NewClient(sidecarURL string) *Client {
	return NewClientWithResty(resty.New().SetHostURL(sidecarURL))
}

func NewClientWithResty(client *resty.Client) *Client {
	return &Client{resty: client}
}

func (c *Client) ListSnapshots(ctx context.Context) (infos []*types.SnapshotInfo, err error) {
	res, err := c.resty.R().
		SetContext(ctx).
		SetHeader("accept", "application/json").
		SetResult(&infos).
		Get("/v1/snapshots")
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("get snapshots: %s", res.Status())
	}
	return
}
