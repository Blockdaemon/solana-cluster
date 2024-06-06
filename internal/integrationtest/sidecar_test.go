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

package integrationtest

import (
	"context"
	"encoding/binary"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.blockdaemon.com/solana/cluster-manager/internal/fetch"
	"go.blockdaemon.com/solana/cluster-manager/internal/ledgertest"
	"go.blockdaemon.com/solana/cluster-manager/internal/sidecar"
	"go.blockdaemon.com/solana/cluster-manager/types"
	"go.uber.org/zap/zaptest"
	"gopkg.in/resty.v1"
)

// TestSidecar creates
// a sidecar server with a fake ledger dir,
// and a sidecar client
func TestSidecar(t *testing.T) {
	server, root := newSidecar(t, 100)
	defer server.Close()
	fmt.Println("Server url", server.URL)
	client := fetch.NewSidecarClientWithOpts(server.URL,
		fetch.SidecarClientOpts{Resty: resty.NewWithClient(server.Client())})

	ctx := context.TODO()

	t.Run("ListSnapshots", func(t *testing.T) {
		infos, err := client.ListSnapshots(ctx)
		require.NoError(t, err)
		assert.Equal(t,
			[]*types.SnapshotInfo{
				{
					Slot:      100,
					BaseSlot:  100,
					Hash:      solana.MustHashFromBase58("7jMmeXZSNcWPrB2RsTdeXfXrsyW5c1BfPjqoLW2X5T7V"),
					TotalSize: 1,
					Files: []*types.SnapshotFile{
						{
							FileName: "snapshot-100-7jMmeXZSNcWPrB2RsTdeXfXrsyW5c1BfPjqoLW2X5T7V.tar.bz2",
							Slot:     100,
							Hash:     solana.MustHashFromBase58("7jMmeXZSNcWPrB2RsTdeXfXrsyW5c1BfPjqoLW2X5T7V"),
							Ext:      ".tar.bz2",
							Size:     1,
							ModTime:  &root.DummyTime,
						},
					},
				},
			},
			infos,
		)
	})

	t.Run("DownloadSnapshot", func(t *testing.T) {
		res, err := client.StreamSnapshot(ctx, "snapshot-100-7jMmeXZSNcWPrB2RsTdeXfXrsyW5c1BfPjqoLW2X5T7V.tar.bz2")
		require.NoError(t, err)
		assert.Equal(t, int64(1), res.ContentLength)
		require.NoError(t, res.Body.Close())
	})
}

func newSidecar(t *testing.T, slots ...uint64) (server *httptest.Server, root *ledgertest.FS) {
	root = ledgertest.NewFS(t)
	for _, slot := range slots {
		var fakeBin [8]byte
		binary.LittleEndian.PutUint64(fakeBin[:], slot)
		root.AddFakeFile(t, fmt.Sprintf("snapshot-%d-%s.tar.bz2", slot, solana.HashFromBytes(fakeBin[:])))
	}
	ledgerDir := root.GetLedgerDir(t)

	handler := &sidecar.SnapshotHandler{
		LedgerDir: ledgerDir,
		Log:       zaptest.NewLogger(t),
	}

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	handler.RegisterHandlers(engine.Group("/v1"))
	server = httptest.NewServer(engine)
	return
}
