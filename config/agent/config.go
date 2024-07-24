package config

import (
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/igortoigildin/go-metrics-altering/internal/logger"
	"go.uber.org/zap"
)

const (
    PollInterval   = 2
    GaugeType      = "gauge"
    CountType      = "counter"
    PollCount      = "PollCount"
    StatusOK       = 200
    ProtocolScheme = "http://"
)

type ConfigAgent struct {
    FlagRunAddr        string
    FlagReportInterval int
    FlagPollInterval   int
    FlagLogLevel       string
    FlagHashKey        string
    PauseDuration      time.Duration // Time agent will wait to send metrics again
    URL                string
}

func LoadConfig() (*ConfigAgent, error) {
    cfg := new(ConfigAgent)
    var err error
    flag.StringVar(&cfg.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
    flag.StringVar(&cfg.FlagLogLevel, "l", "info", "log level")
    flag.IntVar(&cfg.FlagReportInterval, "r", 10, "frequency of metrics being sent to the server")
    flag.IntVar(&cfg.FlagPollInterval, "p", 2, "frequency of metrics being received from the runtime package")
    flag.StringVar(&cfg.FlagHashKey, "k", "", "hash key")
    flag.Parse()
    if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
        cfg.FlagRunAddr = envRunAddr
    }
    if envHashValue := os.Getenv("KEY"); envHashValue != "" {
        cfg.FlagHashKey = envHashValue
    }
    if envRoportInterval := os.Getenv("REPORT_INTERVAL"); envRoportInterval != "" {
        cfg.FlagReportInterval, err = strconv.Atoi(envRoportInterval)
        if err != nil {
            logger.Log.Fatal("error while parsing report interval", zap.Error(err))
        }
    }
    if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
        cfg.FlagPollInterval, err = strconv.Atoi(envPollInterval)
        if err != nil {
            logger.Log.Fatal("error while parsing poll interval", zap.Error(err))
        }
    }
    if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
        cfg.FlagLogLevel = envLogLevel
    }
    cfg.PauseDuration = time.Duration(cfg.FlagReportInterval) * time.Second
    cfg.URL = ProtocolScheme + cfg.FlagRunAddr
    return cfg, err
}

