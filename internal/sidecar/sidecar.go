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

// Package sidecar contains the Solana cluster sidecar logic.
package sidecar

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.blockdaemon.com/solana/cluster-manager/internal/ledger"
	"go.blockdaemon.com/solana/cluster-manager/types"
	"go.uber.org/zap"
)

// Handler implements the sidecar API methods.
type Handler struct {
	LedgerDir fs.FS
	Log       *zap.Logger
}

// NewHandler creates a new sidecar API using the provided ledger dir and logger.
func NewHandler(ledgerDir string, log *zap.Logger) *Handler {
	return &Handler{
		LedgerDir: os.DirFS(ledgerDir),
		Log:       log,
	}
}

// RegisterHandlers registers this API with Gin web framework.
func (s *Handler) RegisterHandlers(group gin.IRoutes) {
	group.GET("/snapshots", s.ListSnapshots)
	group.HEAD("/snapshot.tar.bz2", s.DownloadBestSnapshot)
	group.GET("/snapshot.tar.bz2", s.DownloadBestSnapshot)
	group.HEAD("/snapshot.tar.zst", s.DownloadBestSnapshot)
	group.GET("/snapshot.tar.zst", s.DownloadBestSnapshot)
	group.HEAD("/snapshot/:name", s.DownloadSnapshot)
	group.GET("/snapshot/:name", s.DownloadSnapshot)
}

// ListSnapshots is an API handler listing available snapshots on the node.
func (s *Handler) ListSnapshots(c *gin.Context) {
	infos, err := ledger.ListSnapshots(s.LedgerDir)
	if err != nil {
		s.Log.Error("Failed to list snapshots", zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if infos == nil {
		infos = make([]*types.SnapshotInfo, 0)
	}
	c.JSON(http.StatusOK, infos)
}

// DownloadBestSnapshot selects the best full snapshot and sends it to the client.
func (s *Handler) DownloadBestSnapshot(c *gin.Context) {
	files, err := ledger.ListSnapshotFiles(s.LedgerDir)
	if err != nil {
		s.Log.Error("Failed to list snapshot files", zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	for _, file := range files {
		if file.IsFull() {
			s.serveSnapshot(c, file.FileName)
			return
		}
	}
	c.String(http.StatusAccepted, "no snapshot available")
}

// DownloadSnapshot sends a snapshot to the client.
func (s *Handler) DownloadSnapshot(c *gin.Context) {
	// Parse name and reject odd requests.
	name := c.Param("name")
	snapshot := ledger.ParseSnapshotFileName(name)
	if snapshot == nil {
		returnSnapshotNotFound(c)
		return
	}
	switch snapshot.Ext {
	case ".tar.bz2", ".tar.gz", "tar.zst", ".tar.xz", ".tar":
		// ok
	default:
		returnSnapshotNotFound(c)
		return
	}

	s.serveSnapshot(c, name)
}

func (s *Handler) serveSnapshot(c *gin.Context, name string) {
	log := s.Log.With(zap.String("snapshot", name))

	// Open file.
	baseFile, err := s.LedgerDir.Open(name)
	if errors.Is(err, fs.ErrNotExist) {
		log.Info("Requested snapshot not found")
		returnSnapshotNotFound(c)
		return
	} else if err != nil {
		log.Error("Failed to open file", zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer baseFile.Close()
	snapFile, ok := baseFile.(io.ReadSeeker)
	if !ok {
		log.Error("Snapshot file is not an io.ReedSeeker")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	info, err := baseFile.Stat()
	if err != nil {
		log.Warn("Stat failed on snapshot", zap.String("snapshot", name), zap.Error(err))
		returnSnapshotNotFound(c)
		return
	}

	http.ServeContent(c.Writer, c.Request, name, info.ModTime(), snapFile)
}

func returnSnapshotNotFound(c *gin.Context) {
	c.String(http.StatusNotFound, "snapshot not found")
}
