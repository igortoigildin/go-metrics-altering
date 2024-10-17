package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	config "github.com/igortoigildin/go-metrics-altering/config/server"
	server "github.com/igortoigildin/go-metrics-altering/internal/server/api"
	storage "github.com/igortoigildin/go-metrics-altering/internal/server/api"
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
		log.Println("error while logading config", err)
		return
	}

	if err = logger.Initialize(cfg.FlagLogLevel); err != nil {
		log.Println("error while initializing logger", err)
		return
	}

	var r chi.Router
	storage := storage.New(cfg)
	r = server.Router(ctx, cfg, storage)

	logger.Log.Info("Starting server on", zap.String("address", cfg.FlagRunAddr))

	err = http.ListenAndServe(cfg.FlagRunAddr, r)
	if err != nil {
		logger.Log.Error("cannot start the server", zap.Error(err))
		return
	}
}
