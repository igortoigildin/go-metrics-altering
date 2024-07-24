package agent

import (
	config "github.com/igortoigildin/go-metrics-altering/config/agent"
	"github.com/igortoigildin/go-metrics-altering/internal/agent/memory"
	"github.com/igortoigildin/go-metrics-altering/internal/logger"
	"go.uber.org/zap"
)

func RunAgent() {
    cfg, err := config.LoadConfig()
    if err != nil {
        logger.Log.Fatal("error while logading config", zap.Error(err))
    }
    if err := logger.Initialize(cfg.FlagLogLevel); err != nil {
        logger.Log.Fatal("error while initializing logger", zap.Error(err))
    }
    MemoryStats := memory.NewMemoryStats()
    go MemoryStats.UpdateMetrics(cfg)    // updating metrics in memory every pollInterval
    go MemoryStats.SendJSONMetrics(cfg)  // v2 - metrics in body json
    go MemoryStats.SendBatchMetrics(cfg) // v3 - sending batchs of metrics json
    MemoryStats.SendURLmetrics(cfg)      // v1 - metrics in url path
}

