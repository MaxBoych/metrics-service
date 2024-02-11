package postgres

import (
	"context"
	"github.com/MaxBoych/MetricsService/internal/metrics/repository/memory"
	"github.com/MaxBoych/MetricsService/pkg/logger"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

type DBStorage struct {
	ms *memory.MemStorage
	db *pgx.Conn
}

func NewDBStorage() *DBStorage {
	return &DBStorage{}
}

func (o *DBStorage) Connect(url string) error {
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		logger.Log.Error("Unable to connect to database", zap.String("err", err.Error()))
		return err
	}
	logger.Log.Info("connecting to database", zap.String("address", url))
	o.db = conn

	return nil
}

func (o *DBStorage) Ping() error {
	err := o.db.Ping(context.Background())
	if err != nil {
		logger.Log.Error("Unable to ping database", zap.String("err", err.Error()))
		return err
	}
	logger.Log.Error("successfully pinged to database")
	return nil
}

func (o *DBStorage) CreateTables() error {
	createGaugesTableSQL := `
    CREATE TABLE IF NOT EXISTS gauges (
        id BIGSERIAL PRIMARY KEY,
        "name" VARCHAR(255) NOT NULL,
        gauge DOUBLE PRECISION NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );`

	createCountersTableSQL := `
    CREATE TABLE IF NOT EXISTS counters (
        id BIGSERIAL PRIMARY KEY,
        "name" VARCHAR(255) NOT NULL,
        gauge BIGINT NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );`

	_, err := o.db.Exec(context.Background(), createGaugesTableSQL)
	if err != nil {
		logger.Log.Error("Unable to create gauges table", zap.String("err", err.Error()))
		return err
	}

	_, err = o.db.Exec(context.Background(), createCountersTableSQL)
	if err != nil {
		logger.Log.Error("Unable to create counters table", zap.String("err", err.Error()))
		return err
	}

	logger.Log.Info("Tables created successfully")
	return nil
}

func (o *DBStorage) Close() {
	if o.db != nil {
		o.db.Close(context.Background())
	}
}
