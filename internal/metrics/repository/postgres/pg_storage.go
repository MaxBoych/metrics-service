package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/MaxBoych/MetricsService/internal/metrics/models"
	"github.com/MaxBoych/MetricsService/pkg/logger"
	"github.com/MaxBoych/MetricsService/pkg/values"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"log"
	"time"
)

type Storage struct {
	db *pgxpool.Pool
}

func NewDBStorage() *Storage {
	return &Storage{}
}

func (o *Storage) Connect(ctx context.Context, dsn string) error {
	var err error
	var pool *pgxpool.Pool

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return err
	}

	for _, interval := range values.RetryIntervals {
		pool, err = pgxpool.ConnectConfig(ctx, cfg)
		if err == nil {
			logger.Log.Info("connecting to database", zap.String("address", dsn))
			o.db = pool
			break
		}

		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) {
			logger.Log.Error("Unable to connect to database", zap.String("err", err.Error()))
			return err // ошибка не на стороне базы данных
		}

		logger.Log.Error("Failed to ping the database", zap.String("err", err.Error()), zap.String("interval", interval.String()))
		time.Sleep(interval)
	}

	if err != nil {
		logger.Log.Error("Failed to connect the database after retries:", zap.String("err", err.Error()))
		return err
	}

	if err = o.Ping(ctx); err != nil {
		return err
	}
	return nil
}

func (o *Storage) Ping(ctx context.Context) error {
	var err error

	for _, interval := range values.RetryIntervals {
		err = o.db.Ping(ctx)
		if err == nil {
			logger.Log.Info("Successfully pinged the database")
			return nil
		}

		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) {
			return err // ошибка не на стороне базы данных
		}

		logger.Log.Error("Failed to ping the database",
			zap.String("err", err.Error()),
			zap.String("interval", interval.String()),
		)
		time.Sleep(interval)
	}

	log.Println("Failed to ping the database after retries:", err)
	return err
}

func (o *Storage) executeTx(ctx context.Context, operation func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := o.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	err = operation(ctx, tx)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			logger.Log.Error("Failed to rollback transaction",
				zap.String("rollbackErr", rbErr.Error()),
				zap.String("originalErr", err.Error()),
			)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (o *Storage) Init(ctx context.Context) error {
	return o.executeTx(ctx, func(ctx context.Context, tx pgx.Tx) error {

		createGaugesTableSQL := fmt.Sprintf(`
    	CREATE TABLE IF NOT EXISTS "%s" (
        	"%s" BIGSERIAL PRIMARY KEY,
        	"%s" TEXT NOT NULL UNIQUE,
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
        	"%s" TEXT NOT NULL UNIQUE,
        	"%s" BIGINT NOT NULL,
        	"%s" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        	"%s" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    	);`,
			CountersTableName,
			IDColumnName,
			NameColumnName,
			ValueColumnName,
			CreatedAtColumnName,
			UpdatedAtColumnName)

		_, err := tx.Exec(context.Background(), createGaugesTableSQL)
		if err != nil {
			logger.Log.Error("Unable to create gauges table", zap.String("err", err.Error()))
			return err
		}

		_, err = tx.Exec(context.Background(), createCountersTableSQL)
		if err != nil {
			logger.Log.Error("Unable to create counters table", zap.String("err", err.Error()))
			return err
		}

		query, args, err := squirrel.Insert(CountersTableName).
			Columns(insertMetric...).
			Values(PollCountCounterName, 0, squirrel.Expr("NOW()"), squirrel.Expr("NOW()")).
			PlaceholderFormat(squirrel.Dollar).
			Suffix(fmt.Sprintf("ON CONFLICT (%[1]s) DO NOTHING", NameColumnName)).
			ToSql()
		if err != nil {
			logger.Log.Error("Cannot to build INSERT query", zap.String("err", err.Error()))
			return err
		}
		_, err = tx.Exec(ctx, query, args...)
		if err != nil {
			logger.Log.Error("Cannot to execute INSERT query", zap.String("err", err.Error()))
			return err
		}

		logger.Log.Info("Tables created successfully")
		return nil
	})
}

func (o *Storage) Close() {
	if o.db != nil {
		o.db.Close()
	}
}

func (o *Storage) UpdateGauge(ctx context.Context, m models.Metrics) (*models.Metrics, error) {
	err := o.executeTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return o.updateGauge(ctx, tx, m)
	})

	return &m, err
}

func (o *Storage) updateGauge(ctx context.Context, tx pgx.Tx, m models.Metrics) error {
	query, args, err := squirrel.Insert(GaugesTableName).
		Columns(insertMetric...).
		Values(m.ID, *m.Value, squirrel.Expr("NOW()"), squirrel.Expr("NOW()")).
		Suffix(fmt.Sprintf("ON CONFLICT (%[1]s) DO UPDATE SET %[2]s = EXCLUDED.%[2]s, %[3]s = EXCLUDED.%[3]s",
			NameColumnName,        // 1
			ValueColumnName,       // 2
			UpdatedAtColumnName)). // 3
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		logger.Log.Error("Cannot to build sql UPSERT query", zap.String("err", err.Error()))
		return err
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		logger.Log.Error("Cannot to execute sql UPSERT query", zap.String("err", err.Error()))
		return err
	}

	if err = o.count(ctx, tx); err != nil {
		return err
	}

	return nil
}

// Ответ на вопрос из ревью.
// Это каунтер-счетчик "PollCount". Каждый раз при изменении какой-либо метрики, происходит инкремент.
// Задача на этот PollCount стояла еще в первом спринте для in-memory хранилища.
// При переходе в PostreSQL мы по сути дублируем все методы на SQL-лад, поэтому и этот счетчик сюда также перекочевал.
func (o *Storage) count(ctx context.Context, tx pgx.Tx) error {
	incrementValue := 1
	query, args, err := squirrel.Update(CountersTableName).
		Set(ValueColumnName, squirrel.Expr(fmt.Sprintf("%s + ?", ValueColumnName), incrementValue)).
		Where(squirrel.Eq{NameColumnName: PollCountCounterName}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		logger.Log.Error("Cannot to build sql UPDATE query", zap.String("err", err.Error()))
		return err
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		logger.Log.Error("Cannot to execute sql UPDATE query", zap.String("err", err.Error()))
		return err
	}

	return nil
}

func (o *Storage) UpdateCounter(ctx context.Context, m models.Metrics) (*models.Metrics, error) {
	err := o.executeTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return o.updateCounter(ctx, tx, m)
	})

	return &m, err
}

func (o *Storage) updateCounter(ctx context.Context, tx pgx.Tx, m models.Metrics) error {
	query, args, err := squirrel.Insert(CountersTableName).
		Columns(insertMetric...).
		Values(m.ID, *m.Delta, squirrel.Expr("NOW()"), squirrel.Expr("NOW()")).
		Suffix(fmt.Sprintf("ON CONFLICT (%[1]s) DO UPDATE SET %[2]s = %[3]s.%[2]s + EXCLUDED.%[2]s, %[4]s = NOW()",
			NameColumnName,        // 1
			ValueColumnName,       // 2
			CountersTableName,     // 3
			UpdatedAtColumnName)). // 4
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		logger.Log.Error("Cannot to build sql UPSERT query", zap.String("err", err.Error()))
		return err
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		logger.Log.Error("Cannot to execute sql UPSERT query", zap.String("err", err.Error()))
		return err
	}

	if err = o.count(ctx, tx); err != nil {
		return err
	}

	return nil
}

func (o *Storage) GetGauge(ctx context.Context, name string) (*models.Gauge, error) {
	query, args, err := squirrel.Select(selectMetric...).
		From(GaugesTableName).
		Where(squirrel.Eq{NameColumnName: name}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		logger.Log.Error("Cannot to build sql SELECT query", zap.String("err", err.Error()))
		return nil, err
	}

	metric := models.GaugeMetric{}

	err = o.db.QueryRow(
		ctx,
		query,
		args...,
	).Scan(
		&metric.ID,
		&metric.Name,
		&metric.Value,
		&metric.CreatedAt,
		&metric.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Log.Error("There is no such row", zap.String("err", err.Error()))
			return nil, err
		}

		logger.Log.Error("Error while scanning, sql SELECT query", zap.String("err", err.Error()))
		return nil, err
	}

	gauge := models.Gauge(metric.Value)
	return &gauge, nil
}

func (o *Storage) GetCounter(ctx context.Context, name string) (*models.Counter, error) {
	query, args, err := squirrel.Select(selectMetric...).
		From(CountersTableName).
		Where(squirrel.Eq{NameColumnName: name}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		logger.Log.Error("Cannot to build sql SELECT query", zap.String("err", err.Error()))
		return nil, err
	}

	metric := models.CounterMetric{}

	err = o.db.QueryRow(
		ctx,
		query,
		args...,
	).Scan(
		&metric.ID,
		&metric.Name,
		&metric.Value,
		&metric.CreatedAt,
		&metric.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Log.Error("There is no such row", zap.String("err", err.Error()))
			return nil, err
		}

		logger.Log.Error("Error while scanning, sql SELECT query", zap.String("err", err.Error()))
		return nil, err
	}

	counter := models.Counter(metric.Value)
	return &counter, err
}

func (o *Storage) GetAll(ctx context.Context) (*models.Data, error) {
	data := &models.Data{
		Gauges:   make(map[string]models.Gauge),
		Counters: make(map[string]models.Counter),
	}

	err := o.executeTx(ctx, func(ctx context.Context, tx pgx.Tx) error {

		psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

		// gauges
		gaugeQuery, _, err := psql.Select(NameColumnName, ValueColumnName).From(GaugesTableName).ToSql()
		if err != nil {
			logger.Log.Error("ERROR make sql builder, sql SELECT query", zap.String("err", err.Error()))
			return err
		}

		gaugeRows, err := tx.Query(ctx, gaugeQuery)
		if err != nil {
			logger.Log.Error("ERROR execute sql SELECT query", zap.String("err", err.Error()))
			return err
		}
		defer gaugeRows.Close()

		for gaugeRows.Next() {
			var name string
			var value float64
			if err := gaugeRows.Scan(&name, &value); err != nil {
				logger.Log.Error("ERROR scan gauge rows", zap.String("err", err.Error()))
				return err
			}
			data.Gauges[name] = models.Gauge(value)
		}

		// counters
		counterQuery, _, err := psql.Select(NameColumnName, ValueColumnName).From(CountersTableName).ToSql()
		if err != nil {
			logger.Log.Error("ERROR make sql builder, sql SELECT query", zap.String("err", err.Error()))
			return err
		}

		counterRows, err := tx.Query(ctx, counterQuery)
		if err != nil {
			logger.Log.Error("ERROR execute sql SELECT query", zap.String("err", err.Error()))
			return err
		}
		defer counterRows.Close()

		for counterRows.Next() {
			var name string
			var value int64
			if err := counterRows.Scan(&name, &value); err != nil {
				logger.Log.Error("ERROR scan counter rows", zap.String("err", err.Error()))
				return err
			}
			data.Counters[name] = models.Counter(value)
		}

		return nil
	})

	return data, err
}

func (o *Storage) UpdateMany(ctx context.Context, ms []models.Metrics) ([]models.Metrics, error) {
	err := o.executeTx(ctx, func(ctx context.Context, tx pgx.Tx) error {

		var err error
		for _, m := range ms {
			if m.MType == models.GaugeMetricName {
				err = o.updateGauge(ctx, tx, m)
			} else if m.MType == models.CounterMetricName {
				err = o.updateCounter(ctx, tx, m)
			}
			if err != nil {
				return err
			}
		}

		return nil
	})

	return ms, err
}
