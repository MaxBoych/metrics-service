package memory

import (
	"context"
	"github.com/MaxBoych/MetricsService/internal/metrics/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemStorage_UpdateGauge(t *testing.T) {
	ctx := context.Background()

	type want struct {
		gauges   map[string]models.Gauge
		counters map[string]models.Counter
	}
	tests := []struct {
		name   string
		gauges map[string]models.Gauge
		want   want
	}{
		{
			name:   "UPDATE GAUGE pass test 1",
			gauges: map[string]models.Gauge{"a": 123, "b": 456, "c": 789},
			want: want{
				gauges:   map[string]models.Gauge{"a": 123, "b": 456, "c": 789},
				counters: map[string]models.Counter{"PollCount": 3},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ms := NewMemStorage()
			for name, value := range test.gauges {
				v := float64(value)
				m := models.Metrics{ID: name, MType: models.GaugeMetricName, Value: &v}
				ms.UpdateGauge(ctx, m)
			}

			for name, wanted := range test.want.gauges {
				value, ok := ms.Gauges[name]
				assert.Truef(t, ok, "Name '%s' must exist in the repository", name)
				assert.Equalf(t, wanted, value, "Value for '%s' must be '%v' but got '%v'", name, wanted, value)
			}
			assert.Equal(t, test.want.counters["PollCount"], ms.Counters["PollCount"], "The number of updated elements does not match the counter")
		})
	}
}

func TestMemStorage_UpdateCounter(t *testing.T) {
	ctx := context.Background()

	type want struct {
		counters map[string]models.Counter
	}
	tests := []struct {
		name     string
		counters map[string]models.Counter
		want     want
	}{
		{
			name:     "UPDATE COUNTER pass test 1",
			counters: map[string]models.Counter{"a": 123, "b": 456, "c": 789},
			want: want{
				counters: map[string]models.Counter{"a": 246, "b": 912, "c": 1578, "PollCount": 6},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ms := NewMemStorage()
			for name, value := range test.counters {
				v := int64(value)
				m := models.Metrics{ID: name, MType: models.CounterMetricName, Delta: &v}
				ms.UpdateCounter(ctx, m)
				ms.UpdateCounter(ctx, m)
			}

			for name, wanted := range test.want.counters {
				value, ok := ms.Counters[name]
				assert.Truef(t, ok, "Name '%s' must exist in the repository", name)
				assert.Equalf(t, wanted, value, "Value for '%s' must be '%v' but got '%v'", name, wanted, value)
			}
			assert.Equal(t, test.want.counters["PollCount"], ms.Counters["PollCount"], "The number of updated elements does not match the counter")
		})
	}
}

func TestMemStorage_GetGauge(t *testing.T) {
	ctx := context.Background()

	type want struct {
		gauges map[string]models.Gauge
	}
	tests := []struct {
		name   string
		gauges map[string]models.Gauge
		want   want
	}{
		{
			name:   "GET GAUGE pass test 1",
			gauges: map[string]models.Gauge{"a": 123.0, "b": 456.0, "c": 789.0},
			want: want{
				gauges: map[string]models.Gauge{"a": 123.0, "b": 456.0, "c": 789.0},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ms := NewMemStorage()
			for name, value := range test.gauges {
				v := float64(value)
				m := models.Metrics{ID: name, MType: models.GaugeMetricName, Value: &v}
				ms.UpdateGauge(ctx, m)
			}

			for name, wanted := range test.want.gauges {
				value, _ := ms.GetGauge(ctx, name)
				assert.NotNilf(t, value, "Name '%s' must exist in the repository", name)
				assert.Equalf(t, wanted, *value, "Value for '%s' must be '%v' but got '%s'", name, wanted, value.String())
			}
		})
	}
}

func TestMemStorage_GetCounter(t *testing.T) {
	ctx := context.Background()

	type want struct {
		counters map[string]models.Counter
	}
	tests := []struct {
		name     string
		counters map[string]models.Counter
		want     want
	}{
		{
			name:     "GET COUNTER pass test 1",
			counters: map[string]models.Counter{"a": 123, "b": 456, "c": 789},
			want: want{
				counters: map[string]models.Counter{"a": 123, "b": 456, "c": 789},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ms := NewMemStorage()
			for name, value := range test.counters {
				v := int64(value)
				m := models.Metrics{ID: name, MType: models.CounterMetricName, Delta: &v}
				ms.UpdateCounter(ctx, m)
			}

			for name, wanted := range test.want.counters {
				value, _ := ms.GetCounter(ctx, name)
				assert.NotNilf(t, value, "Name '%s' must exist in the repository", name)
				assert.Equalf(t, wanted, *value, "Value for '%s' must be '%v' but got '%s'", name, wanted, value)
			}
		})
	}
}
