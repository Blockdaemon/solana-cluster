// Package ledger interacts with the ledger dir.
package ledger

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gagliardetto/solana-go"
	"github.com/samber/lo"
	"go.blockdaemon.com/solana/cluster-manager/types"
)

// ListSnapshots shows all available snapshots of a ledger dir in the specified FS.
func ListSnapshots(ledgerDir fs.FS) ([]*types.SnapshotInfo, error) {
	// List and stat snapshot files.
	dirEntries, err := fs.ReadDir(ledgerDir, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to list ledger dir: %w", err)
	}
	var files []*types.SnapshotFile
	for _, dirEntry := range dirEntries {
		if !dirEntry.Type().IsRegular() {
			continue
		}
		info := ParseSnapshotFileName(dirEntry.Name())
		if info == nil {
			continue
		}
		if err := SnapshotStat(ledgerDir, info); err != nil {
			continue
		}
		files = append(files, info)
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Compare(files[j]) < 0
	})

	// Reconstruct snapshot chains for all available snapshots.
	infos := make([]*types.SnapshotInfo, 0, len(files))
	for _, file := range files {
		if info := buildSnapshotInfo(files, file); info != nil {
			infos = append(infos, info)
		}
	}
	return infos, nil
}

// buildSnapshotInfo builds a snapshot info object against the target snapshot file.
// The files array must be sorted.
func buildSnapshotInfo(files []*types.SnapshotFile, target *types.SnapshotFile) *types.SnapshotInfo {
	// Start at target snapshot and reconstruct chain of snapshots.
	chain := []*types.SnapshotFile{target}
	totalSize := target.Size
	for {
		base := chain[len(chain)-1].BaseSlot
		if base == 0 {
			break // complete chain
		}
		// Find snapshot matching base slot number.
		index := sort.Search(len(files), func(i int) bool {
			return files[i].Slot >= base
		})
		if index >= len(files) || files[index].Slot != base {
			return nil // incomplete chain
		}
		// Extend snapshot chain.
		chain = append(chain, files[index])
		totalSize += files[index].Size
	}
	return &types.SnapshotInfo{
		Slot:      target.Slot,
		Hash:      target.Hash,
		Files:     lo.Reverse(chain), // oldest-to-newest
		TotalSize: totalSize,
	}
}

// ParseSnapshotFileName parses a snapshot's name.
func ParseSnapshotFileName(name string) *types.SnapshotFile {
	// Split file name into base and stem.
	stem := name
	var ext string
	for i := 0; i < 2; i++ {
		extPart := filepath.Ext(stem)
		stem = strings.TrimSuffix(stem, extPart)
		ext = extPart + ext
	}
	if strings.ContainsAny(stem, " \t\n") {
		return nil
	}
	// Parse file name fields.
	if strings.HasPrefix(stem, "snapshot-") {
		var slot uint64
		var hashStr string
		n, err := fmt.Sscanf(stem, "snapshot-%d-%s", &slot, &hashStr)
		if n != 2 || err != nil {
			return nil
		}
		hash, err := solana.HashFromBase58(hashStr)
		if err != nil {
			return nil
		}
		return &types.SnapshotFile{
			FileName: name,
			Slot:     slot,
			Hash:     hash,
			Ext:      ext,
		}
	}
	if strings.HasPrefix(stem, "incremental-snapshot-") {
		var baseSlot, incrementalSlot uint64
		var hashStr string
		n, err := fmt.Sscanf(stem, "incremental-snapshot-%d-%d-%s", &baseSlot, &incrementalSlot, &hashStr)
		if n != 3 || err != nil {
			return nil
		}
		hash, err := solana.HashFromBase58(hashStr)
		if err != nil {
			return nil
		}
		return &types.SnapshotFile{
			FileName: name,
			Slot:     incrementalSlot,
			BaseSlot: baseSlot,
			Hash:     hash,
			Ext:      ext,
		}
	}
	return nil
}

// SnapshotStat fills stat info into the snapshot file.
func SnapshotStat(fs_ fs.FS, snap *types.SnapshotFile) error {
	stat, err := fs.Stat(fs_, snap.FileName)
	if err != nil {
		return err
	}
	snap.Size = uint64(stat.Size())
	if modTime := stat.ModTime(); !modTime.IsZero() {
		snap.ModTime = &modTime
	}
	return nil
}
