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
	"context"
	"sync"
	"time"

	"go.blockdaemon.com/solana/cluster-manager/internal/discovery"
	"go.uber.org/zap"
)

type Scraper struct {
	prober     *Prober
	discoverer discovery.Discoverer
	rootCtx    context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup

	Log *zap.Logger
}

func NewScraper(prober *Prober, discoverer discovery.Discoverer) *Scraper {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scraper{
		prober:     prober,
		discoverer: discoverer,
		rootCtx:    ctx,
		cancel:     cancel,
		Log:        zap.NewNop(),
	}
}

func (s *Scraper) Start(results chan<- ProbeResult, interval time.Duration) {
	s.wg.Add(1)
	go s.run(results, interval)
}

func (s *Scraper) Close() {
	s.cancel()
	s.wg.Wait()
}

func (s *Scraper) run(results chan<- ProbeResult, interval time.Duration) {
	s.Log.Info("Starting scraper")
	defer s.Log.Info("Stopping scraper")

	defer s.wg.Done()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		ctx, cancel := context.WithCancel(s.rootCtx)
		go s.scrape(ctx, results)

		select {
		case <-s.rootCtx.Done():
			cancel()
			return
		case <-ticker.C:
			cancel()
		}
	}
}

func (s *Scraper) scrape(ctx context.Context, results chan<- ProbeResult) {
	discoveryStart := time.Now()
	targets, err := s.discoverer.DiscoverTargets(ctx)
	if err != nil {
		s.Log.Error("Service discovery failed", zap.Error(err))
		return
	}

	scrapeStart := time.Now()
	s.Log.Debug("Scrape starting",
		zap.Duration("discovery_duration", time.Since(discoveryStart)),
		zap.Int("num_targets", len(targets)))

	var wg sync.WaitGroup
	wg.Add(len(targets))
	for _, target := range targets {
		go func(target string) {
			defer wg.Done()
			infos, err := s.prober.Probe(ctx, target)
			results <- ProbeResult{
				Group:  s.prober.group,
				Time:   time.Now(),
				Target: s.prober.scheme + "://" + target,
				Infos:  infos,
				Err:    err,
			}
		}(target)
	}
	wg.Wait()

	s.Log.Debug("Scrape finished",
		zap.Duration("scrape_duration", time.Since(scrapeStart)))
}
