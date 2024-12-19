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
	"testing"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/assert"
	"go.blockdaemon.com/solana/cluster-manager/types"
)

var dummyTime1 = time.Date(2022, 4, 27, 15, 33, 20, 0, time.UTC)

var snapshotEntry1 = &SnapshotEntry{
	SnapshotKey: NewSnapshotKey("mainnet", "host1", 100, 100),
	UpdatedAt:   dummyTime1,
	Info: &types.SnapshotInfo{
		Slot:      100,
		Hash:      solana.Hash{0x03},
		Files:     []*types.SnapshotFile{},
		TotalSize: 0,
	},
}

var snapshotEntry2 = &SnapshotEntry{
	SnapshotKey: NewSnapshotKey("mainnet", "host1", 99, 99),
	UpdatedAt:   dummyTime1.Add(-20 * time.Second),
	Info: &types.SnapshotInfo{
		Slot:      99,
		Hash:      solana.Hash{0x04},
		Files:     []*types.SnapshotFile{},
		TotalSize: 0,
	},
}

var snapshotEntry3 = &SnapshotEntry{
	SnapshotKey: NewSnapshotKey("mainnet", "host2", 100, 100),
	UpdatedAt:   dummyTime1,
	Info: &types.SnapshotInfo{
		Slot:      100,
		Hash:      solana.Hash{0x03},
		Files:     []*types.SnapshotFile{},
		TotalSize: 0,
	},
}

var snapshotEntry4 = &SnapshotEntry{
	SnapshotKey: NewSnapshotKey("devnet", "host1", 100, 100),
	UpdatedAt:   dummyTime1,
	Info: &types.SnapshotInfo{
		Slot:      151,
		Hash:      solana.Hash{0x03},
		Files:     []*types.SnapshotFile{},
		TotalSize: 0,
	},
}

func TestDB(t *testing.T) {
	db := NewDB()

	assert.Equal(t, 0, db.DeleteSnapshotsByTarget("mainnet", "host1"))
	assert.Equal(t, 0, db.DeleteSnapshotsByTarget("mainnet", "host2"))

	assert.Len(t, db.GetSnapshotsByTarget("mainnet", "host1"), 0)
	assert.Len(t, db.GetSnapshotsByTarget("mainnet", "host2"), 0)
	assert.Len(t, db.GetBestSnapshots(-1), 0)

	db.UpsertSnapshots(snapshotEntry1)
	assert.Len(t, db.GetSnapshotsByTarget("mainnet", "host1"), 1)
	assert.Len(t, db.GetSnapshotsByTarget("mainnet", "host2"), 0)
	assert.Len(t, db.GetBestSnapshots(-1), 1)

	db.UpsertSnapshots(snapshotEntry1, snapshotEntry2)
	assert.Len(t, db.GetSnapshotsByTarget("mainnet", "host1"), 2)
	assert.Len(t, db.GetSnapshotsByTarget("mainnet", "host2"), 0)
	assert.Equal(t,
		[]*SnapshotEntry{
			snapshotEntry1,
			snapshotEntry2,
		},
		db.GetBestSnapshots(-1))

	db.UpsertSnapshots(snapshotEntry2, snapshotEntry3)
	assert.Len(t, db.GetSnapshotsByTarget("mainnet", "host1"), 2)
	assert.Len(t, db.GetSnapshotsByTarget("mainnet", "host2"), 1)
	assert.Equal(t,
		[]*SnapshotEntry{
			snapshotEntry1,
			snapshotEntry3,
			snapshotEntry2,
		},
		db.GetBestSnapshots(-1))

	// Add a devnet target and ensure the result is the same
	db.UpsertSnapshots(snapshotEntry4)
	assert.Len(t, db.GetSnapshotsByTarget("mainnet", "host1"), 2)
	assert.Len(t, db.GetSnapshotsByTarget("mainnet", "host2"), 1)
	assert.Equal(t,
		[]*SnapshotEntry{
			snapshotEntry4,
			snapshotEntry1,
			snapshotEntry3,
			snapshotEntry2,
		},
		db.GetBestSnapshots(-1))

	assert.Equal(t, 2, db.DeleteSnapshotsByTarget("mainnet", "host1"))
	assert.Len(t, db.GetSnapshotsByTarget("mainnet", "host1"), 0)
	assert.Len(t, db.GetSnapshotsByTarget("mainnet", "host2"), 1)
	assert.Equal(t,
		[]*SnapshotEntry{
			snapshotEntry4,
			snapshotEntry3,
		},
		db.GetBestSnapshots(-1))

	db.UpsertSnapshots(snapshotEntry1, snapshotEntry2)

	assert.Equal(t, 1, db.DeleteOldSnapshots(snapshotEntry2.UpdatedAt.Add(time.Second)))
	assert.Len(t, db.GetSnapshotsByTarget("mainnet", "host1"), 1)
	assert.Len(t, db.GetSnapshotsByTarget("mainnet", "host2"), 1)
	assert.Equal(t,
		[]*SnapshotEntry{
			snapshotEntry4,
			snapshotEntry1,
			snapshotEntry3,
		},
		db.GetBestSnapshots(-1))

    assert.Equal(t,
		[]*SnapshotEntry{
			snapshotEntry4,
		},
		db.GetBestSnapshotsByGroup(-1, "devnet"))

	assert.Equal(t,
		[]*SnapshotEntry{
			snapshotEntry1,
			snapshotEntry3
		}, 	)
}
