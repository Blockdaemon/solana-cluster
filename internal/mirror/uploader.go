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

package mirror

import (
	"context"

	"github.com/minio/minio-go/v7"
	"go.blockdaemon.com/solana/cluster-manager/internal/fetch"
)

// Uploader streams snapshots from a node to an S3 mirror.
type Uploader struct {
	S3Client     *minio.Client
	Bucket       string
	ObjectPrefix string
}

// StatSnapshot checks whether a snapshot has been uploaded already.
func (u *Uploader) StatSnapshot(ctx context.Context, fileName string) (minio.ObjectInfo, error) {
	objectName := u.getSnapshotObjectName(fileName)
	return u.S3Client.StatObject(ctx, u.Bucket, objectName, minio.StatObjectOptions{})
}

// UploadSnapshot streams a snapshot from the given sidecar client to S3.
func (u *Uploader) UploadSnapshot(ctx context.Context, sourceClient *fetch.SidecarClient, fileName string) (minio.UploadInfo, error) {
	res, err := sourceClient.StreamSnapshot(ctx, fileName)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return minio.UploadInfo{}, err
	}
	objectName := u.getSnapshotObjectName(fileName)
	return u.S3Client.PutObject(ctx, u.Bucket, objectName, res.Body, res.ContentLength, minio.PutObjectOptions{})
}

func (u *Uploader) getSnapshotObjectName(fileName string) string {
	return u.ObjectPrefix + fileName
}
