package cmd

import (
	"github.com/spf13/cobra"
	"go.blockdaemon.com/solana/cluster-manager/internal/cmd/sidecar"
)

var Cmd = cobra.Command{
	Use: "solana-snapshots",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func init() {
	Cmd.AddCommand(
		&sidecar.Cmd,
	)
}
