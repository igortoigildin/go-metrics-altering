package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	_ "net/http/pprof"
	"testing"

	config "github.com/igortoigildin/go-metrics-altering/config/server"
	"github.com/igortoigildin/go-metrics-altering/internal/models"
	"github.com/igortoigildin/go-metrics-altering/internal/server/api/mocks"
	local "github.com/igortoigildin/go-metrics-altering/internal/storage/inmemory"
	psql "github.com/igortoigildin/go-metrics-altering/internal/storage/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_updates(t *testing.T) {
	gaugeValue := float64(1.5)
	counterValue := int64(5)

	type metric struct {
		ID    string `json:"id"`
		MType string `json:"type"`
		Value any    `json:"value,omitempty"`
		Delta any    `json:"delta,omitempty"`
	}

	input := []metric{
		{
			ID:    "first_metric",
			MType: "gauge",
			Value: &gaugeValue,
		},
		{
			ID:    "second_metric",
			MType: "incorrect_type",
			Value: &gaugeValue,
		},
		{
			ID:    "third_metric",
			MType: "gauge",
			Value: "adfad",
		},
		{
			ID:    "4th_metric",
			MType: "gauge",
			Value: &gaugeValue,
		},
		{
			ID:    "5th_metric",
			MType: "counter",
			Delta: &counterValue,
		},
	}

	tests := []struct {
		name           string
		metric         []metric // udpated metrics which need to be saved
		respError      string
		mockError      error
		respStatusCode int
		inputIndex     []int
	}{
		{
			name:           "Success",
			inputIndex:     []int{0, 4},
			metric:         input,
			respStatusCode: 200,
		},
		{
			name:           "Unsupported metric type",
			inputIndex:     []int{1},
			metric:         input,
			respError:      "usupported request type",
			respStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:           "Storage error",
			inputIndex:     []int{3},
			metric:         input,
			respStatusCode: http.StatusInternalServerError,
			mockError:      errors.New("unexpected error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewStorage(t)

			if tt.respError == "" && tt.mockError == nil {
				repo.On("Update", mock.Anything, tt.metric[0].MType, tt.metric[0].ID, tt.metric[0].Value).Return(nil).Times(1)
				repo.On("Update", mock.Anything, tt.metric[4].MType, tt.metric[4].ID, tt.metric[4].Delta).Return(nil).Times(1)
			}

			if tt.mockError != nil {
				repo.On("Update", mock.Anything, tt.metric[3].MType, tt.metric[3].ID, tt.metric[3].Value).Return(tt.mockError).Times(1)
			}

			var metrics []metric
			for _, v := range tt.inputIndex {
				metrics = append(metrics, tt.metric[v])
			}
			js, _ := json.Marshal(metrics)

			handler := updates(repo)
			req, err := http.NewRequest(http.MethodPost, "/updates/", bytes.NewReader([]byte(js)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			require.Equal(t, tt.respStatusCode, rr.Code)
		})
	}
}

func Test_updateMetric(t *testing.T) {
	gaugeValue := float64(1.5)
	counterValue := int64(5)

	type metric struct {
		ID    string `json:"id"`
		MType string `json:"type"`
		Value any    `json:"value,omitempty"`
		Delta any    `json:"delta,omitempty"`
	}

	tests := []struct {
		name           string
		metric         metric // udpated metric which need to be saved
		respError      string
		mockError      error
		respStatusCode int
		inputIndex     int
		method         string
	}{
		{
			name: "Success_gauge",
			metric: metric{
				ID:    "gauge_correct",
				MType: "gauge",
				Value: &gaugeValue,
			},
			respStatusCode: 200,
			method:         http.MethodPost,
		},
		{
			name: "Success_counter",
			metric: metric{
				ID:    "counter_correct",
				MType: "counter",
				Delta: &counterValue,
			},
			respStatusCode: 200,
			method:         http.MethodPost,
		},
		{
			name: "Unsupported metric type",
			metric: metric{
				ID:    "third_metric",
				MType: "incorrect_type",
				Value: &gaugeValue,
			},
			respError:      "usupported request type",
			respStatusCode: http.StatusUnprocessableEntity,
			method:         http.MethodPost,
		},
		{
			name:       "Storage error",
			inputIndex: 3,
			metric: metric{
				ID:    "4th_metric",
				MType: "gauge",
				Value: &gaugeValue,
			},
			respStatusCode: http.StatusInternalServerError,
			mockError:      errors.New("unexpected error"),
			method:         http.MethodPost,
		},
		{
			name: "Incorrect method",
			metric: metric{
				ID:    "5th_metric",
				MType: "counter",
				Delta: &counterValue,
			},
			respError:      "Method not allowed",
			respStatusCode: http.StatusMethodNotAllowed,
			method:         http.MethodGet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewStorage(t)

			if tt.respError == "" && tt.mockError == nil {
				repo.On("Update", mock.Anything, tt.metric.MType, tt.metric.ID, tt.metric.Value).Return(nil).Maybe()
				repo.On("Update", mock.Anything, tt.metric.MType, tt.metric.ID, tt.metric.Delta).Return(nil).Maybe()
			}

			if tt.respError != "" {
				repo.On("Update", mock.Anything, tt.metric.MType, tt.metric.ID, tt.metric.Value).Return(tt.respError).Maybe()
				repo.On("Update", mock.Anything, tt.metric.MType, tt.metric.ID, tt.metric.Delta).Return(tt.respError).Maybe()
			}

			if tt.mockError != nil {
				repo.On("Update", mock.Anything, tt.metric.MType, tt.metric.ID, tt.metric.Value).Return(tt.mockError)
			}

			js, _ := json.Marshal(tt.metric)

			handler := updateMetric(repo)
			req, err := http.NewRequest(tt.method, "/update/", bytes.NewReader([]byte(js)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			require.Equal(t, tt.respStatusCode, rr.Code)
		})
	}
}

func Test_getMetric(t *testing.T) {
	gaugeValue := float64(10)
	countValue := int64(5)
	type req struct {
		ID    string `json:"id"`
		MType string `json:"type"`
	}

	tests := []struct {
		name           string
		req            req
		respError      string
		mockError      error
		respStatusCode int
		inputIndex     int
		response       models.Metrics
		method         string
	}{
		{
			name: "Success with gauge",
			req: req{
				ID:    "first_metric",
				MType: "gauge",
			},
			respStatusCode: 200,
			response: models.Metrics{
				ID:    "first_metric",
				MType: "gauge",
				Value: &gaugeValue,
			},
			method: http.MethodGet,
		},
		{
			name: "Unsupported metric type",
			req: req{
				ID:    "second_metric",
				MType: "incorrect_type",
			},
			respError:      "unsupported metric type",
			respStatusCode: http.StatusUnprocessableEntity,
			method:         http.MethodGet,
		},
		{
			name: "Success with counter",
			req: req{
				ID:    "third_metric",
				MType: "counter",
			},
			respStatusCode: http.StatusOK,
			response: models.Metrics{
				ID:    "third_metric",
				MType: "counter",
				Delta: &countValue,
			},
			method: http.MethodGet,
		},
		{
			name: "Storage error",
			req: req{
				ID:    "4th_metric",
				MType: "counter",
			},
			respStatusCode: http.StatusInternalServerError,
			mockError:      errors.New("unexpected error"),
			method:         http.MethodGet,
		},
		{
			name: "Incorrect method",
			req: req{
				ID:    "5th_metric",
				MType: "counter",
			},
			respStatusCode: http.StatusMethodNotAllowed,
			method:         http.MethodPut,
			respError:      "unsupported request type",
		},
		{
			name:           "Incorrect JSON",
			respStatusCode: http.StatusUnprocessableEntity,
			method:         http.MethodGet,
			respError:      "bad request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewStorage(t)

			if tt.respError == "" && tt.mockError == nil {
				repo.On("Get", mock.Anything, tt.req.MType, tt.req.ID).Return(tt.response, nil).Maybe()
				repo.On("Get", mock.Anything, tt.req.MType, tt.req.ID).Return(tt.response, nil).Maybe()
			}

			if tt.respError != "" {
				repo.On("Get", mock.Anything, tt.req.MType, tt.req.ID).Return(tt.respError, nil).Maybe()
				repo.On("Get", mock.Anything, tt.req.MType, tt.req.ID).Return(models.Metrics{}, tt.mockError).Maybe()
			}

			if tt.mockError != nil {
				repo.On("Get", mock.Anything, tt.req.MType, tt.req.ID).Return(models.Metrics{}, tt.mockError).Maybe()
			}

			js, _ := json.Marshal(tt.req)

			handler := getMetric(repo)
			req, err := http.NewRequest(tt.method, "/value/", bytes.NewReader([]byte(js)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			require.Equal(t, tt.respStatusCode, rr.Code)
		})
	}
}

func Test_ping(t *testing.T) {
	tests := []struct {
		name           string
		response       map[string]interface{}
		respError      string
		mockError      error
		respStatusCode int
		method         string
	}{
		{
			name:           "Success",
			respStatusCode: http.StatusOK,
			method:         http.MethodGet,
		},
		{
			name:           "Wrong method",
			respStatusCode: http.StatusMethodNotAllowed,
			method:         http.MethodPost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewStorage(t)

			if tt.respError == "" && tt.mockError == nil {
				repo.On("Ping", context.Background()).Return(nil).Maybe()
			}

			handler := ping(repo)

			req, err := http.NewRequest(tt.method, "/ping", bytes.NewReader([]byte(nil)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			require.Equal(t, tt.respStatusCode, rr.Code)
		})
	}
}

func Test_valuePathHandler(t *testing.T) {
	gaugeValue := float64(10)
	countValue := int64(5)
	type mod struct {
		ID    string `json:"id"`
		MType string `json:"type"`
	}

	tests := []struct {
		name           string
		mod            mod
		respError      string
		mockError      error
		respStatusCode int
		inputIndex     int
		response       models.Metrics
		method         string
	}{
		{
			name: "Success with gauge",
			mod: mod{
				ID:    "firstmetric",
				MType: "gauge",
			},
			respStatusCode: http.StatusOK,
			response: models.Metrics{
				Value: &gaugeValue,
			},
			method: http.MethodGet,
		},
		{
			name: "Success with counter",
			mod: mod{
				ID:    "PollCount",
				MType: "counter",
			},
			respStatusCode: http.StatusOK,
			response: models.Metrics{
				Delta: &countValue,
			},
			method: http.MethodGet,
		},
		{
			name: "Unsupported metric type",
			mod: mod{
				ID:    "unknown",
				MType: "wrong_type",
			},
			respStatusCode: http.StatusUnprocessableEntity,
			respError:      "unknown metric type",
			method:         http.MethodGet,
			response:       models.Metrics{},
		},
		{
			name: "Mock error",
			mod: mod{
				ID:    "unknown",
				MType: "wrong_type",
			},
			respStatusCode: http.StatusUnprocessableEntity,
			mockError:      errors.New("unexpected error"),
			method:         http.MethodGet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewStorage(t)

			if tt.respError == "" && tt.mockError == nil {
				repo.On("Get", mock.Anything, tt.mod.MType, tt.mod.ID).Return(tt.response, nil).Once()
			}

			if tt.mockError != nil {
				repo.On("Get", mock.Anything, tt.mod.MType, tt.mod.ID).Return(models.Metrics{}, tt.mockError).Maybe()
			}

			handler := valuePathHandler(repo)
			mux := &http.ServeMux{}
			mux.HandleFunc("/value", handler)

			req := httptest.NewRequest(http.MethodGet, "/value", nil)
			req.SetPathValue("metricType", tt.mod.MType)
			req.SetPathValue("metricName", tt.mod.ID)

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)
			resp := rr.Result()

			require.Equal(t, tt.respStatusCode, resp.StatusCode)
		})
	}
}

func Test_updatePathHandler(t *testing.T) {
	type mod struct {
		ID     string `json:"id"`
		MType  string `json:"type"`
		ValStr string
		Value  float64
		Delta  int64
	}

	tests := []struct {
		name           string
		mod            mod
		respError      string
		mockError      error
		respStatusCode int
		inputIndex     int
		response       models.Metrics
		method         string
	}{
		{
			name: "Success with gauge",
			mod: mod{
				ID:     "firstmetric",
				MType:  "gauge",
				ValStr: "45",
				Value:  float64(45),
			},
			respStatusCode: http.StatusOK,
			response:       models.Metrics{},
			method:         http.MethodGet,
		},
		{
			name: "Success with counter",
			mod: mod{
				ID:     "PollCount",
				MType:  "counter",
				ValStr: "1",
				Delta:  int64(1),
			},
			respStatusCode: http.StatusOK,
			response:       models.Metrics{},
			method:         http.MethodGet,
		},
		{
			name: "Unsupported metric type",
			mod: mod{
				ID:     "unknown",
				MType:  "wrong_type",
				ValStr: "1",
			},
			respStatusCode: http.StatusBadRequest,
			respError:      "unknown metric type",
			method:         http.MethodGet,
			response:       models.Metrics{},
		},
		{
			name: "Metric id not provided",
			mod: mod{
				ID:     "",
				MType:  "wrong_type",
				ValStr: "1",
			},
			respStatusCode: http.StatusBadRequest,
			respError:      "metricName not provided",
			method:         http.MethodGet,
			response:       models.Metrics{},
		},
		{
			name: "Metric with wrong value",
			mod: mod{
				ID:     "new_metric",
				MType:  "wrong_type",
				ValStr: "1.090",
			},
			respStatusCode: http.StatusBadRequest,
			respError:      "metricName not provided",
			method:         http.MethodGet,
			response:       models.Metrics{},
		},
		{
			name: "Mock error",
			mod: mod{
				ID:     "unknown",
				MType:  "wrong_type",
				ValStr: "1",
			},
			respStatusCode: http.StatusBadRequest,
			mockError:      errors.New("unexpected error"),
			method:         http.MethodGet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewStorage(t)

			if tt.respError == "" && tt.mockError == nil {
				repo.On("Update", mock.Anything, tt.mod.MType, tt.mod.ID, tt.mod.Delta).Return(nil).Maybe()
				repo.On("Update", mock.Anything, tt.mod.MType, tt.mod.ID, tt.mod.Value).Return(nil).Maybe()
			}

			if tt.mockError != nil {
				repo.On("Update", mock.Anything, tt.mod.MType, tt.mod.ID, mock.Anything).Return(tt.mockError).Maybe()
			}

			handler := updatePathHandler(repo)
			mux := &http.ServeMux{}
			mux.HandleFunc("/update", handler)

			req := httptest.NewRequest(http.MethodPost, "/update", nil)
			req.SetPathValue("metricType", tt.mod.MType)
			req.SetPathValue("metricName", tt.mod.ID)
			req.SetPathValue("metricValue", tt.mod.ValStr)

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)
			resp := rr.Result()

			require.Equal(t, tt.respStatusCode, resp.StatusCode)
		})
	}
}

func TestNew(t *testing.T) {
	cfg := config.ConfigServer{}
	s := New(&cfg)

	_, ok := s.(*local.LocalStorage)
	assert.True(t, ok)

	cfg.FlagDBDSN = "temp"
	p := New(&cfg)

	_, ok = p.(*psql.PGStorage)
	assert.True(t, ok)
}

func Test_getAllmetrics(t *testing.T) {
	tests := []struct {
		name           string
		respError      string
		mockError      error
		respStatusCode int
		inputIndex     int
		response       models.Metrics
		method         string
	}{
		{
			name:           "Success",
			respStatusCode: http.StatusOK,
		},
		{
			name:           "Mock error",
			mockError:      errors.New("error unknown"),
			respStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewStorage(t)

			if tt.respError == "" && tt.mockError == nil {
				repo.On("GetAll", mock.Anything).Return(nil, nil).Maybe()
			}

			if tt.mockError != nil {
				repo.On("GetAll", mock.Anything).Return(nil, tt.mockError).Maybe()
			}

			handler := getAllmetrics(repo)
			req, err := http.NewRequest(tt.method, "/", nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			require.Equal(t, tt.respStatusCode, rr.Code)
		})
	}
}
