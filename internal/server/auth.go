package server

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	config "github.com/igortoigildin/go-metrics-altering/config/server"
	"github.com/igortoigildin/go-metrics-altering/internal/logger"
)

func auth(next http.HandlerFunc, cfg *config.ConfigServer) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        key := []byte(cfg.FlagHashKey)
        if len(key) > 0 && r.Header.Get("HashSHA256") != "" {
            if r.Body != nil {
                messageMAC := []byte(r.Header.Get("HashSHA256"))

                bodyBytes, err := io.ReadAll(r.Body)
                if err != nil {
                    logger.Log.Info("cannot reads from reqeust body")
                    w.WriteHeader(http.StatusInternalServerError)
                    return
                }
                r.Body.Close()
                r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
                
                if !ValidMAC(bodyBytes, messageMAC, key) {
                    logger.Log.Info("wrong hash")
                    w.WriteHeader(http.StatusBadRequest)
                    return
                }
            }
        }       
        next(w, r)
    })
}


// ValidMAC reports whether messageMAC is a valid HMAC tag for message.
func ValidMAC(message, messageMAC, key []byte) bool {
    mac := hmac.New(sha256.New, key)
    mac.Write([]byte(message))
    expectedMAC := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal(messageMAC, []byte(expectedMAC))
}

