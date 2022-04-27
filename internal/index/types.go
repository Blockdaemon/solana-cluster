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

package index

import (
	"time"

	"go.blockdaemon.com/solana/cluster-manager/types"
)

type SnapshotEntry struct {
	SnapshotKey
	Info      *types.SnapshotInfo
	UpdatedAt time.Time
}

type SnapshotKey struct {
	Target      string
	InverseSlot uint64 // newest-to-oldest sort
}

func NewSnapshotKey(target string, slot uint64) SnapshotKey {
	return SnapshotKey{
		Target:      target,
		InverseSlot: ^slot,
	}
}

func (k SnapshotKey) Slot() uint64 {
	return ^k.InverseSlot
}
