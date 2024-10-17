// Package logging provides middleware for logging http requests and responses.

package logging

import (
	"bufio"
	"bytes"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/igortoigildin/go-metrics-altering/pkg/logger"
	"github.com/stretchr/testify/assert"
)

type testLoggerWriter struct {
	*httptest.ResponseRecorder
}

func (cw testLoggerWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, nil
}

func TestRequestLoggerReadFrom(t *testing.T) {
	data := []byte("file data")
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "file", time.Time{}, bytes.NewReader(data))
	})

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler := WithLogging(testHandler)
	handler.ServeHTTP(w, r)

	assert.Equal(t, data, w.Body.Bytes())
}

func TestWithLogging(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	r := httptest.NewRequest("GET", "/", nil)
	w := testLoggerWriter{
		ResponseRecorder: httptest.NewRecorder(),
	}

	if err := logger.Initialize("info"); err != nil {
		log.Println("error while initializing logger", err)
		return
	}

	handler := WithLogging(testHandler)

	handler.ServeHTTP(w, r)
}
