package config

import (
	"errors"
	"flag"
	"html/template"
	"os"
	"strconv"
)

const (
    GaugeType = "gauge"
    CountType = "counter"
    PollCount = "PollCount"
)

var errCfgVarEmpty = errors.New("configs variable not set")

type ConfigServer struct {
    FlagRunAddr       string
    Template          *template.Template
    FlagLogLevel      string
    FlagStoreInterval int
    FlagStorePath     string
    FlagRestore       bool
    FlagDBDSN         string
    FlagHashKey       string
}

func LoadConfig() (*ConfigServer, error) {
    // ps := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",`localhost`, `postgres`, `XXXXX`, `metrics`)
    cfg := new(ConfigServer)
    var err error
    flag.StringVar(&cfg.FlagRunAddr, "a", ":8080", "address and port to run server")
    flag.StringVar(&cfg.FlagLogLevel, "l", "info", "log level")
    flag.IntVar(&cfg.FlagStoreInterval, "i", 1, "metrics backup interval")
    flag.StringVar(&cfg.FlagStorePath, "f", "/tmp/metrics-db.json", "metrics backup storage path")
    flag.BoolVar(&cfg.FlagRestore, "r", false, "true if load from backup is needed")
    flag.StringVar(&cfg.FlagDBDSN, "d", "", "string with DB DSN")
    flag.StringVar(&cfg.FlagHashKey, "k", "", "hash key")
    flag.Parse()
    if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
        cfg.FlagRunAddr = envRunAddr
    }
    if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
        cfg.FlagLogLevel = envLogLevel
    }
    if envHashKey := os.Getenv("KEY"); envHashKey != "" {
        cfg.FlagHashKey = envHashKey
    }
    if envStorageInterval := os.Getenv("STORE_INTERVAL"); envStorageInterval != "" {
        // parse string env variable
        v, err := strconv.Atoi(envStorageInterval)
        if err != nil {
            return nil, err
        }
        cfg.FlagStoreInterval = v
    }
    if envStorePath := os.Getenv("FILE_STORAGE_PATH"); envStorePath != "" {
        cfg.FlagStorePath = envStorePath
    }
    if envDBDSN := os.Getenv("DATABASE_DSN"); envDBDSN != "" {
        cfg.FlagDBDSN = envDBDSN
    }
    if envFlagRestore := os.Getenv("RESTORE"); envFlagRestore != "" {
        // parse bool env variable
        v, err := strconv.ParseBool(envFlagRestore)
        if err != nil {
            return nil, err
        }
        cfg.FlagRestore = v
    }
    // check if any config variables is empty
    if !cfg.validate() {
        return nil, errCfgVarEmpty
    }
    return cfg, err
}

func (cfg *ConfigServer) validate() bool {
    if cfg.FlagRunAddr == "" {
        return false
    }
    if cfg.FlagLogLevel == "" {
        return false
    }
    if cfg.FlagStorePath == "" {
        return false
    }
    return true
}

