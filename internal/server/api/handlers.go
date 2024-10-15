package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	_ "net/http/pprof" // подключаем пакет pprof

	"github.com/go-chi/chi"
	config "github.com/igortoigildin/go-metrics-altering/config/server"
	"github.com/igortoigildin/go-metrics-altering/internal/models"
	"github.com/igortoigildin/go-metrics-altering/internal/storage"
	"github.com/igortoigildin/go-metrics-altering/pkg/logger"
	processjson "github.com/igortoigildin/go-metrics-altering/pkg/processJSON"
	"go.uber.org/zap"
)

//go:generate go run github.com/vektra/mockery/v2@v2.45.0 --name=Storage


func ping(Storage storage.Storage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if r.Method != http.MethodGet {
			logger.Log.Info("got request with bad method", zap.String("method", r.Method))
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		err := Storage.Ping(ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

func updates(Storage storage.Storage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if r.Method != http.MethodPost {
			logger.Log.Info("got request with bad method", zap.String("method", r.Method))
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		metrics := make([]models.Metrics, 0)
		err := processjson.ReadJSON(r, &metrics)
		if err != nil {
			logger.Log.Info("cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// iterating through []Metrics and adding it to db one by one
		for _, metric := range metrics {
			if metric.MType != config.GaugeType && metric.MType != config.CountType {
				logger.Log.Info("usupported request type", zap.String("type", metric.MType))
				w.WriteHeader(http.StatusUnprocessableEntity)
				return
			}
			switch metric.MType {
			case config.GaugeType:
				err := Storage.Update(ctx, metric.MType, metric.ID, metric.Value)
				if err != nil {
					logger.Log.Info("error while updating value", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			case config.CountType:
				err := Storage.Update(ctx, metric.MType, metric.ID, metric.Delta)
				if err != nil {
					logger.Log.Info("error while updating value", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
	})
}

func updateMetric(Storage storage.Storage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if r.Method != http.MethodPost {
			logger.Log.Info("got request with bad method", zap.String("method", r.Method))
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req models.Metrics
		err := processjson.ReadJSON(r, &req)
		if err != nil {
			logger.Log.Info("cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if req.MType != config.GaugeType && req.MType != config.CountType {
			logger.Log.Info("usupported request type", zap.String("type", req.MType))
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		switch req.MType {
		case config.GaugeType:
			err := Storage.Update(ctx, req.MType, req.ID, req.Value)
			if err != nil {
				logger.Log.Info("error while updating value", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		case config.CountType:
			err := Storage.Update(ctx, req.MType, req.ID, req.Delta)
			if err != nil {
				logger.Log.Info("error while updating value", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		resp := models.Metrics{
			ID:    req.ID,
			MType: req.MType,
			Value: req.Value,
			Delta: req.Delta,
		}
		err = processjson.WriteJSON(w, http.StatusOK, resp, nil)
		if err != nil {
			logger.Log.Info("error encoding response", zap.Error(err))
			return
		}
	})
}

func getAllmetrics(Storage storage.Storage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Add("Content-Encoding", "gzip")
		metrics, err := Storage.GetAll(r.Context())
		if err != nil {
			logger.Log.Info("error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := t.Execute(w, metrics); err != nil {
			logger.Log.Info("error executing template", zap.Error(err))
		}
	})
}

func getMetric(Storage storage.Storage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if r.Method != http.MethodPost {
			logger.Log.Info("got request with bad method", zap.String("method", r.Method))
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req models.Metrics
		err := processjson.ReadJSON(r, &req)
		if err != nil {
			logger.Log.Info("cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if req.MType != config.GaugeType && req.MType != config.CountType {
			logger.Log.Info("usupported request type", zap.String("type", req.MType))
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		resp := models.Metrics{
			ID:    req.ID,
			MType: req.MType,
		}

		switch req.MType {
		case config.GaugeType:
			res, err := Storage.Get(ctx, req.MType, req.ID)
			if err != nil {
				switch {
				case errors.Is(err, sql.ErrNoRows):
					logger.Log.Info("error while obtaining metric", zap.Error(err))
					w.WriteHeader(http.StatusNotFound)
					return
				default:
					logger.Log.Info("error while obtaining metric", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
			resp.Value = res.Value
		case config.CountType:
			res, err := Storage.Get(ctx, req.MType, req.ID)
			if err != nil {
				switch {
				case errors.Is(err, sql.ErrNoRows):
					logger.Log.Info("error while obtaining metric", zap.Error(err))
					w.WriteHeader(http.StatusNotFound)
					return
				default:
					logger.Log.Info("error while obtaining metric", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
			resp.Delta = res.Delta
		}

		w.Header().Add("Content-Encoding", "gzip")
		err = processjson.WriteJSON(w, http.StatusOK, resp, nil)
		if err != nil {
			logger.Log.Info("error encoding response", zap.Error(err))
			return
		}
	})
}

func updatePathHandler(LocalStorage storage.Storage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		if metricName == "" {
			logger.Log.Info("metricName not provided")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		metricValue := chi.URLParam(r, "metricValue")

		switch metricType {
		case config.GaugeType:
			metricValueConverted, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {

				logger.Log.Info("error parsing metric value to float", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			LocalStorage.Update(context.TODO(), config.GaugeType, metricName, metricValueConverted)
		case config.CountType:
			metricValueConverted, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				logger.Log.Info("error parsing metric value to int", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			LocalStorage.Update(context.TODO(), config.CountType, metricName, metricValueConverted)
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	})
}

func valuePathHandler(LocalStorage storage.Storage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		switch metricType {
		case config.GaugeType:
			metric, err := LocalStorage.Get(context.TODO(), config.GaugeType, metricName)
			if err != nil {
				logger.Log.Info("error while loading metric", zap.String("metric name", metricName))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write([]byte(strconv.FormatFloat(*metric.Value, 'f', -1, 64)))
		case metricType:
			metric, err := LocalStorage.Get(context.TODO(), config.CountType, config.PollCount)
			w.Write([]byte(strconv.FormatInt(*metric.Delta, 10)))
			if err != nil {
				logger.Log.Info("error while loading metric", zap.String("metric name", metricName))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		default:
			logger.Log.Info("usupported request type", zap.String("type", metricType))
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
	})
}
