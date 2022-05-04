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

// Package tracker provides the `tracker` command.
package tracker

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"go.blockdaemon.com/solana/cluster-manager/internal/index"
	"go.blockdaemon.com/solana/cluster-manager/internal/logger"
	"go.blockdaemon.com/solana/cluster-manager/internal/scraper"
	"go.blockdaemon.com/solana/cluster-manager/internal/tracker"
	"go.blockdaemon.com/solana/cluster-manager/types"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var Cmd = cobra.Command{
	Use:   "tracker",
	Short: "Snapshot tracker server",
	Long: "Connects to sidecars on nodes and scrapes the available snapshot versions.\n" +
		"Provides an API allowing fetch jobs to find the latest snapshots.\n" +
		"Do not expose this API publicly.",
	Run: func(_ *cobra.Command, _ []string) {
		run()
	},
}

var (
	configPath     string
	internalListen string
	listen         string
)

func init() {
	flags := Cmd.Flags()
	flags.StringVar(&configPath, "config", "", "Path to config file")
	flags.StringVar(&internalListen, "internal-listen", ":8457", "Internal listen URL")
	flags.StringVar(&listen, "listen", ":8458", "Listen URL")
	flags.AddFlagSet(logger.Flags)
}

func run() {
	log := logger.GetLogger()

	// Install signal handlers.
	onReload := make(chan os.Signal, 1)
	signal.Notify(onReload, syscall.SIGHUP)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Install HTTP handlers.
	http.HandleFunc("/reload", func(wr http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(wr, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		onReload <- syscall.SIGHUP
		http.Error(wr, "reloaded", http.StatusOK)
	})
	httpErrLog, err := zap.NewStdLogAt(log.Named("prometheus"), zap.ErrorLevel)
	if err != nil {
		panic(err.Error())
	}
	http.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			ErrorLog: httpErrLog,
		},
	))

	// Create result collector.
	db := index.NewDB()
	collector := scraper.NewCollector(db)
	collector.Log = log.Named("collector")
	collector.Start()
	defer collector.Close()

	gin.SetMode(gin.ReleaseMode)
	server := gin.New()
	httpLog := log.Named("http")
	server.Use(ginzap.Ginzap(httpLog, time.RFC3339, true))
	server.Use(ginzap.RecoveryWithZap(httpLog, false))

	handler := tracker.NewHandler(db)
	handler.RegisterHandlers(server.Group("/v1"))

	// Start services.
	group, ctx := errgroup.WithContext(ctx)
	if internalListen != "" {
		httpLog.Info("Starting internal server", zap.String("listen", internalListen))
	}
	runGroupServer(ctx, group, internalListen, nil) // default handler
	httpLog.Info("Starting server", zap.String("listen", listen))
	runGroupServer(ctx, group, listen, server) // public handler

	// Create config reloader.
	config, err := types.LoadConfig(configPath)
	if err != nil {
		log.Fatal("Failed to load config", zap.Error(err))
	}

	// Create scrape managers.
	manager := scraper.NewManager(collector.Probes())
	manager.Log = log.Named("scraper")
	manager.Update(config)

	// TODO Config reloading

	// Wait until crash or graceful exit.
	if err := group.Wait(); err != nil {
		log.Error("Crashed", zap.Error(err))
	} else {
		log.Info("Shutting down")
	}
}

func runGroupServer(ctx context.Context, group *errgroup.Group, listen string, handler http.Handler) {
	group.Go(func() error {
		server := http.Server{
			Addr:    listen,
			Handler: handler,
		}
		go func() {
			<-ctx.Done()
			_ = server.Close()
		}()
		if err := server.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
			return nil
		} else {
			return err
		}
	})
}
