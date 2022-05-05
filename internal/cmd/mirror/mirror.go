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

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/cobra"
	"go.blockdaemon.com/solana/cluster-manager/internal/fetch"
	"go.blockdaemon.com/solana/cluster-manager/internal/logger"
	"go.blockdaemon.com/solana/cluster-manager/internal/mirror"
	"go.uber.org/zap"
)

var Cmd = cobra.Command{
	Use:   "mirror",
	Short: "Daemon to periodically upload snapshots to S3",
	Long:  "Uploads snapshots to an S3 bucket",
	Run: func(cmd *cobra.Command, _ []string) {
		run(cmd)
	},
}

var (
	refreshInterval time.Duration
	trackerURL      string
	s3URL           string
	s3Bucket        string
	objectPrefix    string
	s3Region        string
	s3Secure        bool
)

func init() {
	flags := Cmd.Flags()
	flags.DurationVar(&refreshInterval, "refresh", 30*time.Second, "Refresh interval to discover new snapshots")
	flags.StringVar(&trackerURL, "tracker", "", "URL to tracker API")
	flags.StringVar(&s3URL, "s3-url", "", "URL to S3 API")
	flags.BoolVar(&s3Secure, "s3-secure", true, "Use secure S3 transport")
	flags.StringVar(&s3Region, "region", "", "S3 region (optional)")
	flags.StringVar(&s3Bucket, "bucket", "", "Bucket name")
	flags.StringVar(&objectPrefix, "prefix", "", "Object prefix (optional)")
}

func run(cmd *cobra.Command) {
	log := logger.GetConsoleLogger()
	_ = log

	if trackerURL == "" || s3URL == "" || s3Bucket == "" {
		cobra.CheckErr(cmd.Usage())
		cobra.CheckErr("required argument missing")
	}

	trackerClient := fetch.NewTrackerClient(trackerURL)

	s3Client, err := minio.New(s3URL, &minio.Options{
		Creds:  credentials.NewEnvAWS(),
		Secure: s3Secure,
		Region: s3Region,
	})
	if err != nil {
		log.Fatal("Failed to connect to S3", zap.Error(err))
	}

	uploader := mirror.Uploader{
		S3Client:     s3Client,
		Bucket:       s3Bucket,
		ObjectPrefix: objectPrefix,
	}

	worker := mirror.Worker{
		Tracker:   trackerClient,
		Uploader:  &uploader,
		Log:       log.Named("uploader"),
		Refresh:   refreshInterval,
		SyncCount: 10, // TODO
	}
	worker.Run(context.TODO())
}
