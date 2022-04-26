package types

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"go.blockdaemon.com/solana/cluster-manager/internal/discovery"
)

// Config describes the root-level config file.
type Config struct {
	ScrapeInterval time.Duration  `json:"scrape_interval"`
	TargetGroups   []*TargetGroup `json:"target_groups"`
}

// LoadConfig reads the config object from the file system.
func LoadConfig(filePath string) (*Config, error) {
	configBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	conf := new(Config)
	jsonErr := json.Unmarshal(configBytes, conf)
	return conf, jsonErr
}

// TargetGroup explains how to retrieve snapshots from a group of Solana nodes.
type TargetGroup struct {
	Group        string      `json:"group"`
	Scheme       string      `json:"scheme"`
	SnapshotPath string      `json:"snapshot_path"`
	BasicAuth    *BasicAuth  `json:"basic_auth"`
	BearerAuth   *BearerAuth `json:"bearer_auth"`
	TLSConfig    *TLSConfig  `json:"tls_config"`

	StaticTargets *StaticTargets `json:"static_targets"`
	FileTargets   *FileTargets   `json:"file_targets"`
}

func (t *TargetGroup) Discoverer() discovery.Discoverer {
	if t.StaticTargets != nil {
		return t.StaticTargets
	}
	if t.FileTargets != nil {
		return t.FileTargets
	}
	return nil
}

// StaticTargets is a hardcoded list of Solana nodes.
type StaticTargets struct {
	Targets []string `json:"targets"`
}

func (s *StaticTargets) DiscoverTargets(_ context.Context) ([]string, error) {
	return s.Targets, nil
}

// FileTargets reads targets from a JSON file.
type FileTargets struct {
	Path string `json:"path"`
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
