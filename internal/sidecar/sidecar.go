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
}

// ListSnapshots is an API handler listing available snapshots on the node.
func (s *Handler) ListSnapshots(c *gin.Context) {
	infos, err := ledger.ListSnapshots(s.LedgerDir)
	if err != nil {
		s.Log.Error("Failed to list snapshots", zap.Error(err))
		return
	}
	if infos == nil {
		infos = make([]*types.SnapshotInfo, 0)
	}
	c.JSON(http.StatusOK, infos)
}
