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

package cmd

import (
	"github.com/spf13/cobra"
	"go.blockdaemon.com/solana/cluster-manager/internal/cmd/fetch"
	"go.blockdaemon.com/solana/cluster-manager/internal/cmd/sidecar"
	"go.blockdaemon.com/solana/cluster-manager/internal/cmd/tracker"
)

var Cmd = cobra.Command{
	Use: "solana-snapshots",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func init() {
	Cmd.AddCommand(
		&fetch.Cmd,
		&sidecar.Cmd,
		&tracker.Cmd,
	)
}
