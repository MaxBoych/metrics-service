package storage

type Gauge float64
type Counter int64

type Repository interface {
	Replace(new Gauge)
	Add(new Counter)
}

type MemStorage struct {
	gauge   Gauge
	counter Counter
}

func (ms *MemStorage) Replace(new Gauge) {
	ms.gauge = new
}

func (ms *MemStorage) Add(new Counter) {
	ms.counter += new
}
