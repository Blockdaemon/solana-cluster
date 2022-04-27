package sidecar

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.blockdaemon.com/solana/cluster-manager/internal/fetch"
	"go.blockdaemon.com/solana/cluster-manager/internal/ledgertest"
	"go.blockdaemon.com/solana/cluster-manager/types"
	"go.uber.org/zap/zaptest"
	"gopkg.in/resty.v1"
)

func TestHandler(t *testing.T) {
	root := ledgertest.NewFS(t)
	root.AddFakeFile(t, "snapshot-100-AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr.tar.bz2")
	ledgerDir := root.GetLedgerDir(t)

	handler := &Handler{
		LedgerDir: ledgerDir,
		Log:       zaptest.NewLogger(t),
	}

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	handler.RegisterHandlers(engine.Group("/v1"))

	server := httptest.NewServer(engine)
	client := fetch.NewClientWithResty(
		resty.NewWithClient(server.Client()).
			SetHostURL(server.URL))

	ctx := context.TODO()

	t.Run("ListSnapshots", func(t *testing.T) {
		infos, err := client.ListSnapshots(ctx)
		require.NoError(t, err)
		assert.Equal(t,
			[]*types.SnapshotInfo{
				{
					Slot:      100,
					Hash:      solana.MustHashFromBase58("AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr"),
					TotalSize: 1,
					Files: []*types.SnapshotFile{
						{
							FileName: "snapshot-100-AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr.tar.bz2",
							Slot:     100,
							Hash:     solana.MustHashFromBase58("AvFf9oS8A8U78HdjT9YG2sTTThLHJZmhaMn2g8vkWYnr"),
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
}
