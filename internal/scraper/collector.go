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
	"time"

	"go.blockdaemon.com/solana/cluster-manager/internal/index"
	"go.blockdaemon.com/solana/cluster-manager/types"
)

// Collector streams probe results into the database.
type Collector struct {
	resChan chan ProbeResult
	DB      *index.DB
}

func NewCollector(db *index.DB) *Collector {
	this := &Collector{
		DB: db,
	}
	go this.run()
	return this
}

// Probes returns a send-channel that collects and indexes probe results.
func (c *Collector) Probes() chan<- ProbeResult {
	return c.resChan
}

// Close stops the collector and closes the send-channel.
func (c *Collector) Close() {
	close(c.resChan)
}

func (c *Collector) run() {
	for res := range c.resChan {
		c.DB.DeleteSnapshotsByTarget(res.Target)
		entries := make([]*index.SnapshotEntry, len(res.Infos))
		for i, info := range res.Infos {
			entries[i] = &index.SnapshotEntry{
				SnapshotKey: index.NewSnapshotKey(res.Target, info.Slot),
				Info:        info,
				UpdatedAt:   res.Time,
			}
		}
	}
}

type ProbeResult struct {
	Time   time.Time
	Target string
	Infos  []*types.SnapshotInfo
	Err    error
}
