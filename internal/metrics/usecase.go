package metrics

import (
	"context"
	"github.com/MaxBoych/MetricsService/internal/metrics/models"
)

type UseCase interface {
	UpdateGauge(ctx context.Context, params models.Metrics) (*models.Metrics, error)
	UpdateCounter(ctx context.Context, params models.Metrics) (*models.Metrics, error)
	UpdateMany(ctx context.Context, ms []models.Metrics) ([]models.Metrics, error)

	GetGauge(ctx context.Context, params models.Metrics) (*models.Gauge, error)
	GetCounter(ctx context.Context, params models.Metrics) (*models.Counter, error)
	GetAll(ctx context.Context) (*models.Data, error)

	Ping(ctx context.Context) error
}
