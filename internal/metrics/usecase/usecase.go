package usecase

import (
	"context"
	"errors"
	"github.com/MaxBoych/MetricsService/internal/metrics"
	"github.com/MaxBoych/MetricsService/internal/metrics/models"
	"github.com/MaxBoych/MetricsService/internal/metrics/repository/postgres"
	"github.com/MaxBoych/MetricsService/pkg/logger"
	"github.com/MaxBoych/MetricsService/pkg/values"
	"github.com/jackc/pgconn"
	"go.uber.org/zap"
	"os"
	"time"
)

type MetricsUseCase struct {
	repo metrics.Repository
}

func NewMetricsUseCase(repo metrics.Repository) *MetricsUseCase {
	return &MetricsUseCase{
		repo: repo,
	}
}

func (o *MetricsUseCase) GetAll(ctx context.Context) (*models.Data, error) {
	res, err := executeWithRetries(ctx, func(ctx context.Context) (interface{}, error) {
		return o.repo.GetAll(ctx)
	})

	if err != nil {
		logger.Log.Error("Not storage problem, bad request", zap.String("err", err.Error()))
	} else if data, ok := res.(*models.Data); ok {
		return data, nil
	}
	return nil, errors.New("unsupported type")
}

func (o *MetricsUseCase) GetGauge(ctx context.Context, m models.Metrics) (*models.Gauge, error) {
	res, err := executeWithRetries(ctx, func(ctx context.Context) (interface{}, error) {
		return o.repo.GetGauge(ctx, m.ID)
	})

	if err != nil {
		logger.Log.Error("Not storage problem, bad request", zap.String("err", err.Error()))
	} else if data, ok := res.(*models.Gauge); ok {
		return data, nil
	}
	return nil, errors.New("unsupported type")
}

func (o *MetricsUseCase) GetCounter(ctx context.Context, m models.Metrics) (*models.Counter, error) {
	res, err := executeWithRetries(ctx, func(ctx context.Context) (interface{}, error) {
		return o.repo.GetCounter(ctx, m.ID)
	})

	if err != nil {
		logger.Log.Error("Not storage problem, bad request", zap.String("err", err.Error()))
	} else if data, ok := res.(*models.Counter); ok {
		return data, nil
	}
	return nil, errors.New("unsupported type")
}

func (o *MetricsUseCase) UpdateGauge(ctx context.Context, m models.Metrics) (*models.Metrics, error) {
	res, err := executeWithRetries(ctx, func(ctx context.Context) (interface{}, error) {
		return o.repo.UpdateGauge(ctx, m)
	})

	if err != nil {
		logger.Log.Error("Not storage problem, bad request", zap.String("err", err.Error()))
	} else if data, ok := res.(*models.Metrics); ok {
		return data, nil
	}
	return nil, errors.New("unsupported type")
}

func (o *MetricsUseCase) UpdateCounter(ctx context.Context, m models.Metrics) (*models.Metrics, error) {
	res, err := executeWithRetries(ctx, func(ctx context.Context) (interface{}, error) {
		return o.repo.UpdateCounter(ctx, m)
	})

	if err != nil {
		logger.Log.Error("Not storage problem, bad request", zap.String("err", err.Error()))
	} else if data, ok := res.(*models.Metrics); ok {
		return data, nil
	}
	return nil, errors.New("unsupported type")
}

func (o *MetricsUseCase) UpdateMany(ctx context.Context, ms []models.Metrics) ([]models.Metrics, error) {
	res, err := executeWithRetries(ctx, func(ctx context.Context) (interface{}, error) {
		return o.repo.UpdateMany(ctx, ms)
	})

	if err != nil {
		logger.Log.Error("Not storage problem, bad request", zap.String("err", err.Error()))
	} else if data, ok := res.([]models.Metrics); ok {
		return data, nil
	}
	return nil, errors.New("unsupported type")
}

func (o *MetricsUseCase) Ping(ctx context.Context) error {
	_, err := executeWithRetries(ctx, func(ctx context.Context) (interface{}, error) {
		db, ok := o.repo.(*postgres.PGStorage)
		if ok {
			return nil, db.Ping(ctx)
		}
		return nil, errors.New("storage is not PG")
	})
	return err
}

func executeWithRetries(
	ctx context.Context,
	operation func(ctx context.Context) (interface{}, error),
) (interface{}, error) {

	var result interface{}
	var err error

	for _, interval := range values.RetryIntervals {
		result, err = operation(ctx)
		if err == nil {
			return result, nil
		}

		var pgErr *pgconn.PgError
		var fileErr *os.PathError
		if !errors.As(err, &pgErr) && !errors.As(err, &fileErr) {
			return nil, err // ошибка не на стороне storage
		}

		logger.Log.Error("Failed to do operation",
			zap.String("err", err.Error()),
			zap.String("try again in", interval.String()),
		)

		time.Sleep(interval)
	}

	result, err = operation(ctx)
	return result, err
}
