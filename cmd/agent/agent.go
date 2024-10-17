package main

import (
	"fmt"
	"sync"

	config "github.com/igortoigildin/go-metrics-altering/config/agent"
	"github.com/igortoigildin/go-metrics-altering/internal/agent"
	"github.com/igortoigildin/go-metrics-altering/internal/agent/memory"
	"github.com/igortoigildin/go-metrics-altering/internal/models"
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

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Log.Info("error while logading config", zap.Error(err))
		return
	}
	if err := logger.Initialize(cfg.FlagLogLevel); err != nil {
		logger.Log.Fatal("error while initializing logger", zap.Error(err))
		return
	}

	logger.Log.Info("loading metrics...")

	memoryStats := memory.NewMemoryStats()
	var wg sync.WaitGroup
	metricsChan := make(chan models.Metrics, 33)

	wg.Add(1)
	go func() {
		defer wg.Done()
		memoryStats.UpdateRunTimeStat(cfg)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		memoryStats.UpdateCPURAMStat(cfg)
	}()

	for w := 1; w <= cfg.FlagRateLimit; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			agent.SendMetrics(metricsChan, cfg)
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		memoryStats.ReadMetrics(cfg, metricsChan)
	}()

	wg.Wait()
}
