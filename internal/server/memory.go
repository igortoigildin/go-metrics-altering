package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"time"

	config "github.com/igortoigildin/go-metrics-altering/config/server"
	"github.com/igortoigildin/go-metrics-altering/internal/logger"
	"github.com/igortoigildin/go-metrics-altering/internal/models"
	"go.uber.org/zap"
)

type DB struct {
	conn *sql.DB
}

func NewDB(conn *sql.DB) *DB {
	return &DB{
		conn: conn,
	}
}

func InitRepo(c context.Context, cfg *config.ConfigServer) *DB {
	dbDSN := cfg.FlagDBDSN
	conn, err := sql.Open("pgx", dbDSN)
	if err != nil {
		logger.Log.Fatal("error while connecting to DB", zap.Error(err))
	}
	return &DB{
		conn: conn,
	}
}

// iterate through memStorage
func (m MemStorage) ConvertToSlice() []models.Metrics {
	metricSlice := make([]models.Metrics, 33)
	var model models.Metrics
	for i, v := range m.Gauge {
		model.ID = i
		model.MType = config.GaugeType

		model.Value = &v
		metricSlice = append(metricSlice, model)
	}
	for j, u := range m.Counter {
		model.ID = j
		model.MType = config.CountType
		model.Delta = &u
		metricSlice = append(metricSlice, model)
	}
	return metricSlice
}

// save slice with metrics to the file
func Save(fname string, metricSlice []models.Metrics) error {
	data, err := json.MarshalIndent(metricSlice, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(fname, data, 0606)
}

// load metrics from local file
func (m *MemStorage) Load(fname string) error {
	data, err := os.ReadFile(fname)
	if err != nil {
		return err
	}
	var metricSlice []models.Metrics
	if err := json.Unmarshal(data, &metricSlice); err != nil {
		return err
	}
	for _, v := range metricSlice {
		if v.MType == "gauge" {
			m.Gauge[v.ID] = *v.Value
		} else if v.MType == "counter" {
			m.Counter[v.ID] = *v.Delta
		}
	}
	return nil
}

// save metrics from memStorage to the file every StoreInterval
func (m *MemStorage) saveMetrics(cfg *config.ConfigServer) {
	pauseDuration := time.Duration(cfg.FlagStoreInterval) * time.Second
	for {
		time.Sleep(pauseDuration)
		metricSlice := m.ConvertToSlice()
		if err := Save(cfg.FlagStorePath, metricSlice); err != nil {
			logger.Log.Info("error saving metrics to the file", zap.Error(err))
		}
	}
}