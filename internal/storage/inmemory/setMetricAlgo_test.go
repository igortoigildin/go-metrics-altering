package local

import (
	"sync"
	"testing"

	"github.com/igortoigildin/go-metrics-altering/internal/models"
	"github.com/stretchr/testify/assert"
)

func Test_counterRepo_Update(t *testing.T) {
	type fields struct {
		Counter map[string]int64
		rm      sync.Mutex
	}
	type args struct {
		metricType  string
		metricName  string
		metricValue any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		wantVal int64
	}{
		{
			name: "#1",
			args: args{
				metricType:  "count",
				metricName:  "new",
				metricValue: int64(1),
			},
			wantErr: false,
			wantVal: int64(2),
		},
		{
			name: "#2",
			args: args{
				metricType:  "count",
				metricName:  "new",
				metricValue: int64(10),
			},
			wantErr: false,
			wantVal: int64(11),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &counterRepo{
				Counter: map[string]int64{"new": int64(1)},
			}
			if err := c.Update(tt.args.metricType, tt.args.metricName, tt.args.metricValue); (err != nil) != tt.wantErr {
				t.Errorf("counterRepo.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.wantVal, c.Counter[tt.args.metricName])
		})
	}
}

func Test_gaugeRepo_Update(t *testing.T) {
	type fields struct {
		Gauge map[string]float64
		rm    sync.Mutex
	}
	type args struct {
		metricType  string
		metricName  string
		metricValue any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		wantVal float64
	}{
		{
			name: "#1",
			args: args{
				metricType:  "count",
				metricName:  "new",
				metricValue: float64(2),
			},
			wantErr: false,
			wantVal: float64(2),
		},
		{
			name: "#2",
			args: args{
				metricType:  "count",
				metricName:  "new",
				metricValue: float64(11),
			},
			wantErr: false,
			wantVal: float64(11),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &gaugeRepo{
				Gauge: map[string]float64{"new": float64(1)},
			}
			if err := g.Update(tt.args.metricType, tt.args.metricName, tt.args.metricValue); (err != nil) != tt.wantErr {
				t.Errorf("gaugeRepo.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.wantVal, g.Gauge[tt.args.metricName])
		})
	}
}

func Test_counterRepo_Get(t *testing.T) {
	type fields struct {
		Counter map[string]int64
		rm      sync.Mutex
	}
	type args struct {
		metricType string
		metricName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    models.Metrics
		wantErr bool
		wantVal int64
	}{
		{
			name: "#1",
			args: args{
				metricType: "count",
				metricName: "first",
			},
			wantErr: false,
			wantVal: int64(2),
		},
		{
			name: "#2",
			args: args{
				metricType: "count",
				metricName: "second",
			},
			wantErr: false,
			wantVal: int64(11),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &counterRepo{
				Counter: map[string]int64{"first": int64(2), "second": int64(11)},
			}
			got, err := c.Get(tt.args.metricType, tt.args.metricName)
			if (err != nil) != tt.wantErr {
				t.Errorf("counterRepo.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.wantVal, *got.Delta)
		})
	}
}

func Test_gaugeRepo_Get(t *testing.T) {
	type fields struct {
		Gauge map[string]float64
		rm    sync.Mutex
	}
	type args struct {
		metricType string
		metricName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    models.Metrics
		wantErr bool
		wantVal float64
	}{
		{
			name: "#1",
			args: args{
				metricType: "gauge",
				metricName: "first",
			},
			wantErr: false,
			wantVal: float64(2),
		},
		{
			name: "#2",
			args: args{
				metricType: "gauge",
				metricName: "second",
			},
			wantErr: false,
			wantVal: float64(11),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &gaugeRepo{
				Gauge: map[string]float64{"first": float64(2), "second": float64(11)},
			}
			got, err := g.Get(tt.args.metricType, tt.args.metricName)
			if (err != nil) != tt.wantErr {
				t.Errorf("gaugeRepo.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.wantVal, *got.Value)
		})
	}
}
