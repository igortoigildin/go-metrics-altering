// Package agent accumulates, runtime metrics
// and sends it to predefined server every poll interval.

package agent

import (
	"time"

	config "github.com/igortoigildin/go-metrics-altering/config/agent"
	"github.com/igortoigildin/go-metrics-altering/internal/models"
	"github.com/igortoigildin/go-metrics-altering/pkg/logger"
	"go.uber.org/zap"
)

// SendMetrics reads metrics from metricsChan and sends it server address as defined by agent config.
func SendMetrics(metricsChan <-chan models.Metrics, cfg *config.ConfigAgent) {
	for {
		time.Sleep(cfg.PauseDuration)

		for metric := range metricsChan {

			switch metric.MType {
			case config.CountType:
				err := sendURLCounter(cfg, int(*metric.Delta))
				if err != nil {
					logger.Log.Error("unexpected sending url counter metric error:", zap.Error(err))
				}
				err = SendJSONCounter(int(*metric.Delta), cfg)
				if err != nil {
					logger.Log.Error("unexpected sending json counter metric error:", zap.Error(err))
				}
			case config.GaugeType:
				err := SendURLGauge(cfg, *metric.Value, metric.ID)
				if err != nil {
					logger.Log.Error("unexpected sending url gauge metric error:", zap.Error(err))
				}
				err = SendJSONGauge(metric.ID, cfg, *metric.Value)
				if err != nil {
					logger.Log.Error("unexpected sending json gauge metric error:", zap.Error(err))
				}
			}
		}
	}
}
