package scraper

import "go.blockdaemon.com/solana/cluster-manager/types"

type Collector struct {
	res chan ProbeResult
}

func NewCollector() *Collector {
	panic("not implemented")
}

func (c *Collector) Probes() chan<- ProbeResult {
	return c.res
}

type ProbeResult struct {
	Infos []*types.SnapshotInfo
	Err   error
}
