package metrics

import (
	"context"
	"github.com/MaxBoych/MetricsService/internal/metrics/models"
)

type Repository interface {
	UpdateGauge(ctx context.Context, m models.Metrics) (*models.Metrics, error)
	UpdateCounter(ctx context.Context, m models.Metrics) (*models.Metrics, error)
	UpdateMany(ctx context.Context, ms []models.Metrics) ([]models.Metrics, error)

	GetGauge(ctx context.Context, name string) (*models.Gauge, error)
	GetCounter(ctx context.Context, name string) (*models.Counter, error)
	GetAll(ctx context.Context) (*models.Data, error)
}
