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

// Package tracker contains the Solana cluster tracker logic.
package tracker

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.blockdaemon.com/solana/cluster-manager/internal/index"
	"go.blockdaemon.com/solana/cluster-manager/types"
)

// Handler implements the tracker API methods.
type Handler struct {
	DB *index.DB
}

// NewHandler creates a new tracker API using the provided database.
func NewHandler(db *index.DB) *Handler {
	return &Handler{DB: db}
}

// RegisterHandlers registers this API with Gin web framework.
func (h *Handler) RegisterHandlers(group gin.IRoutes) {
	group.GET("/best_snapshots", h.GetBestSnapshots)
}

// GetBestSnapshots returns the currently available best snapshots.
func (h *Handler) GetBestSnapshots(c *gin.Context) {
	var query struct {
		Max int `form:"max"`
	}
	if err := c.BindQuery(&query); err != nil {
		return
	}
	const maxItems = 25
	if query.Max < 0 || query.Max > 25 {
		query.Max = maxItems
	}
	entries := h.DB.GetBestSnapshots(query.Max)
	sources := make([]types.SnapshotSource, len(entries))
	for i, entry := range entries {
		sources[i] = types.SnapshotSource{
			SnapshotInfo: *entry.Info,
			Target:       entry.Target,
			UpdatedAt:    entry.UpdatedAt,
		}
	}
	c.JSON(http.StatusOK, sources)
}
