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

// DeleteSnapshotsByTarget deletes all snapshots owned by a given target.
// Returns the number
func (d *DB) DeleteSnapshotsByTarget(target string) int {
	txn := d.DB.Txn(true)
	defer txn.Abort()
	n, err := txn.DeleteAll(tableSnapshotEntry, "id_prefix", target)
	if err != nil {
		panic("failed to delete snapshots by target: " + err.Error())
	}
	txn.Commit()
	return n
}

// UpsertSnapshots inserts the given snapshot entries.
//
// Snapshots that have the same (target, slot) combination get replaced.
// Returns the number of snapshots that have been replaced (excluding new inserts).
func (d *DB) UpsertSnapshots(entries ...*SnapshotEntry) (numReplaced int) {
	txn := d.DB.Txn(true)
	defer txn.Abort()
	for _, entry := range entries {
		if deleteSnapshotEntry(txn, &entry.SnapshotKey) {
			numReplaced++
		}
		insertSnapshotEntry(txn, entry)
	}
	txn.Commit()
	return
}

// GetSnapshotsByTarget returns all snapshots served by a host
// ordered by newest to oldest.
func (d *DB) GetSnapshotsByTarget(target string) (entries []*SnapshotEntry) {
	res, err := d.DB.Txn(false).Get(tableSnapshotEntry, "id_prefix", target)
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

// GetBestSnapshots returns highest-to-oldest snapshots.
// The `max` argument controls the max number of snapshots to return.
// If max is negative, it returns all snapshots.
func (d *DB) GetBestSnapshots(max int) (entries []*SnapshotEntry) {
	res, err := d.DB.Txn(false).Get(tableSnapshotEntry, "slot")
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

func deleteSnapshotEntry(txn *memdb.Txn, key *SnapshotKey) bool {
	existing, err := txn.First(tableSnapshotEntry, "id", key.Target, key.InverseSlot)
	if err != nil {
		panic("lookup failed: " + err.Error())
	}
	if existing == nil {
		return false
	}
	if err := txn.Delete(tableSnapshotEntry, existing); err != nil {
		panic("failed to delete existing snapshot entry: " + err.Error())
	}
	return true
}

func insertSnapshotEntry(txn *memdb.Txn, snap *SnapshotEntry) {
	if err := txn.Insert(tableSnapshotEntry, snap); err != nil {
		panic("failed to insert snapshot entry: " + err.Error())
	}
}
