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

package fetch

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.blockdaemon.com/solana/cluster-manager/types"
)

func TestShouldFetchSnapshot(t *testing.T) {
	cases := []struct {
		name string

		local  []uint64
		remote []uint64
		minAge uint64
		maxAge uint64

		minSlot uint64
		advice  Advice
	}{
		{
			name:    "NothingRemote",
			local:   []uint64{},
			remote:  []uint64{},
			minAge:  500,
			maxAge:  10000,
			minSlot: 0,
			advice:  AdviceNothingFound,
		},
		{
			name:    "NothingLocal",
			local:   []uint64{},
			remote:  []uint64{123456},
			minAge:  500,
			maxAge:  10000,
			minSlot: 113456,
			advice:  AdviceFetchFull,
		},
		{
			name:    "LowSlotNumber",
			local:   []uint64{},
			remote:  []uint64{100},
			minAge:  50,
			maxAge:  10000,
			minSlot: 0,
			advice:  AdviceFetchFull,
		},
		{
			name:    "Refresh",
			local:   []uint64{100000},
			remote:  []uint64{123456},
			minAge:  500,
			maxAge:  10000,
			minSlot: 113456,
			advice:  AdviceFetchFull,
		},
		{
			name:    "NotNewEnough",
			local:   []uint64{100000},
			remote:  []uint64{100002},
			minAge:  500,
			maxAge:  10000,
			minSlot: 0,
			advice:  AdviceUpToDate,
		},
		{
			name:    "UpToDate",
			local:   []uint64{223456},
			remote:  []uint64{123456},
			minAge:  500,
			maxAge:  10000,
			minSlot: 0,
			advice:  AdviceUpToDate,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			minSlot, advice := ShouldFetchSnapshot(
				fakeSnapshotInfo(tc.local),
				fakeSnapshotSources(tc.remote),
				tc.minAge,
				tc.maxAge,
			)
			assert.Equal(t, tc.minSlot, minSlot, "different minSlot")
			assert.Equal(t, tc.advice, advice, "different advice")
		})
	}
}

func fakeSnapshotInfo(slots []uint64) []*types.SnapshotInfo {
	infos := make([]*types.SnapshotInfo, len(slots))
	for i, slot := range slots {
		infos[i] = &types.SnapshotInfo{
			Slot:     slot,
			BaseSlot: slot,
		}
	}
	return infos
}

func fakeSnapshotSources(slots []uint64) []types.SnapshotSource {
	infos := make([]types.SnapshotSource, len(slots))
	for i, slot := range slots {
		infos[i] = types.SnapshotSource{
			SnapshotInfo: types.SnapshotInfo{
				Slot:     slot,
				BaseSlot: slot,
			},
		}
	}
	return infos
}
