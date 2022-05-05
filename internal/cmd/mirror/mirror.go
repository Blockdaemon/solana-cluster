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

// Package mirror provides the `mirror` command.
package mirror

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"go.blockdaemon.com/solana/cluster-manager/internal/fetch"
	"go.blockdaemon.com/solana/cluster-manager/internal/logger"
	"go.blockdaemon.com/solana/cluster-manager/internal/mirror"
)

var Cmd = cobra.Command{
	Use:   "mirror",
	Short: "Maintenance daemon for snapshot S3 buckets",
	Long:  "Uploads snapshots to an S3 bucket",
	Run: func(cmd *cobra.Command, _ []string) {
		run(cmd)
	},
}

var (
	refreshInterval time.Duration
	trackerURL      string
	bucket          string
	objectPrefix    string
)

func init() {
	flags := Cmd.Flags()
	flags.DurationVar(&refreshInterval, "refresh", 30*time.Second, "Refresh interval to discover new snapshots")
	flags.StringVar(&trackerURL, "tracker", "", "URL to tracker API")
	flags.StringVar(&bucket, "bucket", "", "Bucket name")
	flags.StringVar(&objectPrefix, "prefix", "", "Object prefix")
}

func run(cmd *cobra.Command) {
	log := logger.GetConsoleLogger()
	_ = log

	if trackerURL == "" || bucket == "" {
		_ = cmd.Usage()
		return
	}

	// TODO

	trackerClient := fetch.NewTrackerClient(trackerURL)

	uploader := mirror.Uploader{
		S3Client:     nil,
		Bucket:       bucket,
		ObjectPrefix: objectPrefix,
	}

	worker := mirror.Worker{
		Tracker:   trackerClient,
		Uploader:  &uploader,
		Log:       log.Named("uploader"),
		Refresh:   refreshInterval,
		SyncCount: 10, // TODO
	}

	for range time.Tick(refreshInterval) {
		worker.Run(context.TODO())
	}
}
