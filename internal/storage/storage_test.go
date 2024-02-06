package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemStorage_UpdateGauge(t *testing.T) {
	type want struct {
		gauges   map[string]Gauge
		counters map[string]Counter
	}
	tests := []struct {
		name   string
		gauges map[string]Gauge
		want   want
	}{
		{
			name:   "UPDATE GAUGE pass test 1",
			gauges: map[string]Gauge{"a": 123, "b": 456, "c": 789},
			want: want{
				gauges:   map[string]Gauge{"a": 123, "b": 456, "c": 789},
				counters: map[string]Counter{"PollCount": 3},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ms := NewMemStorage()
			for name, value := range test.gauges {
				ms.UpdateGauge(name, value)
			}

			for name, wanted := range test.want.gauges {
				value, ok := ms.Gauges[name]
				assert.Truef(t, ok, "Name '%s' must exist in the storage", name)
				assert.Equalf(t, wanted, value, "Value for '%s' must be '%v' but got '%v'", name, wanted, value)
			}
			assert.Equal(t, test.want.counters["PollCount"], ms.Counters["PollCount"], "The number of updated elements does not match the counter")
		})
	}
}

func TestMemStorage_UpdateCounter(t *testing.T) {
	type want struct {
		counters map[string]Counter
	}
	tests := []struct {
		name     string
		counters map[string]Counter
		want     want
	}{
		{
			name:     "UPDATE COUNTER pass test 1",
			counters: map[string]Counter{"a": 123, "b": 456, "c": 789},
			want: want{
				counters: map[string]Counter{"a": 246, "b": 912, "c": 1578, "PollCount": 6},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ms := NewMemStorage()
			for name, value := range test.counters {
				ms.UpdateCounter(name, value)
				ms.UpdateCounter(name, value)
			}

			for name, wanted := range test.want.counters {
				value, ok := ms.Counters[name]
				assert.Truef(t, ok, "Name '%s' must exist in the storage", name)
				assert.Equalf(t, wanted, value, "Value for '%s' must be '%v' but got '%v'", name, wanted, value)
			}
			assert.Equal(t, test.want.counters["PollCount"], ms.Counters["PollCount"], "The number of updated elements does not match the counter")
		})
	}
}

func TestMemStorage_GetGauge(t *testing.T) {
	type want struct {
		gauges map[string]string
	}
	tests := []struct {
		name   string
		gauges map[string]Gauge
		want   want
	}{
		{
			name:   "GET GAUGE pass test 1",
			gauges: map[string]Gauge{"a": 123, "b": 456, "c": 789},
			want: want{
				gauges: map[string]string{"a": "123", "b": "456", "c": "789"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ms := NewMemStorage()
			for name, value := range test.gauges {
				ms.UpdateGauge(name, value)
			}

			for name, wanted := range test.want.gauges {
				value, ok := ms.GetGauge(name)
				assert.Truef(t, ok, "Name '%s' must exist in the storage", name)
				assert.Equalf(t, wanted, value, "Value for '%s' must be '%s' but got '%s'", name, wanted, value)
			}
		})
	}
}

func TestMemStorage_GetCounter(t *testing.T) {
	type want struct {
		counters map[string]string
	}
	tests := []struct {
		name     string
		counters map[string]Counter
		want     want
	}{
		{
			name:     "GET COUNTER pass test 1",
			counters: map[string]Counter{"a": 123, "b": 456, "c": 789},
			want: want{
				counters: map[string]string{"a": "123", "b": "456", "c": "789"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ms := NewMemStorage()
			for name, value := range test.counters {
				ms.UpdateCounter(name, value)
			}

			for name, wanted := range test.want.counters {
				value, ok := ms.GetCounter(name)
				assert.Truef(t, ok, "Name '%s' must exist in the storage", name)
				assert.Equalf(t, wanted, value, "Value for '%s' must be '%s' but got '%s'", name, wanted, value)
			}
		})
	}
}
