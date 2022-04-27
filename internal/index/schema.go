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

import "github.com/hashicorp/go-memdb"

const tableSnapshotEntry = "snapshot_entry"

var schema = memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		tableSnapshotEntry: {
			Name: tableSnapshotEntry,
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:   "id",
					Unique: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "Target"},
							&memdb.UintFieldIndex{Field: "InverseSlot"},
						},
						AllowMissing: false,
					},
				},
				"slot": {
					Name:    "slot",
					Unique:  false,
					Indexer: &memdb.UintFieldIndex{Field: "InverseSlot"},
				},
			},
		},
	},
}
