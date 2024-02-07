package storage

import (
	"context"
	"github.com/MaxBoych/MetricsService/internal/logger"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

type DBStorage struct {
	db *pgx.Conn
}

func NewDBStorage() *DBStorage {
	return &DBStorage{}
}

func (o *DBStorage) Connect(url string) {
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		logger.Log.Error("Unable to connect to database", zap.String("err", err.Error()))
		return
	}
	logger.Log.Info("connecting to database", zap.String("address", url))
	o.db = conn
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

func (o *DBStorage) CloseDB() {
	if o.db != nil {
		o.db.Close(context.Background())
	}
}
