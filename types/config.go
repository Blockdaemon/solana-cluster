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

package types

import (
	"bufio"
	"context"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config describes the root-level config file.
type Config struct {
	ScrapeInterval time.Duration  `json:"scrape_interval" yaml:"scrape_interval"`
	TargetGroups   []*TargetGroup `json:"target_groups" yaml:"target_groups"`
}

// LoadConfig reads the config object from the file system.
func LoadConfig(filePath string) (*Config, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	conf := new(Config)
	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	confErr := decoder.Decode(conf)
	return conf, confErr
}

// TargetGroup explains how to retrieve snapshots from a group of Solana nodes.
type TargetGroup struct {
	Group      string      `json:"group" yaml:"group"`
	Scheme     string      `json:"scheme" yaml:"scheme"`
	APIPath    string      `json:"api_path" yaml:"api_path"`
	BasicAuth  *BasicAuth  `json:"basic_auth" yaml:"basic_auth"`
	BearerAuth *BearerAuth `json:"bearer_auth" yaml:"bearer_auth"`
	TLSConfig  *TLSConfig  `json:"tls_config" yaml:"tls_config"`

	StaticTargets  *StaticTargets  `json:"static_targets" yaml:"static_targets"`
	FileTargets    *FileTargets    `json:"file_targets" yaml:"file_targets"`
	ConsulSDConfig *ConsulSDConfig `json:"consul_sd_config" yaml:"consul_sd_config"`
}

// StaticTargets is a hardcoded list of Solana nodes.
type StaticTargets struct {
	Targets []string `json:"targets" yaml:"targets"`
}

func (s *StaticTargets) DiscoverTargets(_ context.Context) ([]string, error) {
	return s.Targets, nil
}

// FileTargets reads targets from a JSON file.
type FileTargets struct {
	Path string `json:"path" yaml:"path"`
}

func (d *FileTargets) DiscoverTargets(_ context.Context) ([]string, error) {
	f, err := os.Open(d.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scn := bufio.NewScanner(f)
	for scn.Scan() {
		lines = append(lines, strings.TrimSpace(scn.Text()))
	}

	return lines, scn.Err()
}

// ConsulSDConfig configures Consul service discovery.
type ConsulSDConfig struct {
	Server     string `json:"host" yaml:"host"`
	Token      string `json:"token" yaml:"token"`
	TokenFile  string `json:"token_file" yaml:"token_file"`
	Datacenter string `json:"datacenter" yaml:"datacenter"`
	Service    string `json:"service" yaml:"service"`
	Filter     string `json:"filter" yaml:"filter"`
}
