package types

import (
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	_, testFile, _, ok := runtime.Caller(0)
	assert.True(t, ok)
	exampleConfig := filepath.Join(filepath.Dir(testFile), "../example-config.yml")

	actual, err := LoadConfig(exampleConfig)
	require.NoError(t, err)

	expected := &Config{
		ScrapeInterval: 15 * time.Second,
		TargetGroups: []*TargetGroup{
			{
				Group:  "mainnet",
				Scheme: "http",
				StaticTargets: &StaticTargets{
					Targets: []string{
						"solana-mainnet-1.example.org:8899",
						"solana-mainnet-2.example.org:8899",
						"solana-mainnet-3.example.org:8899",
					},
				},
			},
		},
	}

	assert.Equal(t, expected, actual)
}
