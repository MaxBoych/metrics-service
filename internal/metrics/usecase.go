package metrics

import (
	"context"
	"github.com/MaxBoych/MetricsService/internal/metrics/models"
)

type UseCase interface {
	UpdateGauge(ctx context.Context, params models.Metrics) *models.Gauge
	UpdateCounter(ctx context.Context, params models.Metrics) *models.Counter
	GetGauge(ctx context.Context, params models.Metrics) *models.Gauge
	GetCounter(ctx context.Context, params models.Metrics) *models.Counter
	GetAllMetrics(ctx context.Context) *models.Data
	Ping(ctx context.Context) error
	UpdateMany(ctx context.Context, ms []models.Metrics) error
}
