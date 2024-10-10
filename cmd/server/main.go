package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	config "github.com/igortoigildin/go-metrics-altering/config/server"
	server "github.com/igortoigildin/go-metrics-altering/internal/server/api"
	local "github.com/igortoigildin/go-metrics-altering/internal/storage/local"
	psql "github.com/igortoigildin/go-metrics-altering/internal/storage/postgres"
	"github.com/igortoigildin/go-metrics-altering/pkg/logger"
	"go.uber.org/zap"
)

var buildVersion string = "N/A"
var buildDate string = "N/A"
var buildCommit string = "N/A"

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	ctx := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Log.Fatal("error while logading config", zap.Error(err))
	}

	if err := logger.Initialize(cfg.FlagLogLevel); err != nil {
		logger.Log.Fatal("error while initializing logger", zap.Error(err))
	}

	PGStorage := psql.InitPostgresRepo(ctx, cfg)
	localStorage := local.InitLocalStorage()

	if cfg.FlagRestore {
		err := localStorage.LoadMetricsFromFile(cfg.FlagStorePath)
		if err != nil {
			logger.Log.Info("error loading metrics from the file", zap.Error(err))
		}
	}

	go localStorage.SaveAllMetricsToFile(cfg.FlagStoreInterval, cfg.FlagStorePath, cfg.FlagStorePath)

	logger.Log.Info("Starting server on", zap.String("address", cfg.FlagRunAddr))

	var r chi.Router
	// Check whether metrics should be saved to DB or locally.
	if cfg.FlagDBDSN != "" {
		r = server.Router(ctx, cfg, PGStorage)
	} else {
		r = server.Router(ctx, cfg, localStorage)
	}

	http.ListenAndServe(cfg.FlagRunAddr, r)
	if err != nil {
		logger.Log.Fatal("cannot start the server", zap.Error(err))
	}
}
