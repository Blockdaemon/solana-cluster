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
	"sync"

	"go.blockdaemon.com/solana/cluster-manager/internal/discovery"
	"go.blockdaemon.com/solana/cluster-manager/types"
	"go.uber.org/zap"
)

// Manager maintains a group of scrapers.
type Manager struct {
	res      chan<- ProbeResult
	scrapers []*Scraper

	Log *zap.Logger
}

func NewManager(results chan<- ProbeResult) *Manager {
	return &Manager{
		res: results,

		Log: zap.NewNop(),
	}
}

// Reset shuts down all scrapers.
func (m *Manager) Reset() {
	var wg sync.WaitGroup
	wg.Add(len(m.scrapers))
	for _, scraper := range m.scrapers {
		go func(scraper *Scraper) {
			defer wg.Done()
			scraper.Close()
		}(scraper)
	}
	m.scrapers = nil
}

// Update shuts down and reloads all scrapers from config.
func (m *Manager) Update(conf *types.Config) {
	m.Reset()
	for _, group := range conf.TargetGroups {
		log := m.Log.With(zap.String("group", group.Group))
		if err := m.loadGroup(group, log); err != nil {
			log.Error("Failed to load group", zap.Error(err))
		}
	}
	for _, scraper := range m.scrapers {
		scraper.Start(m.res, conf.ScrapeInterval)
	}
}

func (m *Manager) loadGroup(group *types.TargetGroup, log *zap.Logger) error {
	disc, err := discovery.NewFromConfig(group)
	if err != nil {
		return err
	}

	prober, err := NewProber(group)
	if err != nil {
		return err
	}

	scraper := NewScraper(prober, disc)
	scraper.Log = log
	m.scrapers = append(m.scrapers, scraper)

	return nil
}
