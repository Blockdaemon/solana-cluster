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

import "github.com/prometheus/client_golang/prometheus"

var (
	// targetsDiscovered tracks the current number of targets found per group
	// via Consul service discovery on each scrape cycle.
	targetsDiscovered = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "solana_cluster_tracker_targets_total",
			Help: "Current number of targets discovered via service discovery.",
		},
		[]string{"group"},
	)

	// snapshotsFound counts the total number of snapshots found across all
	// targets per group over time.
	snapshotsFound = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "solana_cluster_tracker_snapshots_found_total",
			Help: "Total number of snapshots found across all targets.",
		},
		[]string{"group"},
	)
)

func init() {
	prometheus.MustRegister(targetsDiscovered, snapshotsFound)
}
