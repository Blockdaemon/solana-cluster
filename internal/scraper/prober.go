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

package scraper

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"

	"go.blockdaemon.com/solana/cluster-manager/internal/fetch"
	"go.blockdaemon.com/solana/cluster-manager/types"
)

// Prober checks snapshot info from Solana nodes.
type Prober struct {
	group string
	client  *http.Client
	scheme  string
	apiPath string
	header  http.Header
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
		group.BasicAuth.Apply(header)
	}
	if group.BearerAuth != nil {
		group.BearerAuth.Apply(header)
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
		group: 	 group.Group,
		client:  client,
		scheme:  group.Scheme,
		apiPath: group.APIPath,
		header:  header,
	}, nil
}

// Probe fetches the snapshots of a single target.
func (p *Prober) Probe(ctx context.Context, target string) ([]*types.SnapshotInfo, error) {
	u := url.URL{
		Scheme: p.scheme,
		Host:   target,
		Path:   p.apiPath,
	}
	return fetch.NewSidecarClient(u.String()).ListSnapshots(ctx)
}
