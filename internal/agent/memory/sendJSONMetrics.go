package memory

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	config "github.com/igortoigildin/go-metrics-altering/config/agent"
	"github.com/igortoigildin/go-metrics-altering/internal/logger"
	"github.com/igortoigildin/go-metrics-altering/internal/models"
	"go.uber.org/zap"
)

func (m *MemoryStats) SendJSONMetrics(cfg *config.ConfigAgent) {
    for {
        time.Sleep(cfg.PauseDuration)
        for i := range m.GaugeMetrics {
            err := SendJSONGauge(i, cfg, m.GaugeMetrics)
            if err != nil {
                logger.Log.Info("unexpected sending json metric error:", zap.Error(err))
            }
        }
        err := SendJSONCounter(m.CounterMetric, cfg)
        if err != nil {
            logger.Log.Info("unexpected sending json metric error:", zap.Error(err))
        }
    }
}

func SendJSONGauge(metricName string, cfg *config.ConfigAgent, gaugeMetrics map[string]float64) error {
    agent := resty.New()
    valueGauge := gaugeMetrics[metricName]
    metric := models.Metrics{
        ID:    metricName,
        MType: config.GaugeType,
        Value: &valueGauge,
    }
    req := agent.R().SetHeader("Content-Type", "application/json").SetHeader("Content-Encoding", "gzip").
        SetHeader("Accept-Encoding", "gzip")
    
    metricsJSON, err := json.Marshal(metric)
    if err != nil {
        logger.Log.Info("marshalling json error:", zap.Error(err))
        return err
    }
    //
    if cfg.FlagHashKey != "" {
        key := []byte(cfg.FlagHashKey)
        h := hmac.New(sha256.New, key)
        h.Write(metricsJSON)
        dst := h.Sum(nil)
        req.SetHeader("HashSHA256", fmt.Sprintf("%x", dst))
    }
    //
    req.Method = resty.MethodPost
    var compressedRequest bytes.Buffer
    writer := gzip.NewWriter(&compressedRequest)
    _, err = writer.Write(metricsJSON)
    if err != nil {
        logger.Log.Info("error while compressing request:", zap.Error(err))
        return err
    }
    err = writer.Close()
    if err != nil {
        logger.Log.Info("error while closing gzip writer:", zap.Error(err))
        return err
    }
    _, err = req.SetBody(compressedRequest.Bytes()).Post(cfg.URL + "/update/")
    if err != nil {
        // send again n times if timeout error
        switch {
        case os.IsTimeout(err):
            for _, delay := range []time.Duration{time.Second, 2 * time.Second, 3 * time.Second} {
                time.Sleep(delay)
                if _, err = req.Post(req.URL); err == nil {
                    break
                }
                logger.Log.Info("timeout error, server not reachable:", zap.Error(err))
            }
            return ErrConnectionFailed
        default:
            logger.Log.Info("unexpected sending metric error:", zap.Error(err))
            return err
        }
    }
    return nil
}

func SendJSONCounter(counter int, cfg *config.ConfigAgent) error {
    agent := resty.New()
    valueDelta := int64(counter)
    metric := models.Metrics{
        ID:    config.PollCount,
        MType: config.CountType,
        Delta: &valueDelta,
    }
    req := agent.R().SetHeader("Content-Type", "application/json").SetHeader("Content-Encoding", "gzip").
        SetHeader("Accept-Encoding", "gzip")
    metricJSON, err := json.Marshal(metric)
    if err != nil {
        logger.Log.Info("marshalling json error:", zap.Error(err))
        return err
    }
    //
    if cfg.FlagHashKey != "" {
        key := []byte(cfg.FlagHashKey)
        h := hmac.New(sha256.New, key)
        h.Write(metricJSON)
        dst := h.Sum(nil)
        req.SetHeader("HashSHA256", fmt.Sprintf("%x", dst))
    }
    //
    var compressedRequest bytes.Buffer
    writer := gzip.NewWriter(&compressedRequest)
    _, err = writer.Write(metricJSON)
    if err != nil {
        logger.Log.Info("error while compressing request:", zap.Error(err))
        return err
    }
    err = writer.Close()
    if err != nil {
        logger.Log.Info("error while closing gzip writer:", zap.Error(err))
        return err
    }
    _, err = req.SetBody(compressedRequest.Bytes()).Post(cfg.URL + "/update/")
    if err != nil {
        // send again n times if timeout error
        switch {
        case os.IsTimeout(err):
            for _, delay := range []time.Duration{time.Second, 2 * time.Second, 3 * time.Second} {
                time.Sleep(delay)
                if _, err = req.Post(req.URL); err == nil {
                    break
                }
                logger.Log.Info("timeout error, server not reachable:", zap.Error(err))
            }
            return ErrConnectionFailed
        default:
            logger.Log.Info("unexpected sending metric error via URL:", zap.Error(err))
            return err
        }
    }
    return nil
}

