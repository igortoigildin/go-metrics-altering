// Package local storage implementation for inmemory Storage interface.
package local

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	config "github.com/igortoigildin/go-metrics-altering/config/agent"
	"github.com/igortoigildin/go-metrics-altering/internal/models"
	"github.com/igortoigildin/go-metrics-altering/pkg/logger"
	processmap "github.com/igortoigildin/go-metrics-altering/pkg/processMap"
	"go.uber.org/zap"
)

const pollCount = "PollCount"

type LocalStorage struct {
	rm       sync.RWMutex
	Gauge    map[string]float64
	Counter  map[string]int64
	strategy MetricAlgo
}

func New() *LocalStorage {
	return &LocalStorage{
		Counter: map[string]int64{pollCount: 0},
		Gauge: map[string]float64{},
	}
}

func (m *LocalStorage) setMetricAlgo(metricType string) {
	if metricType == config.CountType {
		count := counterRepo{
			Counter: m.Counter,
		}
		m.strategy = &count
	} else {
		gauge := gaugeRepo{
			Gauge: m.Gauge,
		}
		m.strategy = &gauge
	}
}

func (m *LocalStorage) Update(ctx context.Context, metricType string, metricName string, metricValue any) error {
	m.setMetricAlgo(metricType)

	m.rm.Lock()
	defer m.rm.Unlock()

	err := m.strategy.Update(metricType, metricName, metricValue)
	if err != nil {
		logger.Log.Error("error while updating metric", zap.Error(err))
		return err
	}
	return nil
}

func (m *LocalStorage) Get(ctx context.Context, metricType string, metricName string) (models.Metrics, error) {
	m.setMetricAlgo(metricType)

	m.rm.RLock()
	defer m.rm.RUnlock()

	metric, err := m.strategy.Get(metricType, metricName)
	if err != nil {
		logger.Log.Error("error while getting metric", zap.Error(err))
		return models.Metrics{}, err
	}
	return metric, nil
}

func (m *LocalStorage) GetAll(ctx context.Context) (map[string]any, error) {
	m.rm.Lock()
	defer m.rm.Unlock()

	res := processmap.ConvertToSingleMap(m.Gauge, m.Counter)

	return res, nil
}

// LoadMetricsFromFile loads metrics from the stated file.
func (m *LocalStorage) LoadMetricsFromFile(fname string) error {
	data, err := os.ReadFile(fname)
	if err != nil {
		return err
	}
	var metricSlice []models.Metrics
	if err := json.Unmarshal(data, &metricSlice); err != nil {
		return err
	}
	
	m.rm.Lock()
	defer m.rm.Unlock()

	for _, v := range metricSlice {
		if v.MType == "gauge" {
			m.Gauge[v.ID] = *v.Value
		} else if v.MType == "counter" {
			m.Counter[v.ID] = *v.Delta
		}
	}
	return nil
}

func (m *LocalStorage) Ping(ctx context.Context) error {
	if m.Gauge == nil {
		logger.Log.Info("gauge local storage not initialized")
		return errors.New("gauge local storage not initialized")
	}

	if m.Counter == nil {
		logger.Log.Info("counter local storage not initialized")
		return errors.New("counter local storage not initialized")
	}
	return nil
}

// SaveAllMetricsToFile periodically saves metrics from local storage to provided file.
func (m *LocalStorage) SaveAllMetricsToFile(FlagStoreInterval int, FlagStorePath string, fname string) error {
	//pauseDuration := time.Duration(FlagStoreInterval) * time.Second
	for {
		//time.Sleep(pauseDuration)
		metrics, _ := m.GetAll(context.Background())
		slice := []models.Metrics{}

		for key, v := range metrics {
			var metric models.Metrics
			if val, ok := v.(float64); ok {
				metric.ID = key
				metric.Value = &val
				metric.MType = config.GaugeType
				slice = append(slice, metric)
				continue
			}
			if val, ok := v.(int64); ok {
				metric.ID = key
				metric.Delta = &val
				metric.MType = config.CountType
				slice = append(slice, metric)
			}
		}

		data, err := json.MarshalIndent(slice, "", "  ")
		if err != nil {
			logger.Log.Info("marshalling error", zap.Error(err))
			return err
		}

		err = os.WriteFile(fname, data, 0606)
		if err != nil {
			logger.Log.Info("error saving metrics to the file", zap.Error(err))
			return err
		}
	}
}
