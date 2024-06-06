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

package ledger

import (
	"encoding/json"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.blockdaemon.com/solana/cluster-manager/internal/ledgertest"
	"go.blockdaemon.com/solana/cluster-manager/types"
)

func TestListSnapshots(t *testing.T) {
	// Construct a new ledger dir with a bunch of snapshots.
	root := ledgertest.NewFS(t)
	fakeFiles := []string{
		"snapshot-50-AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr.tar.bz2",
		"incremental-snapshot-50-100-AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr.tar.zst",
		"snapshot-100-AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr.tar.bz2",
		"incremental-snapshot-100-200-AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr.tar.zst",
		"incremental-snapshot-200-300-AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr.tar.zst",
		"incremental-snapshot-99999-1010101-AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr.tar.zst",
	}
	for _, name := range fakeFiles {
		root.AddFakeFile(t, name)
	}
	ledgerDir := root.GetLedgerDir(t)

	snapshots, err := ListSnapshots(ledgerDir)
	require.NoError(t, err)

	j, _ := json.MarshalIndent(snapshots, "", "\t")
	t.Log(string(j))

	assert.Equal(t,
		[]*types.SnapshotInfo{
			{
				Slot:      300,
				BaseSlot:  100,
				Hash:      solana.MustHashFromBase58("AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr"),
				TotalSize: 3,
				Files: []*types.SnapshotFile{
					{
						FileName: "incremental-snapshot-200-300-AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr.tar.zst",
						BaseSlot: 200,
						Slot:     300,
						Hash:     solana.MustHashFromBase58("AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr"),
						Ext:      ".tar.zst",
						Size:     1,
						ModTime:  &root.DummyTime,
					},
					{
						FileName: "incremental-snapshot-100-200-AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr.tar.zst",
						BaseSlot: 100,
						Slot:     200,
						Hash:     solana.MustHashFromBase58("AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr"),
						Ext:      ".tar.zst",
						Size:     1,
						ModTime:  &root.DummyTime,
					},
					{
						FileName: "snapshot-100-AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr.tar.bz2",
						Slot:     100,
						Hash:     solana.MustHashFromBase58("AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr"),
						Ext:      ".tar.bz2",
						Size:     1,
						ModTime:  &root.DummyTime,
					},
				},
			},
			{
				Slot:      200,
				BaseSlot:  100,
				Hash:      solana.MustHashFromBase58("AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr"),
				TotalSize: 2,
				Files: []*types.SnapshotFile{
					{
						FileName: "incremental-snapshot-100-200-AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr.tar.zst",
						BaseSlot: 100,
						Slot:     200,
						Hash:     solana.MustHashFromBase58("AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr"),
						Ext:      ".tar.zst",
						Size:     1,
						ModTime:  &root.DummyTime,
					},
					{
						FileName: "snapshot-100-AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr.tar.bz2",
						Slot:     100,
						Hash:     solana.MustHashFromBase58("AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr"),
						Ext:      ".tar.bz2",
						Size:     1,
						ModTime:  &root.DummyTime,
					},
				},
			},
			{
				Slot:      100,
				BaseSlot:  100,
				Hash:      solana.MustHashFromBase58("AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr"),
				TotalSize: 1,
				Files: []*types.SnapshotFile{
					{
						FileName: "snapshot-100-AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr.tar.bz2",
						Slot:     100,
						Hash:     solana.MustHashFromBase58("AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr"),
						Ext:      ".tar.bz2",
						Size:     1,
						ModTime:  &root.DummyTime,
					},
				},
			},
			{
				Slot:      100,
				BaseSlot:  50,
				Hash:      solana.MustHashFromBase58("AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr"),
				TotalSize: 2,
				Files: []*types.SnapshotFile{
					{
						FileName: "incremental-snapshot-50-100-AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr.tar.zst",
						BaseSlot: 50,
						Slot:     100,
						Hash:     solana.MustHashFromBase58("AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr"),
						Ext:      ".tar.zst",
						Size:     1,
						ModTime:  &root.DummyTime,
					},
					{
						FileName: "snapshot-50-AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr.tar.bz2",
						Slot:     50,
						Hash:     solana.MustHashFromBase58("AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr"),
						Ext:      ".tar.bz2",
						Size:     1,
						ModTime:  &root.DummyTime,
					},
				},
			},
			{
				Slot:      50,
				BaseSlot:  50,
				Hash:      solana.MustHashFromBase58("AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr"),
				TotalSize: 1,
				Files: []*types.SnapshotFile{
					{
						FileName: "snapshot-50-AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr.tar.bz2",
						Slot:     50,
						Hash:     solana.MustHashFromBase58("AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr"),
						Ext:      ".tar.bz2",
						Size:     1,
						ModTime:  &root.DummyTime,
					},
				},
			},
		},
		snapshots,
	)
}

func TestParseSnapshotFileName(t *testing.T) {
	cases := []struct {
		name string
		path string
		info *types.SnapshotFile
	}{
		{
			name: "Empty",
			path: "",
			info: nil,
		},
		{
			name: "MissingParts",
			path: "snapshot-121646378.tar.zst",
			info: nil,
		},
		{
			name: "InvalidSlotNumber",
			path: "snapshot-notaslotnumber-AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr.tar.bz2",
			info: nil,
		},
		{
			name: "InvalidHash",
			path: "snapshot-12345678-bad!hash.tar",
			info: nil,
		},
		{
			name: "IncrementalSnapshot",
			path: "incremental-snapshot-100-200-AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr.tar.zst",
			info: &types.SnapshotFile{
				FileName: "incremental-snapshot-100-200-AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr.tar.zst",
				BaseSlot: 100,
				Slot:     200,
				Hash:     solana.MustHashFromBase58("AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr"),
				Ext:      ".tar.zst",
			},
		},
		{
			name: "IncrementalSnapshotInvalidHash",
			path: "incremental-snapshot-100-12345678-bad!hash.tar",
			info: nil,
		},
		{
			name: "IncrementalSnapshotWeird",
			path: "incremental-snapshot-100.tar",
			info: nil,
		},
		{
			name: "IncrementalSnapshotWhitespace",
			path: "incremental-snapshot- e.tar",
			info: nil,
		},
		{
			name: "IncrementalSnapshotImpossible",
			path: "incremental-snapshot-300-200-AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr.tar.zst",
			info: nil,
		},
		{
			name: "NormalSnapshot",
			path: "snapshot-100-AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr.tar.bz2",
			info: &types.SnapshotFile{
				FileName: "snapshot-100-AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr.tar.bz2",
				Slot:     100,
				Hash:     solana.MustHashFromBase58("AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr"),
				Ext:      ".tar.bz2",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			info := ParseSnapshotFileName(tc.path)
			assert.Equal(t, tc.info, info)
		})
	}
}
