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

// Package mirror maintains a copy of snapshots in S3.
package mirror

import (
	"context"
	"net/http"
	"time"

	"github.com/minio/minio-go/v7"
	"go.blockdaemon.com/solana/cluster-manager/internal/fetch"
	"go.blockdaemon.com/solana/cluster-manager/types"
	"go.uber.org/zap"
)

// Worker mirrors snapshots from nodes to S3.
type Worker struct {
	Tracker  *fetch.TrackerClient
	Uploader *Uploader
	Log      *zap.Logger

	Refresh   time.Duration
	SyncCount int
}

func NewWorker(tracker *fetch.TrackerClient, uploader *Uploader) *Worker {
	return &Worker{
		Tracker:   tracker,
		Uploader:  uploader,
		Log:       zap.NewNop(),
		Refresh:   15 * time.Second,
		SyncCount: 10,
	}
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.Refresh)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *Worker) tick(ctx context.Context) {
	sources, err := w.Tracker.GetBestSnapshots(ctx, w.SyncCount)
	if err != nil {
		w.Log.Error("Failed to find new snapshots", zap.Error(err))
		return
	}

	type fileSource struct {
		target string
		file   *types.SnapshotFile
	}
	files := make(map[uint64]fileSource)
	for _, src := range sources {
		for _, file := range src.Files {
			if _, ok := files[file.Slot]; !ok {
				files[file.Slot] = fileSource{
					target: src.Target,
					file:   file,
				}
			}
		}
	}

	for _, src := range files {
		// TODO Consider using a semaphore
		job := UploadJob{
			Provider: src.target,
			File:     src.file,
			Uploader: w.Uploader,
		}
		go job.Run(ctx)
	}
}

type UploadJob struct {
	Provider string
	File     *types.SnapshotFile
	Uploader *Uploader
	Log      *zap.Logger
}

func (j *UploadJob) Run(ctx context.Context) {
	fileName := j.File.FileName
	log := j.Log.With(zap.String("snapshot", fileName))
	// Check if snapshot exists.
	stat, statErr := j.Uploader.StatSnapshot(ctx, fileName)
	if statErr == nil {
		log.Debug("Already uploaded",
			zap.Time("last_modified", stat.LastModified))
		return
	}
	statResp := minio.ToErrorResponse(statErr)
	if statResp.StatusCode != http.StatusNotFound {
		log.Error("Unexpected error", zap.Error(statErr))
		return
	}

	// TODO use client factory
	sidecarClient := fetch.NewSidecarClient(j.Provider)

	beforeUpload := time.Now()
	uploadInfo, err := j.Uploader.UploadSnapshot(ctx, sidecarClient, fileName)
	uploadDuration := time.Since(beforeUpload)
	if err != nil {
		log.Error("Upload failed", zap.Error(err),
			zap.Duration("upload_duration", uploadDuration))
		return
	}
	log.Info("Upload succeeded",
		zap.String("bucket", uploadInfo.Bucket),
		zap.String("object", uploadInfo.Key),
		zap.Duration("upload_duration", uploadDuration))
}
