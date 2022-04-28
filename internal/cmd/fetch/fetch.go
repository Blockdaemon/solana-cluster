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

// Package fetch provides the `fetch` command.
package fetch

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.blockdaemon.com/solana/cluster-manager/internal/fetch"
)

var Cmd = cobra.Command{
	Use:   "fetch",
	Short: "Fetch snapshot from another node",
	Run: func(_ *cobra.Command, _ []string) {
		run()
	},
}

var (
	ledgerDir  string
	trackerURL string
)

func init() {
	flags := Cmd.Flags()
	flags.StringVar(&ledgerDir, "ledger", "", "Path to ledger dir")
	flags.StringVar(&trackerURL, "tracker", "", "Tracker URL")
}

func run() {
	ctx := context.TODO()
	client := fetch.NewTrackerClient(trackerURL)
	snapshots, err := client.GetBestSnapshots(ctx)
	cobra.CheckErr(err)

	buf, _ := json.MarshalIndent(snapshots, "", "\t")
	fmt.Printf("%s\n", buf)
}
