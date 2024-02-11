package models

import (
	"fmt"
	"time"
)

type Gauge float64
type Counter int64

func (o *Gauge) String() string {
	return fmt.Sprintf("%g", *o)
}

func (o *Counter) String() string {
	return fmt.Sprintf("%d", *o)
}

type Data struct {
	Gauges   map[string]Gauge
	Counters map[string]Counter
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (o *Metrics) String() string {
	return fmt.Sprintf("ID: %s, MType: %s, Delta: %d, Value: %g", o.ID, o.MType, o.Delta, o.Value)
}

type GaugeMetric struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Value     float64   `db:"value"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type CounterMetric struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Value     int64     `db:"value"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
