package main

import (
	"github.com/spf13/cobra"
	"go.blockdaemon.com/solana/cluster-manager/internal/cmd"
)

func main() {
	cobra.CheckErr(cmd.Cmd.Execute())
}
