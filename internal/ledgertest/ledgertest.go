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
