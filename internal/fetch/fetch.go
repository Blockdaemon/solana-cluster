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

import "go.blockdaemon.com/solana/cluster-manager/types"

// ShouldFetchSnapshot returns whether a new snapshot should be fetched.
//
// If advice is AdviceFetch, `minSlot` indicates the lowest slot number at which fetch is useful.
func ShouldFetchSnapshot(
	local []*types.SnapshotInfo,
	remote []types.SnapshotSource,
	minAge uint64, // if diff between remote and local is smaller than minAge, use local
	maxAge uint64, // if diff between latest remote and any other remote is larger than maxAge, abort
) (minSlot uint64, advice Advice) {
	// Check if remote reports to snapshots.
	if len(remote) == 0 {
		advice = AdviceNothingFound
		return
	}

	// Compare local and remote slot numbers.
	remoteSlot := remote[0].Slot
	var localSlot uint64
	var localBaseSlot uint64
	if len(local) > 0 {
		localSlot = local[0].Slot
		localBaseSlot = local[0].BaseSlot
	}

	// Check if local is newer or remote is not new enough to be interesting.
	if int64(remoteSlot)-int64(localSlot) < int64(minAge) {
		advice = AdviceUpToDate
		return
	}

	// Remote is new enough.
	if maxAge < remoteSlot {
		minSlot = remoteSlot - maxAge
	}

	// Check if we have a full snapshot new enough
	if localBaseSlot < remote[0].BaseSlot {
		advice = AdviceFetchFull
		return
	} else if len(local) > 0 && localBaseSlot == remote[0].BaseSlot {
		for _, l := range local[0].Files {
			for _, r := range remote[0].Files {
				if l.BaseSlot == 0 && r.BaseSlot == 0 {
					if l.Hash == r.Hash {
						// If the full snapshot base slot is the same and the hash is the same
						// we can skip fetching the full snapshot.
						advice = AdviceFetchIncremental
						return
					}
				}
			}
		}
	} else {
		advice = AdviceRemoteIsOlder
		return
	}

	// Our advice is to fetch both incremental and full snapshots.
	advice = AdviceFetch
	return
}

// Advice indicates the recommended next action.
type Advice int

const (
	AdviceFetch            = Advice(iota) // download a snapshot
	AdviceFetchFull                       // download a full snapshot
	AdviceFetchIncremental                // download an incremental snapshot
	AdviceRemoteIsOlder                   // remote snapshot is older than local
	AdviceNothingFound                    // no snapshot available
	AdviceUpToDate                        // local snapshot is up-to-date or newer, don't download
)
