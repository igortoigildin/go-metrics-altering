package agent

import (
	"compress/gzip"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	config "github.com/igortoigildin/go-metrics-altering/config/agent"
	"github.com/igortoigildin/go-metrics-altering/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestSendJSONGauge(t *testing.T) {
	type args struct {
		metricName string
		cfg        *config.ConfigAgent
		value      float64
	}
	v := float64(1)
	successBody := models.Metrics{
		ID: "Alloc",
		MType: "gauge",
		Value: &v,
	}
	successResponse := `"id":"Alloc","type":"gauge","value":1`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// r.body is encoded in gzip format by default
		var metric models.Metrics
		reader, err := gzip.NewReader(r.Body)
		if err != nil {
			log.Println(err)
		}
		defer reader.Close()
		err = json.NewDecoder(reader).Decode(&metric)
		if err != nil {
			log.Println(err)
		}
		assert.Equal(t, "/update/", r.URL.String())
		assert.Equal(t, metric, successBody)
		w.Write([]byte(successResponse))
	}))
	defer server.Close()
	cfg := config.ConfigAgent{
		FlagRunAddr: "localhost:8080",
		URL: server.URL,
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Success",
			args: args{
				metricName: "Alloc",
				cfg: &cfg,
				value: 1.00,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SendJSONGauge(tt.args.metricName, tt.args.cfg, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("SendJSONGauge() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}


func TestSendURLGauge(t *testing.T) {
	type args struct {
		cfg        *config.ConfigAgent
		value      float64
		metricName string
	}
	successResponse := `"id":"Alloc","type":"gauge","value":1`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/update/gauge/Alloc/1.000000", r.URL.String())
		w.Write([]byte(successResponse))
	}))
	defer server.Close()
	cfg := config.ConfigAgent{
		FlagRunAddr: "localhost:8080",
		URL: server.URL,
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Success",
			args: args{
				metricName: "Alloc",
				cfg: &cfg,
				value: 1.00,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SendURLGauge(tt.args.cfg, tt.args.value, tt.args.metricName); (err != nil) != tt.wantErr {
				t.Errorf("SendURLGauge() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}