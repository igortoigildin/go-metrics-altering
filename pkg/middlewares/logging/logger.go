// Package logging provides middleware for logging http requests and responses.

package logging

import (
	"net/http"
	"time"

	"github.com/igortoigildin/go-metrics-altering/pkg/logger"

	"go.uber.org/zap"
)

type (
	// info struct in regards to reply
	responseData struct {
		status int
		size   int
	}
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// WithLogging adds code to regester info regarding request and returns new http.Handler
func WithLogging(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, //using original http.ResponseWriter
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r) //using updated http.ResponseWriter
		duration := time.Since(start)
		logger.Log.Info("got incoming HTTP request",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Int("status", responseData.status),
			zap.String("duration", duration.String()),
			zap.Int("size", responseData.size),
		)
	})
}
