package metrics

import (
	"context"
	"github.com/MaxBoych/MetricsService/internal/metrics/models"
)

type Repository interface {
	UpdateGauge(ctx context.Context, name string, new models.Gauge) *models.Gauge
	UpdateCounter(ctx context.Context, name string, new models.Counter) *models.Counter
	GetGauge(ctx context.Context, name string) *models.Gauge
	GetCounter(ctx context.Context, name string) *models.Counter
	GetAllMetrics(ctx context.Context) *models.Data
	UpdateMany(ctx context.Context, ms []models.Metrics) error
}
