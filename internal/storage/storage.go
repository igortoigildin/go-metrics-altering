package storage

import (
	"context"

	config "github.com/igortoigildin/go-metrics-altering/config/server"
	"github.com/igortoigildin/go-metrics-altering/internal/models"
	local "github.com/igortoigildin/go-metrics-altering/internal/storage/inmemory"
	psql "github.com/igortoigildin/go-metrics-altering/internal/storage/postgres"
	"github.com/igortoigildin/go-metrics-altering/pkg/logger"
	"go.uber.org/zap"
)

type Storage interface {
	Update(ctx context.Context, metricType string, metricName string, metricValue any) error
	Get(ctx context.Context, metricType string, metricName string) (models.Metrics, error)
	GetAll(ctx context.Context) (map[string]any, error)
	Ping(ctx context.Context) error
}

func New(cfg *config.ConfigServer) Storage {
	if cfg.FlagDBDSN != "" {
		storage := psql.New(cfg)
		return storage
	}

	memory := local.New()

	if cfg.FlagRestore {
		err := memory.LoadMetricsFromFile(cfg.FlagStorePath)
		if err != nil {
			logger.Log.Error("error loading metrics from the file", zap.Error(err))
		}
	}

	if cfg.FlagStorePath != "" {
		go memory.SaveAllMetricsToFile(cfg.FlagStoreInterval, cfg.FlagStorePath, cfg.FlagStorePath)
	}

	return memory
}