package models

import config "github.com/igortoigildin/go-metrics-altering/config/agent"

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

// Constructs counter metric model
func CounterConstructor(delta int64) Metrics {
	return Metrics{
		ID:    config.PollCount,
		MType: config.CountType,
		Delta: &delta,
	}
}

// Constructs gauge metric model
func GaugeConstructor(value float64, name string) Metrics {
	return Metrics{
		ID:    name,
		MType: config.GaugeType,
		Value: &value,
	}
}
