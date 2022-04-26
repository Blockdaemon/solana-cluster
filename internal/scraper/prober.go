package scraper

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"net"
	"net/http"
	"net/url"
	"time"

	"go.blockdaemon.com/solana/cluster-manager/internal/fetch"
	"go.blockdaemon.com/solana/cluster-manager/types"
)

// Prober checks snapshot info from Solana nodes.
type Prober struct {
	client       *http.Client
	scheme       string
	snapshotPath string
	header       http.Header
}

func NewProber(group *types.TargetGroup) (*Prober, error) {
	var tlsConfig *tls.Config
	if group.TLSConfig != nil {
		var err error
		tlsConfig, err = group.TLSConfig.Build()
		if err != nil {
			return nil, err
		}
	}

	header := make(http.Header)
	if group.BasicAuth != nil {
		auth := group.BasicAuth.Username + ":" + group.BasicAuth.Password
		header.Add("authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
	}
	if group.BearerAuth != nil {
		header.Add("authorization", "Bearer "+group.BearerAuth.Token)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 5 * time.Second,
			}).DialContext,
			TLSClientConfig:       tlsConfig,
			MaxIdleConnsPerHost:   1,
			MaxConnsPerHost:       3,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			ForceAttemptHTTP2:     true,
		},
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) == 1 {
				return nil
			}
			return http.ErrUseLastResponse
		},
	}

	return &Prober{
		client:       client,
		scheme:       group.Scheme,
		snapshotPath: group.SnapshotPath,
		header:       header,
	}, nil
}

// Probe fetches the snapshots of a single target.
func (p *Prober) Probe(ctx context.Context, target string) ([]*types.SnapshotInfo, error) {
	u := url.URL{
		Scheme: p.scheme,
		Host:   target,
		Path:   p.snapshotPath,
	}
	return fetch.NewClient(u.String()).ListSnapshots(ctx)
}
