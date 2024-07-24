package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"text/template"

	"github.com/go-chi/chi"
	config "github.com/igortoigildin/go-metrics-altering/config/server"
	"github.com/igortoigildin/go-metrics-altering/internal/logger"
	"github.com/igortoigildin/go-metrics-altering/internal/models"
	"github.com/igortoigildin/go-metrics-altering/templates"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

var t *template.Template

func MetricRouter(cfg *config.ConfigServer, ctx context.Context) chi.Router {
    var m = InitStorage()
    if cfg.FlagRestore {
        err := m.Load(cfg.FlagStorePath)  // if flag restore is true - load metrics from the local file
        if err != nil {
            logger.Log.Info("error loading metrics from the file", zap.Error(err))
        }
    }
    // start goroutine to save metrics in file
    go m.saveMetrics(cfg)
    // parse template
    t = templates.ParseTemplate()
    r := chi.NewRouter()
    DB := InitRepo(ctx, cfg)
    r.Get("/ping", WithLogging(gzipMiddleware(http.HandlerFunc(DB.Ping))))
    // v1
    r.Get("/value/{metricType}/{metricName}", WithLogging(gzipMiddleware(auth(http.HandlerFunc(m.ValueHandle), cfg))))
    r.Get("/", WithLogging(gzipMiddleware(http.HandlerFunc(auth(m.InformationHandle, cfg)))))

    r.Route("/", func(r chi.Router) {
        // v1
        r.Post("/update/{metricType}/{metricName}/{metricValue}", WithLogging(gzipMiddleware(auth(http.HandlerFunc(m.UpdateHandle), cfg))))
        // v2
        r.Post("/update/", WithLogging(gzipMiddleware(auth(http.HandlerFunc(m.UpdateHandler), cfg))))
        r.Post("/value/", WithLogging(gzipMiddleware(auth(http.HandlerFunc(m.ValueHandler), cfg))))
    })
    return r
}

func (rep *DB) Ping(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    if err := rep.conn.PingContext(ctx); err != nil {
        logger.Log.Info("error", zap.Error(err))
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    rep.conn.Close()
    w.WriteHeader(http.StatusOK)
}

// v2 with requst body
func (m *MemStorage) UpdateHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        logger.Log.Debug("got request with bad method", zap.String("method", r.Method))
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    logger.Log.Debug("decoding request")
    var req models.Metrics
    dec := json.NewDecoder(r.Body)
    if err := dec.Decode(&req); err != nil {
        logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    if req.MType != config.GaugeType && req.MType != config.CountType {
        logger.Log.Debug("usupported request type", zap.String("type", req.MType))
        w.WriteHeader(http.StatusUnprocessableEntity)
        return
    }
    switch req.MType {
    case config.GaugeType:
        m.UpdateGaugeMetric(req.ID, *req.Value)
    case config.CountType:
        m.Counter[config.PollCount] += *req.Delta
    }

    var resp models.Metrics
    delta := m.Counter[config.PollCount]
    if delta == 0 {
        resp = models.Metrics{
            ID:    req.ID,
            MType: req.MType,
            Value: req.Value,
        }
    } else {
        resp = models.Metrics{
            ID:    req.ID,
            MType: req.MType,
            Value: req.Value,
            Delta: &delta,
        }
    }
    enc := json.NewEncoder(w)
    if err := enc.Encode(resp); err != nil {
        logger.Log.Debug("error encoding response", zap.Error(err))
        return
    }
    w.Header().Set("Content-Type", "application/json")
    logger.Log.Debug("sending HTTP 200 response")
}

// v2 with request body
func (m *MemStorage) ValueHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        logger.Log.Debug("got request with bad method", zap.String("method", r.Method))
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    logger.Log.Debug("decoding request")
    var req models.Metrics
    dec := json.NewDecoder(r.Body)
    if err := dec.Decode(&req); err != nil {
        logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    resp := models.Metrics{
        ID:    req.ID,
        MType: req.MType,
    }
    switch req.MType {
    case config.GaugeType:
        if !m.CheckIfGaugeMetricPresent(req.ID) {
            logger.Log.Debug("usupported metric name", zap.String("name", req.ID))
            w.WriteHeader(http.StatusNotFound)
            return
        }
        value := m.GetGaugeMetricFromMemory(req.ID)
        resp.Value = &value
    case config.CountType:
        delta := m.Counter[config.PollCount]
        resp.Delta = &delta
    default:
        logger.Log.Debug("usupported request type", zap.String("type", req.MType))
        w.WriteHeader(http.StatusUnprocessableEntity)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.Header().Add("Content-Encoding", "gzip")
    enc := json.NewEncoder(w)
    if err := enc.Encode(resp); err != nil {
        logger.Log.Debug("error encoding response", zap.Error(err))
        return
    }
    logger.Log.Debug("sending HTTP 200 response")
}

// v1
func (m *MemStorage) UpdateHandle(rw http.ResponseWriter, r *http.Request) {
    metricType := chi.URLParam(r, "metricType")
    metricName := chi.URLParam(r, "metricName")
    if metricName == "" {
        rw.WriteHeader(http.StatusNotFound)
        return
    }
    metricValue := chi.URLParam(r, "metricValue")
    switch metricType {
    case config.GaugeType:
        metricValueConverted, err := strconv.ParseFloat(metricValue, 64)
        if err != nil {
            logger.Log.Debug("error parsing metric value to float", zap.Error(err))
            rw.WriteHeader(http.StatusBadRequest)
            return
        }
        m.UpdateGaugeMetric(metricName, metricValueConverted)
    case config.CountType:
        metricValueConverted, err := strconv.ParseInt(metricValue, 10, 64)
        if err != nil {
            logger.Log.Debug("error parsing metric value to int", zap.Error(err))
            rw.WriteHeader(http.StatusBadRequest)
            return
        }
        m.UpdateCounterMetric(metricName, metricValueConverted)
    default:
        rw.WriteHeader(http.StatusBadRequest)
        return
    }
    rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
}

// v1
func (m *MemStorage) ValueHandle(rw http.ResponseWriter, r *http.Request) {
    metricType := chi.URLParam(r, "metricType")
    metricName := chi.URLParam(r, "metricName")
    switch metricType {
    case config.GaugeType:
        if !m.CheckIfGaugeMetricPresent(metricName) {
            rw.WriteHeader(http.StatusNotFound)
            return
        }
        _, err := rw.Write([]byte(strconv.FormatFloat(m.GetGaugeMetricFromMemory(metricName), 'f', -1, 64)))
        if err != nil {
            logger.Log.Debug("error occured while sending response", zap.Error(err))
        }
    case config.CountType:
        if !m.CheckIfCountMetricPresent(metricName) {
            rw.WriteHeader(http.StatusNotFound)
            return
        }
        _, err := rw.Write([]byte(strconv.FormatInt(m.GetCountMetricFromMemory(metricName), 10)))
        if err != nil {
            logger.Log.Debug("error occured while sending response", zap.Error(err))
        }
    default:
        rw.WriteHeader(http.StatusBadRequest)
        return
    }
    rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
    rw.WriteHeader(http.StatusOK)
}

func (m *MemStorage) InformationHandle(rw http.ResponseWriter, r *http.Request) {
    rw.Header().Set("Content-Type", "text/html; charset=utf-8")
    rw.Header().Add("Content-Encoding", "gzip")
    if err := t.Execute(rw, ConvertToSingleMap(m.Gauge, m.Counter)); err != nil {
        logger.Log.Debug("error executing template", zap.Error(err))
    }
}

