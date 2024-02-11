package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/MaxBoych/MetricsService/internal/metrics/models"
	"github.com/MaxBoych/MetricsService/pkg/logger"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"time"
)

type PGStorage struct {
	db *pgx.Conn
}

func NewDBStorage() *PGStorage {
	return &PGStorage{}
}

func (o *PGStorage) Connect(url string) error {
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		logger.Log.Error("Unable to connect to database", zap.String("err", err.Error()))
		return err
	}
	logger.Log.Info("connecting to database", zap.String("address", url))
	o.db = conn
	err = o.db.Ping(context.Background())
	if err != nil {
		logger.Log.Error("Cannot to ping database", zap.String("err", err.Error()))
		return err
	} else {
		logger.Log.Error("PING was fine")
	}

	return nil
}

func (o *PGStorage) Ping(ctx context.Context) error {
	err := o.db.Ping(ctx)
	if err != nil {
		logger.Log.Error("Unable to ping database", zap.String("err", err.Error()))
		return err
	}
	logger.Log.Error("successfully pinged to database")
	return nil
}

func (o *PGStorage) Init() error {
	ctx := context.Background()

	createGaugesTableSQL := fmt.Sprintf(`
    CREATE TABLE IF NOT EXISTS "%s" (
        "%s" BIGSERIAL PRIMARY KEY,
        "%s" VARCHAR(255) NOT NULL,
        "%s" DOUBLE PRECISION NOT NULL,
        "%s" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        "%s" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );`,
		GaugesTableName,
		IDColumnName,
		NameColumnName,
		ValueColumnName,
		CreatedAtColumnName,
		UpdatedAtColumnName)

	createCountersTableSQL := fmt.Sprintf(`
    CREATE TABLE IF NOT EXISTS "%s" (
        "%s" BIGSERIAL PRIMARY KEY,
        "%s" VARCHAR(255) NOT NULL,
        "%s" DOUBLE PRECISION NOT NULL,
        "%s" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        "%s" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );`,
		CountersTableName,
		IDColumnName,
		NameColumnName,
		ValueColumnName,
		CreatedAtColumnName,
		UpdatedAtColumnName)

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

	query, args, err := squirrel.Insert(CountersTableName).
		Columns(NameColumnName, ValueColumnName).
		Values(PollCountCounterName, 0).
		PlaceholderFormat(squirrel.Dollar).
		//Suffix(fmt.Sprintf("ON CONFLICT (%s) DO NOTHING", NameColumnName)).
		ToSql()
	if err != nil {
		logger.Log.Error("Cannot to build INSERT query", zap.String("err", err.Error()))
		return err
	}
	_, err = o.db.Exec(ctx, query, args...)
	if err != nil {
		logger.Log.Error("Cannot to execute INSERT query", zap.String("err", err.Error()))
		return err
	}

	logger.Log.Info("Tables created successfully")
	return nil
}

func (o *PGStorage) Close() {
	if o.db != nil {
		o.db.Close(context.Background())
	}
}

func (o *PGStorage) UpdateGauge(ctx context.Context, name string, new models.Gauge) *models.Gauge {
	query, args, err := squirrel.Update(GaugesTableName).
		Set(ValueColumnName, float64(new)).
		Set(UpdatedAtColumnName, time.Now()).
		Where(squirrel.Eq{NameColumnName: name}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		logger.Log.Error("Cannot to build sql UPDATE query", zap.String("err", err.Error()))
		return nil
	}

	res, err := o.db.Exec(ctx, query, args...)
	if err != nil {
		logger.Log.Error("Cannot to execute sql UPDATE query", zap.String("err", err.Error()))
		return nil
	}

	gauge := &new
	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		gauge = o.insertGauge(ctx, name, new)
	}
	o.count(ctx)
	return gauge
}

func (o *PGStorage) count(ctx context.Context) {
	incrementValue := 1
	query, args, err := squirrel.Update(CountersTableName).
		Set(ValueColumnName, squirrel.Expr(fmt.Sprintf("%s + ?", ValueColumnName), incrementValue)).
		Where(squirrel.Eq{NameColumnName: PollCountCounterName}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		// Наверное, тут должен быть Tx.Rollback, но я пока не смог нормально разобраться в транзакциях
		logger.Log.Error("Cannot to build sql UPDATE query", zap.String("err", err.Error()))
		return
	}

	_, err = o.db.Exec(ctx, query, args...)
	if err != nil {
		// И тут тоже
		logger.Log.Error("Cannot to execute sql UPDATE query", zap.String("err", err.Error()))
		return
	}
}

func (o *PGStorage) insertGauge(ctx context.Context, name string, new models.Gauge) *models.Gauge {
	query, args, err := squirrel.Insert(GaugesTableName).
		Columns(insertMetric...).
		Values(
			name,
			float64(new),
			time.Now(),
			time.Now(),
		).
		Suffix("RETURNING " + ValueColumnName).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		logger.Log.Error("Cannot to build sql INSERT query", zap.String("err", err.Error()))
		return nil
	}

	var value float64
	err = o.db.QueryRow(ctx, query, args...).Scan(&value)
	if err != nil {
		logger.Log.Error("Cannot to execute sql INSERT query", zap.String("err", err.Error()))
		return nil
	}
	gauge := models.Gauge(value)
	return &gauge
}

func (o *PGStorage) UpdateCounter(ctx context.Context, name string, new models.Counter) *models.Counter {
	query, args, err := squirrel.Update(CountersTableName).
		Set(ValueColumnName, int64(new)).
		Set(UpdatedAtColumnName, time.Now()).
		Where(squirrel.Eq{NameColumnName: name}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		logger.Log.Error("Cannot to build sql UPDATE query", zap.String("err", err.Error()))
		return nil
	}

	res, err := o.db.Exec(ctx, query, args...)
	if err != nil {
		logger.Log.Error("Cannot to execute sql UPDATE query", zap.String("err", err.Error()))
		return nil
	}

	counter := &new
	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		counter = o.insertCounter(ctx, name, new)
	}
	o.count(ctx)
	return counter
}

func (o *PGStorage) insertCounter(ctx context.Context, name string, new models.Counter) *models.Counter {
	query, args, err := squirrel.Insert(CountersTableName).
		Columns(insertMetric...).
		Values(
			name,
			int64(new),
			time.Now(),
			time.Now(),
		).
		Suffix("RETURNING " + ValueColumnName).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		logger.Log.Error("Cannot to build sql INSERT query", zap.String("err", err.Error()))
		return nil
	}

	var value int64
	err = o.db.QueryRow(ctx, query, args...).Scan(&value)
	if err != nil {
		logger.Log.Error("Cannot to execute sql INSERT query", zap.String("err", err.Error()))
		return nil
	}
	counter := models.Counter(value)
	return &counter
}

func (o *PGStorage) GetGauge(ctx context.Context, name string) *models.Gauge {
	query, args, err := squirrel.Select(selectMetric...).
		From(GaugesTableName).
		Where(squirrel.Eq{NameColumnName: name}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		logger.Log.Error("Cannot to build sql SELECT query", zap.String("err", err.Error()))
		return nil
	}

	metric := models.GaugeMetric{}

	err = o.db.QueryRow(
		ctx,
		query,
		args...,
	).Scan(&metric)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Log.Error("There is no such row", zap.String("err", err.Error()))
			return nil
		}

		logger.Log.Error("Error while scanning, sql SELECT query", zap.String("err", err.Error()))
		return nil
	}

	gauge := models.Gauge(metric.Value)
	return &gauge
}

func (o *PGStorage) GetCounter(ctx context.Context, name string) *models.Counter {
	query, args, err := squirrel.Select(selectMetric...).
		From(CountersTableName).
		Where(squirrel.Eq{NameColumnName: name}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		logger.Log.Error("Cannot to build sql SELECT query", zap.String("err", err.Error()))
		return nil
	}

	metric := models.CounterMetric{}

	err = o.db.QueryRow(
		ctx,
		query,
		args...,
	).Scan(&metric)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Log.Error("There is no such row", zap.String("err", err.Error()))
			return nil
		}

		logger.Log.Error("Error while scanning, sql SELECT query", zap.String("err", err.Error()))
		return nil
	}

	counter := models.Counter(metric.Value)
	return &counter
}

func (o *PGStorage) GetAllMetrics(ctx context.Context) *models.Data {
	data := &models.Data{
		Gauges:   make(map[string]models.Gauge),
		Counters: make(map[string]models.Counter),
	}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	// gauges
	gaugeQuery, _, err := psql.Select(NameColumnName, ValueColumnName).From(GaugesTableName).ToSql()
	if err != nil {
		logger.Log.Error("ERROR make sql builder, sql SELECT query", zap.String("err", err.Error()))
		return nil
	}

	gaugeRows, err := o.db.Query(ctx, gaugeQuery)
	if err != nil {
		logger.Log.Error("ERROR execute sql SELECT query", zap.String("err", err.Error()))
		return nil
	}
	defer gaugeRows.Close()

	for gaugeRows.Next() {
		var name string
		var value float64
		if err := gaugeRows.Scan(&name, &value); err != nil {
			logger.Log.Error("ERROR scan gauge rows", zap.String("err", err.Error()))
			return nil
		}
		data.Gauges[name] = models.Gauge(value)
	}

	// counters
	counterQuery, _, err := psql.Select(NameColumnName, ValueColumnName).From(CountersTableName).ToSql()
	if err != nil {
		logger.Log.Error("ERROR make sql builder, sql SELECT query", zap.String("err", err.Error()))
		return nil
	}

	counterRows, err := o.db.Query(ctx, counterQuery)
	if err != nil {
		logger.Log.Error("ERROR execute sql SELECT query", zap.String("err", err.Error()))
		return nil
	}
	defer counterRows.Close()

	for counterRows.Next() {
		var name string
		var value int64
		if err := counterRows.Scan(&name, &value); err != nil {
			logger.Log.Error("ERROR scan counter rows", zap.String("err", err.Error()))
			return nil
		}
		data.Counters[name] = models.Counter(value)
	}

	return data
}
