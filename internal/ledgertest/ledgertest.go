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

// Package ledgertest mocks the ledger dir for testing.
package ledgertest

import (
	"io/fs"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

type FS struct {
	DummyTime time.Time // arbitrary but consistent timestamp
	Root      afero.Fs
}

// ledgerPath is the relative path of the fake ledger dir to file system root.
const ledgerPath = "data/ledger"

// NewFS creates an in-memory file system resembling the ledger dir.
func NewFS(t *testing.T) *FS {
	t.Helper()
	root := afero.NewMemMapFs()
	require.NoError(t, root.MkdirAll(ledgerPath, 0755))
	return &FS{
		DummyTime: time.Date(2022, 4, 27, 15, 33, 20, 0, time.UTC),
		Root:      root,
	}
}

// AddFakeFile adds a file sized one byte to the fake ledger dir.
func (f *FS) AddFakeFile(t *testing.T, name string) {
	t.Helper()
	filePath := filepath.Join(ledgerPath, name)
	file, err := f.Root.Create(filePath)
	require.NoError(t, err)
	require.NoError(t, file.Truncate(1))
	require.NoError(t, file.Close())
	require.NoError(t, f.Root.Chtimes(filePath, f.DummyTime, f.DummyTime))
}

// GetLedgerDir returns the ledger dir as a standard library fs.FS.
func (f *FS) GetLedgerDir(t *testing.T) fs.FS {
	t.Helper()
	ledgerDir, err := fs.Sub(afero.NewIOFS(f.Root), ledgerPath)
	require.NoError(t, err)
	return ledgerDir
}
