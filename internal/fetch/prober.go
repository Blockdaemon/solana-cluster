// Package fetch contains a sidecar API client.
package fetch

import (
	"context"
	"fmt"
	"net/http"

	"go.blockdaemon.com/solana/cluster-manager/types"
	"gopkg.in/resty.v1"
)

// TODO rewrite this package to use OpenAPI code gen

// Client accesses the sidecar API.
type Client struct {
	Resty   *resty.Client
	BaseURL string
}

func NewClient(sidecarURL string) *Client {
	return &Client{
		Resty:   resty.New().SetHostURL(sidecarURL),
		BaseURL: sidecarURL,
	}
}

func (c *Client) ListSnapshots(ctx context.Context) (infos []*types.SnapshotInfo, err error) {
	res, err := c.Resty.R().
		SetContext(ctx).
		SetHeader("accept", "application/json").
		SetResult(&infos).
		Get("/v1/snapshots")
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("get snapshots: %s", res.Status())
	}
	return
}
