package api

import (
	"context"
	"net/http"
	"text/template"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	config "github.com/igortoigildin/go-metrics-altering/config/server"
	"github.com/igortoigildin/go-metrics-altering/internal/storage"
	auth "github.com/igortoigildin/go-metrics-altering/pkg/middlewares/auth"
	compress "github.com/igortoigildin/go-metrics-altering/pkg/middlewares/compress"
	logging "github.com/igortoigildin/go-metrics-altering/pkg/middlewares/logging"
	timeout "github.com/igortoigildin/go-metrics-altering/pkg/middlewares/timeout"
	"github.com/igortoigildin/go-metrics-altering/templates"
)

var t *template.Template

func Router(ctx context.Context, cfg *config.ConfigServer, storage storage.Storage) chi.Router {
	t = templates.ParseTemplate()
	r := chi.NewRouter()

	r.Get("/value/{metricType}/{metricName}", logging.WithLogging(compress.GzipMiddleware(auth.Auth(http.HandlerFunc(valuePathHandler(storage)), cfg))))
	r.Post("/update/{metricType}/{metricName}/{metricValue}", logging.WithLogging(compress.GzipMiddleware(auth.Auth(http.HandlerFunc(updatePathHandler(storage)), cfg))))

	r.Get("/ping", timeout.Timeout(cfg.ContextTimout, logging.WithLogging(compress.GzipMiddleware(http.HandlerFunc(ping(storage))))))
	r.Get("/", timeout.Timeout(cfg.ContextTimout, logging.WithLogging(compress.GzipMiddleware(auth.Auth(http.HandlerFunc(getAllmetrics(storage)), cfg)))))
	r.Post("/updates/", timeout.Timeout(cfg.ContextTimout, logging.WithLogging(compress.GzipMiddleware(auth.Auth(http.HandlerFunc(updates(storage)), cfg)))))

	r.Post("/value/", timeout.Timeout(cfg.ContextTimout, logging.WithLogging(compress.GzipMiddleware(auth.Auth(http.HandlerFunc(getMetric(storage)), cfg)))))
	r.Post("/update/", timeout.Timeout(cfg.ContextTimout, logging.WithLogging(compress.GzipMiddleware(auth.Auth(http.HandlerFunc(updateMetric(storage)), cfg)))))

	r.Mount("/debug", middleware.Profiler())

	return r
}
