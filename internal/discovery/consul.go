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

package discovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/consul/api"
	"go.blockdaemon.com/solana/cluster-manager/types"
)

// Consul service discovery backend
type Consul struct {
	Client  *api.Client
	Service string

	Datacenter string // Consul datacenter (dc param)
	Filter     string // Consul filter expression (filter param)
}

// NewConsulFromConfig invokes NewConsul using typed config.
func NewConsulFromConfig(config *types.ConsulSDConfig) (*Consul, error) {
	client, err := api.NewClient(&api.Config{
		Address:   config.Server,
		Token:     config.Token,
		TokenFile: config.TokenFile,
	})
	if err != nil {
		return nil, err
	}
	sd := NewConsul(client, config.Service)
	sd.Datacenter = config.Datacenter
	sd.Filter = config.Filter
	return sd, nil
}

// NewConsul creates a new service discovery provider for Solana cluster
func NewConsul(client *api.Client, service string) *Consul {
	return &Consul{
		Client:  client,
		Service: service,
	}
}

// DiscoverTargets queries Consul Catalog API to find nodes.
// Returns a list of targets referred to by IP addresses.
func (c *Consul) DiscoverTargets(ctx context.Context) ([]string, error) {
	services, _, err := c.Client.Catalog().Service(c.Service, "", (&api.QueryOptions{
		Datacenter: c.Datacenter,
		Filter:     c.Filter,
	}).WithContext(ctx))
	if err != nil {
		return nil, err
	}
	targets := make([]string, 0, len(services))
	for _, service := range services {
		targets = append(targets, fmt.Sprintf("%s:%d", service.Address, service.ServicePort))
	}
	return targets, nil
}
