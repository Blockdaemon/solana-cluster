package scraper

import (
	"fmt"
	"sync"

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
	disc := group.Discoverer()
	if disc == nil {
		return fmt.Errorf("no target discovery")
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
