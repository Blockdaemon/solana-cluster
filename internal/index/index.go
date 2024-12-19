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

// Package index is an in-memory index for snapshots.
package index

import (
	"time"

	"github.com/hashicorp/go-memdb"
)

type DB struct {
	DB *memdb.MemDB
}

// NewDB creates a new, empty in-memory database.
func NewDB() *DB {
	db, err := memdb.NewMemDB(&schema)
	if err != nil {
		panic("failed to create memDB: " + err.Error()) // unreachable
	}
	return &DB{
		DB: db,
	}
}

// UpsertSnapshots inserts the given snapshot entries.
// All entries must come from the same given target.
//
// Snapshots that have the same (target, slot) combination get replaced.
// Returns the number of snapshots that have been replaced (excluding new inserts).
func (d *DB) UpsertSnapshots(entries ...*SnapshotEntry) {
	txn := d.DB.Txn(true)
	defer txn.Abort()
	for _, entry := range entries {
		insertSnapshotEntry(txn, entry)
	}
	txn.Commit()
}

// GetSnapshotsByTarget returns all snapshots served by a host
// ordered by newest to oldest.
func (d *DB) GetSnapshotsByTarget(group string, target string) (entries []*SnapshotEntry) {
	res, err := d.DB.Txn(false).Get(tableSnapshotEntry, "id_prefix", group, target)
	if err != nil {
		panic("getting snapshots by target failed: " + err.Error())
	}
	for {
		entry := res.Next()
		if entry == nil {
			break
		}
		entries = append(entries, entry.(*SnapshotEntry))
	}
	return
}

// GetAllSnapshots returns a list of all snapshots.
func (d *DB) GetAllSnapshots() (entries []*SnapshotEntry) {
	iter, err := d.DB.Txn(false).LowerBound(tableSnapshotEntry, "id", "", uint64(0))
	if err != nil {
		panic("getting best snapshots failed: " + err.Error())
	}
	for {
		el := iter.Next()
		if el == nil {
			break
		}
		entries = append(entries, el.(*SnapshotEntry))
	}
	return
}

// GetBestSnapshots returns newest-to-oldest snapshots.
// The `max` argument controls the max number of snapshots to return.
// If max is negative, it returns all snapshots.
func (d *DB) GetBestSnapshotsByGroup(max int, group string) (entries []*SnapshotEntry) {
	var res memdb.ResultIterator
	var err error
	if group != "" {
		res, err = d.DB.Txn(false).Get(tableSnapshotEntry, "slot_by_group", group)
	} else {
		res, err = d.DB.Txn(false).Get(tableSnapshotEntry, "slot")
	}

	if err != nil {
		panic("getting best snapshots failed: " + err.Error())
	}
	for max < 0 || len(entries) <= max {
		entry := res.Next()
		if entry == nil {
			break
		}
		entries = append(entries, entry.(*SnapshotEntry))
	}
	return
}

// Fetches the snapshots that are at a given slot.
func (d *DB) GetSnapshotsAtSlotByGroup(group string, slot uint64) (entries []*SnapshotEntry) {
	var res memdb.ResultIterator
	var err error

	if group != "" {
		res, err = d.DB.Txn(false).Get(tableSnapshotEntry, "base_slot_by_group", slot, group)
	} else {
		res, err = d.DB.Txn(false).Get(tableSnapshotEntry, "base_slot", slot)
	}

	if err != nil {
		panic("getting best snapshots failed: " + err.Error())
	}

	for entry := res.Next(); entry != nil; entry = res.Next() {
		entries = append(entries, entry.(*SnapshotEntry))
	}
	return
}

// Fetches the best snapshots
func (d *DB) GetBestSnapshots(max int) (entries []*SnapshotEntry) {
	return d.GetBestSnapshotsByGroup(max, "")
}

// Fetches the snapshots that are at a given slot.
func (d *DB) GetSnapshotsAtSlot(slot uint64) (entries []*SnapshotEntry) {
	return d.GetSnapshotsAtSlotByGroup("", slot)
}

// DeleteOldSnapshots delete snapshot entry older than the given timestamp.
func (d *DB) DeleteOldSnapshots(minTime time.Time) (n int) {
	txn := d.DB.Txn(true)
	defer txn.Abort()
	res, err := txn.Get(tableSnapshotEntry, "id_prefix")
	if err != nil {
		panic("failed to range over all snapshots: " + err.Error())
	}
	for {
		entry := res.Next()
		if entry == nil {
			break
		}
		if entry.(*SnapshotEntry).UpdatedAt.Before(minTime) {
			if err := txn.Delete(tableSnapshotEntry, entry); err != nil {
				panic("failed to delete expired snapshot: " + err.Error())
			}
			n++
		}
	}
	txn.Commit()
	return
}

// DeleteSnapshotsByTarget deletes all snapshots owned by a given target.
// Returns the number of deletions made.
func (d *DB) DeleteSnapshotsByTarget(group string, target string) int {
	txn := d.DB.Txn(true)
	defer txn.Abort()
	n, err := txn.DeleteAll(tableSnapshotEntry, "id_prefix", group, target)
	if err != nil {
		panic("failed to delete snapshots by target: " + err.Error())
	}
	txn.Commit()
	return n
}

func insertSnapshotEntry(txn *memdb.Txn, snap *SnapshotEntry) {
	if err := txn.Insert(tableSnapshotEntry, snap); err != nil {
		panic("failed to insert snapshot entry: " + err.Error())
	}
}
