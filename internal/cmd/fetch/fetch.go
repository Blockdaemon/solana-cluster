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

    "github.com/spf13/cobra"
    "github.com/vbauerster/mpb/v7"
    "github.com/vbauerster/mpb/v7/decor"
    "go.blockdaemon.com/solana/cluster-manager/internal/fetch"
    "go.blockdaemon.com/solana/cluster-manager/internal/ledger"
    "go.blockdaemon.com/solana/cluster-manager/internal/logger"
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
}

func run() {
    log := logger.GetLogger()

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

    // Ask tracker for best snapshots.
    trackerClient := fetch.NewTrackerClientWithResty(
        resty.New().
            SetHostURL(trackerURL).
            SetTimeout(requestTimeout),
    )
    remoteSnaps, err := trackerClient.GetBestSnapshots(ctx, -1)
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

    // Print snapshot to user.
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

    // Download.
    beforeDownload := time.Now()
    group, ctx := errgroup.WithContext(ctx)
    for _, file := range snap.Files {
        file_ := file
        group.Go(func() error {
            err := sidecarClient.DownloadSnapshotFile(ctx, ".", file_.FileName)
            if err != nil {
                log.Error("Download failed",
                    zap.String("snapshot", file_.FileName),
                    zap.String("error", err.Error()))
            }
            return err
        })
    }
    downloadErr := group.Wait()
    downloadDuration := time.Since(beforeDownload)

    if downloadErr == nil {
        log.Info("Download completed", zap.Duration("download_time", downloadDuration))
    } else {
        log.Fatal("Aborting download", zap.Duration("download_time", downloadDuration))
    }
}
