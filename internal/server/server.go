package server

import (
	"context"
	"net/http"
	"sync"

	config "github.com/igortoigildin/go-metrics-altering/config/server"
	"github.com/igortoigildin/go-metrics-altering/internal/logger"
	"go.uber.org/zap"
)

type MemStorage struct {
	rm      sync.RWMutex
	Gauge   map[string]float64
	Counter map[string]int64
}

func RunServer() {
	ctx := context.Background()
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Log.Fatal("error while logading config", zap.Error(err))
	}
	if err := logger.Initialize(cfg.FlagLogLevel); err != nil {
		logger.Log.Fatal("error while initializing logger", zap.Error(err))
	}
	logger.Log.Info("Running server", zap.String("address", cfg.FlagRunAddr))
	switch cfg.FlagDBDSN {
	case "":
		http.ListenAndServe(cfg.FlagRunAddr, MetricRouter(cfg, ctx))
		if err != nil {
			logger.Log.Fatal("cannot start the server", zap.Error(err))
		}
	default:
		http.ListenAndServe(cfg.FlagRunAddr, routerDB(ctx, cfg))
		if err != nil {
			logger.Log.Fatal("cannot start the server", zap.Error(err))
		}
	}
}

func InitStorage() *MemStorage {
	var m MemStorage
	m.Counter = make(map[string]int64)
	m.Counter["PollCount"] = 0
	m.Gauge = make(map[string]float64)
	return &m
}

func (m *MemStorage) UpdateGaugeMetric(metricName string, metricValue float64) {
	if m.Gauge == nil {
		m.Gauge = make(map[string]float64)
	}
	m.rm.Lock()
	m.Gauge[metricName] = metricValue
	m.rm.Unlock()
}

func (m *MemStorage) UpdateCounterMetric(metricName string, metricValue int64) {
	if m.Counter == nil {
		m.Counter = make(map[string]int64)
	}
	m.rm.Lock()
	m.Counter[metricName] += metricValue
	m.rm.Unlock()
}

func (m *MemStorage) GetGaugeMetricFromMemory(metricName string) float64 {
	m.rm.RLock()
	metric := m.Gauge[metricName]
	m.rm.RUnlock()
	return metric
}

func (m *MemStorage) GetCountMetricFromMemory(metricName string) int64 {
	m.rm.RLock()
	metric := m.Counter[metricName]
	m.rm.RUnlock()
	return metric
}

func (m *MemStorage) CheckIfGaugeMetricPresent(metricName string) bool {
	m.rm.RLock()
	_, ok := m.Gauge[metricName]
	m.rm.RUnlock()
	return ok
}

func (m *MemStorage) CheckIfCountMetricPresent(metricName string) bool {
	m.rm.RLock()
	_, ok := m.Counter[metricName]
	m.rm.RUnlock()
	return ok
}

func ConvertToSingleMap(a map[string]float64, b map[string]int64) map[string]interface{} {
	c := make(map[string]interface{}, 33)
	for i, v := range a {
		c[i] = v
	}
	for j, l := range b {
		c[j] = l
	}
	return c
}
