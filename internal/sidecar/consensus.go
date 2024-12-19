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

package sidecar

import (
	"io"
	"net/http"

	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ConsensusHandler implements the consensus-related sidecar API methods.
type ConsensusHandler struct {
	RpcWsUrl string
	Log      *zap.Logger
}

// NewConsensusHandler creates a new sidecar consensus API handler using the provided WS RPC and logger.
func NewConsensusHandler(rpcWsUrl string, log *zap.Logger) *ConsensusHandler {
	return &ConsensusHandler{
		RpcWsUrl: rpcWsUrl,
		Log:      log,
	}
}

// RegisterHandlers registers this API with Gin web framework.
func (h *ConsensusHandler) RegisterHandlers(group gin.IRoutes) {
	group.GET("/slot_updates", h.GetSlotUpdates)
}

// GetSlotUpdates streams RPC "slotsUpdatesSubscribe" events via SSE.
func (h *ConsensusHandler) GetSlotUpdates(c *gin.Context) {
	ctx := c.Request.Context()

	conn, err := ws.Connect(ctx, h.RpcWsUrl)
	if err != nil {
		h.Log.Error("Failed to connect to Solana RPC WebSocket", zap.Error(err))
		c.AbortWithStatus(http.StatusBadGateway)
		return
	}
	defer conn.Close()

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	slotUpdates, err := conn.SlotsUpdatesSubscribe()
	if err != nil {
		h.Log.Error("Failed to connect to subscribe to slot updates", zap.Error(err))
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	defer slotUpdates.Unsubscribe()

	c.Stream(func(w io.Writer) bool {
		update, err := slotUpdates.Recv(c)
		if err != nil {
			h.Log.Error("Failed to receive slot update event", zap.Error(err))
			return false
		}
		c.SSEvent("slot_update", update)
		return true
	})
}
