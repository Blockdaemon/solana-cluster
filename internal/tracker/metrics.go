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

package tracker

import "github.com/prometheus/client_golang/prometheus"

// snapshotAge tracks how old the best available snapshot is, measured in slots
// relative to the current slot fetched from RPC. Updated on each health check.
var snapshotAge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "solana_cluster_tracker_snapshot_age_slots",
		Help: "Age of the best available snapshot in slots relative to current RPC slot.",
	},
	[]string{"group"},
)

func init() {
	prometheus.MustRegister(snapshotAge)
}
