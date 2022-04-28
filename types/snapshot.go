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

package types

import (
	"bytes"
	"time"

	"github.com/gagliardetto/solana-go"
)

// SnapshotSource describes a snapshot, and where to get it from.
type SnapshotSource struct {
	SnapshotInfo
	Target    string
	UpdatedAt time.Time
}

// SnapshotInfo describes a snapshot.
type SnapshotInfo struct {
	Slot      uint64          `json:"slot"`
	Hash      solana.Hash     `json:"hash"`
	Files     []*SnapshotFile `json:"files"`
	TotalSize uint64          `json:"size"`
}

// SnapshotFile is a file that makes up a snapshot (either full or incremental).
type SnapshotFile struct {
	FileName string      `json:"file_name"`
	Slot     uint64      `json:"slot"`
	BaseSlot uint64      `json:"base_slot,omitempty"`
	Hash     solana.Hash `json:"hash"`
	Ext      string      `json:"ext"`

	ModTime *time.Time `json:"mod_time,omitempty"`
	Size    uint64     `json:"size,omitempty"`
}

// Compare implements lexicographic ordering by (slot, base_slot, hash).
func (s *SnapshotFile) Compare(o *SnapshotFile) int {
	if s.Slot < o.Slot {
		return -1
	} else if s.Slot > o.Slot {
		return +1
	} else if s.BaseSlot < o.BaseSlot {
		return -1
	} else if s.BaseSlot > o.BaseSlot {
		return +1
	} else {
		return bytes.Compare(s.Hash[:], o.Hash[:])
	}
}
