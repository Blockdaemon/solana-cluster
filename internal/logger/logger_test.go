package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetLogger(t *testing.T) {
	testCases := []struct {
		name   string
		format string
	}{
		{name: "JSON", format: "json"},
		{name: "Console", format: "console"},
		{name: "Default", format: ""},
		{name: "Unknown", format: "bla"},
	}
	for _, tc := range testCases {
		logFormat = tc.format // global
		log := GetLogger()
		require.NotNil(t, log)
	}
}
