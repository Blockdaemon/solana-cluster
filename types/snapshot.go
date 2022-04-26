package types

import (
	"bytes"
	"time"

	"github.com/gagliardetto/solana-go"
)

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
