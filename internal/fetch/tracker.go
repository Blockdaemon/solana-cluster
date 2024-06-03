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

// Package fetch contains a tracker API client.
package fetch

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	"go.blockdaemon.com/solana/cluster-manager/types"
	"gopkg.in/resty.v1"
)

// TrackerClient accesses the tracker API.
type TrackerClient struct {
	resty *resty.Client
}

func NewTrackerClient(trackerURL string) *TrackerClient {
	return NewTrackerClientWithResty(resty.New().SetHostURL(trackerURL))
}

func NewTrackerClientWithResty(client *resty.Client) *TrackerClient {
	return &TrackerClient{resty: client}
}

func (c *TrackerClient) GetBestSnapshots(ctx context.Context, count int) (sources []types.SnapshotSource, err error) {
	res, err := c.resty.R().
		SetContext(ctx).
		SetHeader("accept", "application/json").
		SetQueryParam("max", strconv.Itoa(count)).
		SetResult(&sources).
		Get("/v1/best_snapshots")
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("get best snapshots: %s", res.Status())
	}
	return
}

func (c *TrackerClient) GetSnapshotAtSlot(ctx context.Context, slot uint64) (sources []types.SnapshotSource, err error) {
	res, err := c.resty.R().
		SetContext(ctx).
		SetHeader("accept", "application/json").
		SetQueryParam("slot", strconv.FormatUint(slot, 10)).
		SetResult(&sources).
		Get("/v1/snapshots")
	spew.Dump(sources)
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("get snapshots at slot %d: %s", slot, res.Status())
	}
	return
}
