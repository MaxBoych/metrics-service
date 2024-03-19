package memory

import (
	"context"
	"errors"
	"github.com/MaxBoych/MetricsService/internal/metrics/models"
	"sync"
)

type Storage struct {
	*models.Data
	Mu sync.RWMutex
}

func NewMemStorage() (ms *Storage) {
	ms = &Storage{}
	// Оставил функцию init, так как в ней содержится информация о метриках, что занимает слишком много места.
	// Пусть лучше это будет отдельной вспомогательной функцией.
	ms.init()
	return
}

func (o *Storage) GetAll(_ context.Context) (*models.Data, error) {
	o.Mu.RLock()
	defer o.Mu.RUnlock()

	return o.Data, nil
}

func (o *Storage) GetGauge(_ context.Context, name string) (*models.Gauge, error) {
	o.Mu.RLock()
	defer o.Mu.RUnlock()

	if value, ok := o.Data.Gauges[name]; ok {
		return &value, nil
	}
	return nil, errors.New("no such gauge metric")
}

func (o *Storage) GetCounter(_ context.Context, name string) (*models.Counter, error) {
	o.Mu.RLock()
	defer o.Mu.RUnlock()

	if value, ok := o.Data.Counters[name]; ok {
		return &value, nil
	}
	return nil, errors.New("no such counter metric")
}

func (o *Storage) UpdateGauge(_ context.Context, m models.Metrics) (*models.Metrics, error) {
	o.Mu.Lock()
	defer o.Mu.Unlock()

	o.Data.Gauges[m.ID] = models.Gauge(*m.Value)
	o.count()

	return &m, nil
}

func (o *Storage) UpdateCounter(_ context.Context, m models.Metrics) (*models.Metrics, error) {
	o.Mu.Lock()
	defer o.Mu.Unlock()

	o.Counters[m.ID] += models.Counter(*m.Delta)
	o.count()

	return &m, nil
}

func (o *Storage) count() {
	o.Counters["PollCount"]++
}

func (o *Storage) UpdateMany(ctx context.Context, ms []models.Metrics) ([]models.Metrics, error) {
	for _, m := range ms {
		if m.MType == models.GaugeMetricName {
			_, _ = o.UpdateGauge(ctx, m)
		} else if m.MType == models.CounterMetricName {
			_, _ = o.UpdateCounter(ctx, m)
		}
	}
	return ms, nil
}

func (o *Storage) init() {
	o.Mu.Lock()
	defer o.Mu.Unlock()

	o.Data = &models.Data{
		Counters: map[string]models.Counter{
			"PollCount": models.Counter(0),
		},

		Gauges: map[string]models.Gauge{

			//Alloc: метрика показывает количество байтов, которые в данный момент активно используются вашей программой.
			//Это не включает память, которая была выделена, но уже освобождена сборщиком мусора.
			"Alloc": models.Gauge(0),

			//Метрика показывает объем системной памяти, выделенной для внутренних структур данных карт Go.
			//Конкретно, это память, используемая для "buckets", которые являются основными строительными блоками хеш-таблицы.
			//Каждый bucket может содержать несколько элементов и служит для обеспечения быстрого доступа к элементам карты.
			"BuckHashSys": models.Gauge(0),

			//Это счетчик, который инкрементируется каждый раз, когда сборщик мусора в Go определяет,
			//что объект в памяти больше не достижим и его память может быть освобождена.
			//Это не количество свободной памяти, а количество операций освобождения.
			"Frees": models.Gauge(0),

			//Указывает, какую часть общего времени процессора использует сборщик мусора.
			//Например, значение 0.25 означает, что 25% времени CPU занято сборкой мусора.
			//Это позволяет разработчикам понять влияние сборки мусора на загрузку процессора и общую производительность приложения.
			"GCCPUFraction": models.Gauge(0),

			//Это количество байтов памяти, выделенное системой сбора мусора для метаданных, управления структурами и трекинга.
			//Сюда не входит память, выделенная непосредственно под объекты; это память, необходимая самому сборщику мусора для его работы.
			"GCSys": models.Gauge(0),

			//Объем байтов в куче, выделенных и все еще используемых. Это "живая" память, которая была выделена и еще не освобождена.
			"HeapAlloc": models.Gauge(0),

			//Объем байтов в куче, которые не используются в данный момент. Это память, которая была выделена,
			//но сейчас пустует и может быть возвращена операционной системе или использована для новых выделений.
			"HeapIdle": models.Gauge(0),

			//Объем байтов в куче, которые активно используются или недоступны для сборки мусора.
			"HeapInuse": models.Gauge(0),

			//Общее количество объектов в куче. Это число увеличивается с каждым новым объектом и уменьшается при сборке мусора.
			"HeapObjects": models.Gauge(0),

			//Объем памяти кучи, который был возвращен операционной системе.
			//Это память из HeapIdle, которая была фактически возвращена обратно в систему.
			"HeapReleased": models.Gauge(0),

			//Общий объем байтов, зарезервированных для кучи из системной памяти.
			"HeapSys": models.Gauge(0),

			//Время последней сборки мусора в наносекундах с момента старта.
			"LastGC": models.Gauge(0),

			//Предел памяти, при достижении которого будет запущена следующая сборка мусора.
			"NextGC": models.Gauge(0),

			//Количество сборок мусора, вызванных принудительно через runtime.GC().
			"NumForcedGC": models.Gauge(0),

			//Количество произведенных сборок мусора.
			"NumGC": models.Gauge(0),

			//Общее время пауз, вызванных сборкой мусора, в наносекундах.
			"PauseTotalNs": models.Gauge(0),

			//Общее количество операций выделения памяти.
			"Mallocs": models.Gauge(0),

			//Общий объем байтов, выделенных за все время (не уменьшается при сборке мусора).
			"TotalAlloc": models.Gauge(0),

			//Различная системная память, выделенная Go runtime, не попадающая под другие категории.
			"OtherSys": models.Gauge(0),

			//Общий объем байтов, зарезервированных для использования Go runtime из операционной системы.
			"Sys": models.Gauge(0),

			//Объем памяти, используемый mcache структурами. mcache предоставляет кэш на уровне потока для выделений малого размера.
			"MCacheInuse": models.Gauge(0),

			//Общий объем памяти, зарезервированный для mcache.
			"MCacheSys": models.Gauge(0),

			//Объем памяти, используемый mspan структурами. mspan используется для управления метаданными фрагментов памяти в куче.
			"MSpanInuse": models.Gauge(0),

			//Общий объем памяти, зарезервированный для mspan.
			"MSpanSys": models.Gauge(0),

			//Объем памяти в стеке, используемый текущими горутинами.
			"StackInuse": models.Gauge(0),

			//Общий объем памяти, зарезервированный для стека горутин.
			"StackSys": models.Gauge(0),

			//Количество поисковых запросов к хеш-таблице, которые не сопровождались выделением памяти.
			"Lookups": models.Gauge(0),

			"TotalMemory": models.Gauge(0),

			"FreeMemory": models.Gauge(0),

			"CPUutilization1": models.Gauge(0),
		},
	}
}
