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

package scraper

import (
	"sync/atomic"
	"time"

	"go.blockdaemon.com/solana/cluster-manager/internal/index"
	"go.blockdaemon.com/solana/cluster-manager/types"
	"go.uber.org/zap"
)

// Collector streams probe results into the database.
type Collector struct {
	resChan chan ProbeResult
	DB      *index.DB
	Log     *zap.Logger

	closed uint32
}

func NewCollector(db *index.DB) *Collector {
	this := &Collector{
		resChan: make(chan ProbeResult),
		DB:      db,
		Log:     zap.NewNop(),
	}
	return this
}

func (c *Collector) Start() {
	go c.run()
}

// Probes returns a send-channel that collects and indexes probe results.
func (c *Collector) Probes() chan<- ProbeResult {
	return c.resChan
}

// Close stops the collector and closes the send-channel.
func (c *Collector) Close() {
	if atomic.CompareAndSwapUint32(&c.closed, 0, 1) {
		close(c.resChan)
	}
}

func (c *Collector) run() {
	for res := range c.resChan {
		if res.Err != nil {
			c.Log.Warn("Scrape failed",
				zap.String("target", res.Target),
				zap.Error(res.Err))
			continue
		}
		c.Log.Debug("Scrape success",
			zap.String("target", res.Target),
			zap.Int("num_snapshots", len(res.Infos)))
		c.DB.DeleteSnapshotsByTarget(res.Target)
		entries := make([]*index.SnapshotEntry, len(res.Infos))
		for i, info := range res.Infos {
			entries[i] = &index.SnapshotEntry{
				SnapshotKey: index.NewSnapshotKey(res.Target, info.Slot),
				Info:        info,
				UpdatedAt:   res.Time,
			}
		}
		c.DB.UpsertSnapshots(entries...)
	}
}

type ProbeResult struct {
	Time   time.Time
	Target string
	Infos  []*types.SnapshotInfo
	Err    error
}
