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

// Package discovery provides service discovery basics.
package discovery

import (
	"context"
	"fmt"

	"go.blockdaemon.com/solana/cluster-manager/types"
)

// Discoverer returns a list of host:port combinations for all targets.
type Discoverer interface {
	DiscoverTargets(ctx context.Context) ([]string, error)
}

// Simple backends can be found in ../types/config.go

// NewFromConfig attempts to create a discoverer from config.
func NewFromConfig(t *types.TargetGroup) (Discoverer, error) {
	if t.StaticTargets != nil {
		return t.StaticTargets, nil
	}
	if t.FileTargets != nil {
		return t.FileTargets, nil
	}
	if t.ConsulSDConfig != nil {
		return NewConsulFromConfig(t.ConsulSDConfig)
	}
	return nil, fmt.Errorf("missing config")
}
