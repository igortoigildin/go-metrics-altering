package agent

import (
	"errors"
	"sync"
	"time"

	config "github.com/igortoigildin/go-metrics-altering/config/agent"
	"github.com/igortoigildin/go-metrics-altering/internal/agent/memory"
	"github.com/igortoigildin/go-metrics-altering/internal/logger"
	"github.com/igortoigildin/go-metrics-altering/internal/models"
	"go.uber.org/zap"
)

var (
	ErrConnectionFailed = errors.New("connection failed")
)

func RunAgent() {
	cfg := Initialize()
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
			sendMetrics(metricsChan, cfg)
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		memoryStats.ReadMetrics(cfg, metricsChan)
	}()

	wg.Wait()
}

func sendMetrics(metricsChan <-chan models.Metrics, cfg *config.ConfigAgent) {
	for {
		time.Sleep(cfg.PauseDuration)
		for metric := range metricsChan {
			switch metric.MType {
			case config.CountType:
				err := sendURLCounter(cfg, int(*metric.Delta))
				if err != nil {
					logger.Log.Info("unexpected sending url counter metric error:", zap.Error(err))
				}
				err = SendJSONCounter(int(*metric.Delta), cfg)
				if err != nil {
					logger.Log.Info("unexpected sending json counter metric error:", zap.Error(err))
				}
			case config.GaugeType:
				err := SendURLGauge(cfg, *metric.Value, metric.ID)
				if err != nil {
					logger.Log.Info("unexpected sending url gauge metric error:", zap.Error(err))
				}
				err = SendJSONGauge(metric.ID, cfg, *metric.Value)
				if err != nil {
					logger.Log.Info("unexpected sending json gauge metric error:", zap.Error(err))
				}
			}
		}
	}
}

// Initializes logger and loads config
func Initialize() *config.ConfigAgent {
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Log.Fatal("error while logading config", zap.Error(err))
	}
	if err := logger.Initialize(cfg.FlagLogLevel); err != nil {
		logger.Log.Fatal("error while initializing logger", zap.Error(err))
	}
	return cfg
}
