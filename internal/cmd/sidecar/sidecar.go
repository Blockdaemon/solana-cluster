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

// Package sidecar provides the `sidecar` command.
package sidecar

import (
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"go.blockdaemon.com/solana/cluster-manager/internal/logger"
	"go.blockdaemon.com/solana/cluster-manager/internal/netx"
	"go.blockdaemon.com/solana/cluster-manager/internal/sidecar"
	"go.uber.org/zap"
)

var Cmd = cobra.Command{
	Use:   "sidecar",
	Short: "Snapshot node sidecar",
	Long:  "Runs on a Solana node and serves available snapshot archives.",
	Run: func(_ *cobra.Command, _ []string) {
		run()
	},
}

var (
	netInterface string
	listenPort   uint16
	ledgerDir    string
)

func init() {
	flags := Cmd.Flags()
	flags.StringVar(&netInterface, "interface", "", "Only accept connections from this interface")
	flags.Uint16Var(&listenPort, "port", 13080, "Listen port")
	flags.StringVar(&ledgerDir, "ledger", "", "Path to ledger dir")
	flags.AddFlagSet(logger.Flags)
}

func run() {
	log := logger.GetLogger()
	listener, listenAddrs, err := netx.ListenTCPInterface("tcp", netInterface, listenPort)
	if err != nil {
		cobra.CheckErr(err)
	}
	for _, addr := range listenAddrs {
		log.Info("Listening for conns", zap.Stringer("addr", &addr))
	}

	gin.SetMode(gin.ReleaseMode)
	server := gin.New()
	httpLog := log.Named("http")
	server.Use(ginzap.Ginzap(httpLog, time.RFC3339, true))
	server.Use(ginzap.RecoveryWithZap(httpLog, false))

	groupV1 := server.Group("/v1")
	handler := sidecar.NewHandler(ledgerDir, httpLog)
	handler.RegisterHandlers(groupV1)

	err = server.RunListener(listener)
	log.Error("Server stopped", zap.Error(err))
}
