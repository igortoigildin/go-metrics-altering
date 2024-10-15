package local

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	config "github.com/igortoigildin/go-metrics-altering/config/agent"
	"github.com/igortoigildin/go-metrics-altering/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestLocalStorage_SetStrategy(t *testing.T) {
	type fields struct {
		Gauge    map[string]float64
		Counter  map[string]int64
		strategy MetricAlgo
	}

	tests := []struct {
		name       string
		fields     fields
		metricType string
	}{
		{
			name:       "Success",
			fields:     fields{},
			metricType: "counter",
		},
		{
			name:       "Success",
			fields:     fields{},
			metricType: "gauge",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &LocalStorage{
				Gauge:    tt.fields.Gauge,
				Counter:  tt.fields.Counter,
				strategy: tt.fields.strategy,
			}
			m.setMetricAlgo(tt.metricType)

			assert.True(t, m.strategy != nil)
		})
	}
}

func TestInitLocalStorage(t *testing.T) {
	tests := []struct {
		name      string
		wantError bool
	}{
		{
			name:      "Success",
			wantError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := New()
			assert.True(t, a.Counter != nil)
			assert.True(t, a.Gauge != nil)
		})
	}
}

func TestLocalStorage_Update(t *testing.T) {
	m := New()
	g, c := float64(45), int64(50)

	type args struct {
		ctx         context.Context
		metricType  string
		metricName  string
		metricValue any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Success_gauge",
			args: args{
				ctx:         context.TODO(),
				metricType:  "gauge",
				metricName:  "temp_metric1",
				metricValue: &g,
			},
			wantErr: false,
		},
		{
			name: "Success_counter",
			args: args{
				ctx:         context.TODO(),
				metricType:  "counter",
				metricName:  "temp_metric2",
				metricValue: &c,
			},
			wantErr: false,
		},
		{
			name: "Fail_counter",
			args: args{
				ctx:         context.TODO(),
				metricType:  "unknown_type",
				metricName:  "temp_metric2",
				metricValue: &c,
			},
			wantErr: true,
		},
		{
			name: "Fail_gauge",
			args: args{
				ctx:         context.TODO(),
				metricType:  "unknown_gauge",
				metricName:  "temp_metric1",
				metricValue: &g,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		_ = m.Update(tt.args.ctx, tt.args.metricType, tt.args.metricName, tt.args.metricValue)

		switch tt.args.metricType {
		case "gauge":
			res, _ := tt.args.metricValue.(float64)


			assert.Equal(t, res, m.Gauge[tt.args.metricName])
		case "counter":
			res, _ := tt.args.metricValue.(int64)

			assert.Equal(t, res, m.Counter[tt.args.metricName])
		}

	}
}

func TestLocalStorage_Get(t *testing.T) {
	m := New()
	m.Counter["count_metric"] = int64(25)
	m.Gauge["gauge_metric"] = float64(50)
	m1 := m.Counter["count_metric"]
	m2 := m.Gauge["gauge_metric"]

	type args struct {
		ctx        context.Context
		metricType string
		metricName string
	}
	tests := []struct {
		name      string
		args      args
		want      models.Metrics
		wantErr   bool
		wantValue any
	}{
		{
			name: "Gauge_success",
			args: args{
				ctx:        context.TODO(),
				metricType: "gauge",
				metricName: "gauge_metric",
			},
			wantErr:   false,
			wantValue: &m2,
		},
		{
			name: "Gauge_success",
			args: args{
				ctx:        context.TODO(),
				metricType: "count",
				metricName: "count_metric",
			},
			wantErr:   false,
			wantValue: &m1,
		},
		{
			name: "Gauge_Fail",
			args: args{
				ctx:        context.TODO(),
				metricType: "count",
				metricName: "unknown_metric",
			},
			wantErr:   true,
			wantValue: nil,
		},
		{
			name: "Gauge_Counter",
			args: args{
				ctx:        context.TODO(),
				metricType: "count",
				metricName: "unknown_metric",
			},
			wantErr:   true,
			wantValue: nil,
		},
	}
	for _, tt := range tests {
		res, _ := m.Get(tt.args.ctx, tt.args.metricType, tt.args.metricName)

		switch tt.args.metricType {
		case "gauge":
			v, _ := tt.wantValue.(*float64)
			val := *v

			assert.Equal(t, val, *res.Value)
		case "counter":
			v, _ := tt.wantValue.(*int64)
			val := *v

			assert.Equal(t, val, *res.Delta)
		}
	}
}

func TestLocalStorage_GetAll(t *testing.T) {
	m := New()
	m.Counter["count_metric"] = int64(25)
	m.Gauge["gauge_metric"] = float64(50)

	res := make(map[string]any)
	res["count_metric"] = int64(25)
	res["gauge_metric"] = float64(50)
	res[pollCount] = 0

	output, _ := m.GetAll(context.TODO())

	assert.Equal(t, res["count_metric"], output["count_metric"])
	assert.Equal(t, res["gauge_metric"], output["gauge_metric"])
	assert.Equal(t, len(res), len(output))
}

func TestLocalStorage_Ping(t *testing.T) {
	m := New()
	err := m.Ping(context.TODO())
	assert.NoError(t, err)

	l := LocalStorage{}
	err = l.Ping(context.TODO())
	assert.Error(t, err)
}

func TestLocalStorage_LoadMetricsFromFile(t *testing.T) {
	// Input data preparation: initializing storage with data.
	m := New()
	m.Counter["count_metric"] = int64(25)
	m.Gauge["gauge_metric"] = float64(50)
	metrics, _ := m.GetAll(context.Background())
	// Input data preparation: filling slice with data from storage.
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
	data, _ := json.MarshalIndent(slice, "", "  ")
	// Input data preparation: writing data to file.
	_ = os.WriteFile("temp", data, 0606)

	l := New()

	_ = l.LoadMetricsFromFile("temp")
	assert.Equal(t, m.Counter["count_metric"], l.Counter["count_metric"])
	assert.Equal(t, m.Counter["gauge_metric"], l.Counter["gauge_metric"])
}

func TestLocalStorage_SaveAllMetricsToFile(t *testing.T) {
	var fileName = "temp"
	m := New()
	m.Counter["count_metric"] = int64(25)
	m.Gauge["gauge_metric"] = float64(50)

	go m.SaveAllMetricsToFile(0, ".", fileName)

	var data []byte
	for data == nil {
		data, _ = os.ReadFile(fileName)
	}

	l := New()
	_ = l.LoadMetricsFromFile(fileName)
	assert.Equal(t, m.Counter["count_metric"], l.Counter["count_metric"])
	assert.Equal(t, m.Counter["gauge_metric"], l.Counter["gauge_metric"])
	_ = os.Remove(fileName)
}
