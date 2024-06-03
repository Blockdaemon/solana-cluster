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

// Package fetch provides the `fetch` command.
package fetch

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
	"go.blockdaemon.com/solana/cluster-manager/internal/fetch"
	"go.blockdaemon.com/solana/cluster-manager/internal/ledger"
	"go.blockdaemon.com/solana/cluster-manager/internal/logger"
	"go.blockdaemon.com/solana/cluster-manager/types"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gopkg.in/resty.v1"
)

var Cmd = cobra.Command{
	Use:   "fetch",
	Short: "Snapshot downloader",
	Long:  "Fetches a snapshot from another node using the tracker API.",
	Run: func(_ *cobra.Command, _ []string) {
		run()
	},
}

var (
	ledgerDir       string
	trackerURL      string
	minSnapAge      uint64
	maxSnapAge      uint64
	baseSlot        uint64
	fullSnap        bool
	incrementalSnap bool
	requestTimeout  time.Duration
	downloadTimeout time.Duration
)

func init() {
	flags := Cmd.Flags()
	flags.StringVar(&ledgerDir, "ledger", "", "Path to ledger dir")
	flags.StringVar(&trackerURL, "tracker", "", "Download as instructed by given tracker URL")
	flags.Uint64Var(&minSnapAge, "min-slots", 500, "Download only snapshots <n> slots newer than local")
	flags.Uint64Var(&maxSnapAge, "max-slots", 10000, "Refuse to download <n> slots older than the newest")
	flags.DurationVar(&requestTimeout, "request-timeout", 3*time.Second, "Max time to wait for headers (excluding download)")
	flags.DurationVar(&downloadTimeout, "download-timeout", 10*time.Minute, "Max time to try downloading in total")
	flags.Uint64Var(&baseSlot, "slot", 0, "Download snapshot for given slot (if available)")
	flags.BoolVar(&fullSnap, "full", true, "Download full snapshot (if available)")
	flags.BoolVar(&incrementalSnap, "incremental", true, "Download incremental snapshot (if available)")
}

func run() {
	log := logger.GetLogger()

	if !fullSnap && !incrementalSnap {
		log.Fatal("Must specify at least one of --full or --incremental")
	}

	// Regardless which API we talk to, we want to cap time from request to response header.
	// This defends against black holes and really slow servers.
	// Download time (reading response body) is not affected.
	http.DefaultTransport.(*http.Transport).ResponseHeaderTimeout = requestTimeout

	// Run until interrupted or time out occurs.
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	ctx, cancel2 := context.WithTimeout(ctx, downloadTimeout)
	defer cancel2()

	// Check what snapshots we have locally.
	localSnaps, err := ledger.ListSnapshots(os.DirFS(ledgerDir))
	if err != nil {
		log.Fatal("Failed to check existing snapshots", zap.Error(err))
	}

	// Get a specific snapshot or "best snapshot" from tracker client
	trackerClient := fetch.NewTrackerClientWithResty(
		resty.New().
			SetHostURL(trackerURL).
			SetTimeout(requestTimeout),
	)

	// If we are not downloading a full snap and we don't have a specific snapshot defined
	// we want to guess the snapshot slot to use as base, we will use the base slot
	if baseSlot == 0 && !fullSnap {
		baseSlot = localSnaps[0].BaseSlot
	}

	var remoteSnaps []types.SnapshotSource

	if baseSlot != 0 {
		log.Info("Fetching snapshots at slot", zap.Uint64("base_slot", baseSlot))

		// Ask tracker for snapshots at a specific location
		remoteSnaps, err = trackerClient.GetSnapshotAtSlot(ctx, baseSlot)
		if err != nil {
			log.Fatal("Failed to fetch snapshot info", zap.Error(err))
		}

		// @TODO check if this snapshot already exists
		buf, _ := json.MarshalIndent(remoteSnaps, "", "\t")
		log.Info("Downloading a snapshot", zap.ByteString("snap", buf))

	} else {
		log.Info("Finding best snapshot")

		// Ask tracker for best snapshots.
		remoteSnaps, err = trackerClient.GetBestSnapshots(ctx, -1)
		if err != nil {
			log.Fatal("Failed to request snapshot info", zap.Error(err))
		}

		// Decide what we want to do.
		_, advice := fetch.ShouldFetchSnapshot(localSnaps, remoteSnaps, minSnapAge, maxSnapAge)
		switch advice {
		case fetch.AdviceNothingFound:
			log.Error("No snapshots available remotely")
			return
		case fetch.AdviceUpToDate:
			log.Info("Existing snapshot is recent enough, no download needed",
				zap.Uint64("existing_slot", localSnaps[0].Slot))
			return
		case fetch.AdviceFetch:
		}
	}

	// Print snapshot to user.
	log.Info("Number of remote snaps found: ", zap.Int("num", len(remoteSnaps)))
	if len(remoteSnaps) == 0 {
		log.Fatal("Could not find any matching snapshots. Bailing.")
	}

	snap := &remoteSnaps[0]
	buf, _ := json.MarshalIndent(snap, "", "\t")
	log.Info("Downloading a snapshot", zap.ByteString("snap", buf), zap.String("target", snap.Target))

	// Setup progress bars for download.
	bars := mpb.New()
	sidecarClient := fetch.NewSidecarClientWithOpts(snap.Target, fetch.SidecarClientOpts{
		ProxyReaderFunc: func(name string, size int64, rd io.Reader) io.ReadCloser {
			bar := bars.New(
				size,
				mpb.BarStyle(),
				mpb.PrependDecorators(decor.Name(name)),
				mpb.AppendDecorators(
					decor.AverageSpeed(decor.UnitKB, "% .1f"),
					decor.Percentage(),
				),
			)
			return bar.ProxyReader(rd)
		},
	})

	// First pass, if we're fetching fullSnap we want to fetch the fullSnaps but **not** the incremental snap
	// Then after completion of the full snap download, we refetch the incremental one so we get the latest one
	if fullSnap {
		downloadSnapshot(ctx, sidecarClient, snap, true, false)

		// If we were downloading a full snapshot, check if there's a newer incremental snapshot we can fetch
		// Find latest incremental snapshot
		log.Info("Finding incremental snapshot for full slot", zap.Uint64("base_slot", snap.BaseSlot))
		remoteSnaps, err = trackerClient.GetSnapshotAtSlot(ctx, snap.BaseSlot)
		if err != nil {
			log.Fatal("Failed to request snapshot info", zap.Error(err))
		}

		spew.Dump(remoteSnaps)
		if len(remoteSnaps) == 0 {
			log.Fatal("No incremental snapshot found")
		}

		snap = &remoteSnaps[0]
		buf, _ = json.MarshalIndent(snap, "", "\t")
		log.Info("Downloading incremental snapshot", zap.ByteString("snap", buf), zap.String("target", snap.Target))
	}

	// This will fetch the latest incremental snapshot (if fullSnap was specified it would already have been fetched and refreshed)
	downloadSnapshot(ctx, sidecarClient, snap, false, true)
}

func downloadSnapshot(ctx context.Context, sidecarClient *fetch.SidecarClient, snap *types.SnapshotSource, full bool, incremental bool) {
	log := logger.GetLogger()

	// Download the snapshot files
	beforeDownload := time.Now()
	group, ctx := errgroup.WithContext(ctx)
	for _, file := range snap.Files {
		if file.BaseSlot != 0 && !incremental {
			continue
		}
		if file.BaseSlot == 0 && !full {
			continue
		}

		file_ := file
		group.Go(func() error {
			err := sidecarClient.DownloadSnapshotFile(ctx, ".", file_.FileName)
			if err != nil {
				log.Error("Full snapshot download failed",
					zap.String("snapshot", file_.FileName),
					zap.String("error", err.Error()))
			}
			return err
		})
	}
	downloadErr := group.Wait()
	downloadDuration := time.Since(beforeDownload)

	if downloadErr == nil {
		log.Info("Snapshot download completed", zap.Duration("download_time", downloadDuration))
	} else {
		log.Fatal("Aborting download", zap.Duration("download_time", downloadDuration))
	}

}
