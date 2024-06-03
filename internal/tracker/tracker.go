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
	"context"
	"net/http"
	"time"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gin-gonic/gin"
	"go.blockdaemon.com/solana/cluster-manager/internal/index"
	"go.blockdaemon.com/solana/cluster-manager/types"
)

// Handler implements the tracker API methods.
type Handler struct {
	DB             *index.DB
	RPC            *rpc.Client
	MaxSnapshotAge uint64
}

// NewHandler creates a new tracker API using the provided database.
func NewHandler(db *index.DB, rpcURL string, maxSnapshotAge uint64) *Handler {
	return &Handler{DB: db, RPC: rpc.New(rpcURL), MaxSnapshotAge: maxSnapshotAge}
}

// RegisterHandlers registers this API with Gin web framework.
func (h *Handler) RegisterHandlers(group gin.IRoutes) {
	group.GET("/snapshots", h.GetSnapshots)
	group.GET("/best_snapshots", h.GetBestSnapshots)
	group.GET("/health", h.Health)
}

func (h *Handler) createJson(c *gin.Context, entries []*index.SnapshotEntry) {
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

func (h *Handler) GetSnapshots(c *gin.Context) {
	var query struct {
		Slot uint64 `form:"slot"`
	}
	if err := c.BindQuery(&query); err != nil {
		return
	}

	var entries []*index.SnapshotEntry
	 if query.Slot == 0 {
		entries = h.DB.GetAllSnapshots()
	} else {
		entries = h.DB.GetSnapshotsAtSlot(query.Slot)
	}

	h.createJson(c, entries)
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
	if query.Max < 0 || query.Max > maxItems {
		query.Max = maxItems
	}
	entries := h.DB.GetBestSnapshots(query.Max)
	h.createJson(c, entries)
}

func (h *Handler) Health(c *gin.Context) {
	var query struct {
		Max int `form:"max"`
	}
	if err := c.BindQuery(&query); err != nil {
		return
	}
	query.Max = 1
	entries := h.DB.GetBestSnapshots(query.Max)

	var health struct {
		MaxSnapshot uint64
		CurrentSlot uint64
		Health      string
	}

	if len(entries) <= 0 {
		health.Health = "no snapshots found"
		c.JSON(http.StatusInternalServerError, health)
	} else {
		health.MaxSnapshot = entries[0].Info.Slot

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()
		out, err := h.RPC.GetSlot(
			ctx,
			rpc.CommitmentFinalized,
		)
		if err != nil {
			health.Health = "rpc unhealthy"
			c.JSON(http.StatusBadGateway, health)
			return
		}
		health.CurrentSlot = out

		if (health.CurrentSlot - health.MaxSnapshot) > h.MaxSnapshotAge {
			health.Health = "snapshot too old"
			c.JSON(http.StatusServiceUnavailable, health)
			return
		} else {
			health.Health = "healthy"
			c.JSON(http.StatusOK, health)
			return
		}
	}
}
