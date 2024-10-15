package storage

import (
	"context"
	"database/sql"

	config "github.com/igortoigildin/go-metrics-altering/config/server"
	"github.com/igortoigildin/go-metrics-altering/internal/models"
	"github.com/igortoigildin/go-metrics-altering/pkg/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

const (
	GaugeType = "gauge"
	CountType = "counter"
	PollCount = "PollCount"
)

type PGStorage struct {
	conn     *sql.DB
	strategy Strategy
}

func NewPGStorage(conn *sql.DB) *PGStorage {
	return &PGStorage{
		conn: conn,
	}
}

func (pg *PGStorage) SetStrategy(metricType string) {
	if metricType == config.CountType {
		count := count{
			conn: pg.conn,
		}
		pg.strategy = &count
	} else {
		gauge := gauge{
			conn: pg.conn,
		}
		pg.strategy = &gauge
	}
}

func New(cfg *config.ConfigServer) *PGStorage {
	dbDSN := cfg.FlagDBDSN
	conn, err := sql.Open("pgx", dbDSN)
	if err != nil {
		logger.Log.Info("error while connecting to DB", zap.Error(err))
	}
	rep := NewPGStorage(conn)
	_, err = rep.conn.ExecContext(context.TODO(), "CREATE TABLE IF NOT EXISTS counters (id SERIAL, name TEXT NOT NULL,"+
		"type TEXT NOT NULL, value bigint, primary key(name));")
	if err != nil {
		logger.Log.Info("error while creating counters table", zap.Error(err))
	}
	_, err = rep.conn.ExecContext(context.TODO(), "CREATE TABLE IF NOT EXISTS gauges (id SERIAL, name TEXT NOT NULL,"+
		"type TEXT NOT NULL, value DOUBLE PRECISION, primary key(name));")
	if err != nil {
		logger.Log.Info("error while creating gauges table", zap.Error(err))
	}
	return rep
}

func (pg *PGStorage) Ping(ctx context.Context) error {
	err := pg.conn.PingContext(ctx)
	if err != nil {
		logger.Log.Info("connection to the database not alive", zap.Error(err))
	}
	return err
}

func (pg *PGStorage) Update(ctx context.Context, metricType string, metricName string, metricValue any) error {
	pg.SetStrategy(metricType)
	return pg.strategy.Update(ctx, metricType, metricName, metricValue)
}

func (pg *PGStorage) Get(ctx context.Context, metricType string, metricName string) (models.Metrics, error) {
	pg.SetStrategy(metricType)
	return pg.strategy.Get(ctx, metricType, metricName)
}

func (pg *PGStorage) GetAll(ctx context.Context) (map[string]any, error) {
	metrics := make(map[string]any, 33)
	rows, err := pg.conn.QueryContext(ctx, "SELECT name, value FROM gauges WHERE type = $1", GaugeType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		var value any
		err = rows.Scan(&name, &value)
		if err != nil {
			return nil, err
		}
		metrics[name] = value
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	rows, err = pg.conn.QueryContext(ctx, "SELECT name, value FROM counters WHERE type = $1", CountType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		var value any
		err = rows.Scan(&name, &value)
		if err != nil {
			return nil, err
		}
		metrics[name] = value
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return metrics, nil
}
