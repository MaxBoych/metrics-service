package usecase

import (
	"context"
	"errors"
	"github.com/MaxBoych/MetricsService/internal/metrics"
	"github.com/MaxBoych/MetricsService/internal/metrics/models"
	"github.com/MaxBoych/MetricsService/internal/metrics/repository/postgres"
)

type MetricsUseCase struct {
	repo metrics.Repository
}

func NewMetricsUseCase(repo metrics.Repository) *MetricsUseCase {
	return &MetricsUseCase{
		repo: repo,
	}
}

func (o *MetricsUseCase) GetAllMetrics(ctx context.Context) *models.Data {
	data := o.repo.GetAllMetrics(ctx)
	return data
}

func (o *MetricsUseCase) GetGauge(ctx context.Context, params models.Metrics) *models.Gauge {
	return o.repo.GetGauge(ctx, params.ID)
}

func (o *MetricsUseCase) GetCounter(ctx context.Context, params models.Metrics) *models.Counter {
	return o.repo.GetCounter(ctx, params.ID)
}

func (o *MetricsUseCase) UpdateGauge(ctx context.Context, params models.Metrics) *models.Gauge {
	return o.repo.UpdateGauge(ctx, params.ID, models.Gauge(*params.Value))
}

func (o *MetricsUseCase) UpdateCounter(ctx context.Context, params models.Metrics) *models.Counter {
	return o.repo.UpdateCounter(ctx, params.ID, models.Counter(*params.Delta))
}

func (o *MetricsUseCase) Ping(ctx context.Context) error {
	db, ok := o.repo.(*postgres.PGStorage)
	if ok {
		return db.Ping(ctx)
	}
	return errors.New("repo is not db")
}
