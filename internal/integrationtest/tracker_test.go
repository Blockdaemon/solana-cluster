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

package integrationtest

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.blockdaemon.com/solana/cluster-manager/internal/fetch"
	"go.blockdaemon.com/solana/cluster-manager/internal/index"
	"go.blockdaemon.com/solana/cluster-manager/internal/scraper"
	"go.blockdaemon.com/solana/cluster-manager/internal/tracker"
	"go.blockdaemon.com/solana/cluster-manager/types"
	"go.uber.org/zap/zaptest"
)

// TestTracker creates
// a bunch of sidecars with a fake ledger dir,
// a tracker running scraping infrastructure,
// a tracker client
func TestTracker(t *testing.T) {
	const sidecarCount = 4
	var targets []string
	for i := uint64(100); i < 100+sidecarCount; i++ {
		server, _ := newSidecar(t, i)
		defer server.Close()
		u, err := url.Parse(server.URL)
		require.NoError(t, err)
		targets = append(targets, u.Host)
	}
	group := &types.TargetGroup{
		Group:         "test",
		Scheme:        "http",
		APIPath:       "",
		StaticTargets: &types.StaticTargets{Targets: targets},
	}

	// Create scraper infra.
	db := index.NewDB()
	collector := scraper.NewCollector(db)
	collector.Log = zaptest.NewLogger(t).Named("scraper")
	collector.Start()
	defer collector.Close()
	prober, err := scraper.NewProber(group)
	require.NoError(t, err)
	scraper_ := scraper.NewScraper(prober, group.StaticTargets)
	defer scraper_.Close()

	// Create tracker server.
	server := newTracker(db)
	defer server.Close()

	// Scrape for a while.
	scraper_.Start(collector.Probes(), 50*time.Millisecond)
	for i := 0; ; i++ {
		time.Sleep(25 * time.Millisecond)
		if i >= 50 {
			t.Fatal("Scrape timeout")
		}
		found := len(db.GetBestSnapshots(-1))
		if found == sidecarCount {
			break
		}
		t.Logf("Found %d snapshots, waiting for %d", found, sidecarCount)
	}
	scraper_.Close() // We can stop scraping

	// Create tracker client.
	client := fetch.NewTrackerClientWithResty(resty.NewWithClient(server.Client()).SetHostURL(server.URL))
	snaps, err := client.GetBestSnapshots(context.TODO(), -1)
	require.NoError(t, err)
	// Remove timestamps and port numbers.
	for i := range snaps {
		snap := &snaps[i]
		assert.False(t, snap.UpdatedAt.IsZero())
		snap.UpdatedAt = time.Time{}
		for _, file := range snap.Files {
			assert.NotNil(t, file.ModTime)
			file.ModTime = nil
		}
		assert.NotEmpty(t, snap.Target)
		snap.Target = ""
	}
	assert.Equal(t,
		[]types.SnapshotSource{
			{
				SnapshotInfo: types.SnapshotInfo{
					Slot: 103,
					Hash: solana.MustHashFromBase58("7w4zb1jh47zY5FPMPyRzDSmYf1CPirVP9LmTr5xWEs6X"),
					Files: []*types.SnapshotFile{
						{
							FileName: "snapshot-103-7w4zb1jh47zY5FPMPyRzDSmYf1CPirVP9LmTr5xWEs6X.tar.bz2",
							Slot:     103,
							Hash:     solana.MustHashFromBase58("7w4zb1jh47zY5FPMPyRzDSmYf1CPirVP9LmTr5xWEs6X"),
							Size:     1,
							Ext:      ".tar.bz2",
						},
					},
					TotalSize: 1,
				},
			},
			{
				SnapshotInfo: types.SnapshotInfo{
					Slot: 102,
					Hash: solana.MustHashFromBase58("7sAawX1cAHVpfZGNtUAYKX2KPzdd1uPUZUTaLteWX4SB"),
					Files: []*types.SnapshotFile{
						{
							FileName: "snapshot-102-7sAawX1cAHVpfZGNtUAYKX2KPzdd1uPUZUTaLteWX4SB.tar.bz2",
							Slot:     102,
							Hash:     solana.MustHashFromBase58("7sAawX1cAHVpfZGNtUAYKX2KPzdd1uPUZUTaLteWX4SB"),
							Size:     1,
							Ext:      ".tar.bz2",
						},
					},
					TotalSize: 1,
				},
			},
			{
				SnapshotInfo: types.SnapshotInfo{
					Slot: 101,
					Hash: solana.MustHashFromBase58("7oGBJ2HXGT17Fs9QNxu6RbH68z4rJxHZyc9gqhLWoFmq"),
					Files: []*types.SnapshotFile{
						{
							FileName: "snapshot-101-7oGBJ2HXGT17Fs9QNxu6RbH68z4rJxHZyc9gqhLWoFmq.tar.bz2",
							Slot:     101,
							Hash:     solana.MustHashFromBase58("7oGBJ2HXGT17Fs9QNxu6RbH68z4rJxHZyc9gqhLWoFmq"),
							Size:     1,
							Ext:      ".tar.bz2",
						},
					},
					TotalSize: 1,
				},
			},
			{
				SnapshotInfo: types.SnapshotInfo{
					Slot: 100,
					Hash: solana.MustHashFromBase58("7jMmeXZSNcWPrB2RsTdeXfXrsyW5c1BfPjqoLW2X5T7V"),
					Files: []*types.SnapshotFile{
						{
							FileName: "snapshot-100-7jMmeXZSNcWPrB2RsTdeXfXrsyW5c1BfPjqoLW2X5T7V.tar.bz2",
							Slot:     100,
							Hash:     solana.MustHashFromBase58("7jMmeXZSNcWPrB2RsTdeXfXrsyW5c1BfPjqoLW2X5T7V"),
							Size:     1,
							Ext:      ".tar.bz2",
						},
					},
					TotalSize: 1,
				},
			},
		},
		snaps)
}

func newTracker(db *index.DB) *httptest.Server {
	handler := tracker.NewHandler(db)
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	handler.RegisterHandlers(engine.Group("/v1"))
	return httptest.NewServer(engine)
}
