package repository

import "github.com/MaxBoych/MetricsService/internal/repository/memory"

type Repository interface {
	UpdateGauge(name string, new memory.Gauge) memory.Gauge
	UpdateCounter(name string, new memory.Counter) memory.Counter
	Count()
	GetGauge(name string) (string, bool)
	GetCounter(name string) (string, bool)
	GetAllMetrics() []string
}
