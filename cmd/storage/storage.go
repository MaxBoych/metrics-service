package storage

import "fmt"

type Gauge float64
type Counter int64

type Repository interface {
	Init()
	UpdateGauge(name string, new Gauge)
	UpdateCounter(name string, new Counter)
	count()
	GetGauge(name string) (string, bool)
	GetCounter(name string) (string, bool)
	GetAllMetrics() []string
}

type MemStorage struct {
	Gauges   map[string]Gauge
	Counters map[string]Counter
}

func (ms *MemStorage) GetAllMetrics() (metrics []string) {
	//var metrics []string
	for k, v := range ms.Gauges {
		metrics = append(metrics, fmt.Sprintf("%s: %v", k, v))
	}
	for k, v := range ms.Counters {
		metrics = append(metrics, fmt.Sprintf("%s: %v", k, v))
	}
	return
}

func (ms *MemStorage) GetGauge(name string) (string, bool) {
	if value, ok := ms.Gauges[name]; ok {
		value := fmt.Sprintf("%v", value)
		return value, true
	}
	return "", false
}

func (ms *MemStorage) GetCounter(name string) (string, bool) {
	if value, ok := ms.Counters[name]; ok {
		value := fmt.Sprintf("%v", value)
		return value, true
	}
	return "", false
}

func (ms *MemStorage) UpdateGauge(name string, new Gauge) {
	ms.Gauges[name] = new
	ms.count()
}

func (ms *MemStorage) UpdateCounter(name string, new Counter) {
	ms.Counters[name] += new
	ms.count()
}

func (ms *MemStorage) count() {
	ms.Counters["PollCount"]++
}

//func (ms *MemStorage) ContainsGauge(name string) bool {
//	if _, ok := ms.Gauges[name]; ok {
//		return true
//	} else {
//		return false
//	}
//}
//
//func (ms *MemStorage) ContainsCounter(name string) bool {
//	if _, ok := ms.Counters[name]; ok {
//		return true
//	} else {
//		return false
//	}
//}

func (ms *MemStorage) Init() {
	ms.Counters = map[string]Counter{
		"PollCount": Counter(0),
		//"testCounter": Counter(0),
	}

	ms.Gauges = map[string]Gauge{

		//Alloc: метрика показывает количество байтов, которые в данный момент активно используются вашей программой.
		//Это не включает память, которая была выделена, но уже освобождена сборщиком мусора.
		"Alloc": Gauge(0),

		//Метрика показывает объем системной памяти, выделенной для внутренних структур данных карт Go.
		//Конкретно, это память, используемая для "buckets", которые являются основными строительными блоками хеш-таблицы.
		//Каждый bucket может содержать несколько элементов и служит для обеспечения быстрого доступа к элементам карты.
		"BuckHashSys": Gauge(0),

		//Это счетчик, который инкрементируется каждый раз, когда сборщик мусора в Go определяет,
		//что объект в памяти больше не достижим и его память может быть освобождена.
		//Это не количество свободной памяти, а количество операций освобождения.
		"Frees": Gauge(0),

		//Указывает, какую часть общего времени процессора использует сборщик мусора.
		//Например, значение 0.25 означает, что 25% времени CPU занято сборкой мусора.
		//Это позволяет разработчикам понять влияние сборки мусора на загрузку процессора и общую производительность приложения.
		"GCCPUFraction": Gauge(0),

		//Это количество байтов памяти, выделенное системой сбора мусора для метаданных, управления структурами и трекинга.
		//Сюда не входит память, выделенная непосредственно под объекты; это память, необходимая самому сборщику мусора для его работы.
		"GCSys": Gauge(0),

		//Объем байтов в куче, выделенных и все еще используемых. Это "живая" память, которая была выделена и еще не освобождена.
		"HeapAlloc": Gauge(0),

		//Объем байтов в куче, которые не используются в данный момент. Это память, которая была выделена,
		//но сейчас пустует и может быть возвращена операционной системе или использована для новых выделений.
		"HeapIdle": Gauge(0),

		//Объем байтов в куче, которые активно используются или недоступны для сборки мусора.
		"HeapInuse": Gauge(0),

		//Общее количество объектов в куче. Это число увеличивается с каждым новым объектом и уменьшается при сборке мусора.
		"HeapObjects": Gauge(0),

		//Объем памяти кучи, который был возвращен операционной системе.
		//Это память из HeapIdle, которая была фактически возвращена обратно в систему.
		"HeapReleased": Gauge(0),

		//Общий объем байтов, зарезервированных для кучи из системной памяти.
		"HeapSys": Gauge(0),

		//Время последней сборки мусора в наносекундах с момента старта.
		"LastGC": Gauge(0),

		//Предел памяти, при достижении которого будет запущена следующая сборка мусора.
		"NextGC": Gauge(0),

		//Количество сборок мусора, вызванных принудительно через runtime.GC().
		"NumForcedGC": Gauge(0),

		//Количество произведенных сборок мусора.
		"NumGC": Gauge(0),

		//Общее время пауз, вызванных сборкой мусора, в наносекундах.
		"PauseTotalNs": Gauge(0),

		//Общее количество операций выделения памяти.
		"Mallocs": Gauge(0),

		//Общий объем байтов, выделенных за все время (не уменьшается при сборке мусора).
		"TotalAlloc": Gauge(0),

		//Различная системная память, выделенная Go runtime, не попадающая под другие категории.
		"OtherSys": Gauge(0),

		//Общий объем байтов, зарезервированных для использования Go runtime из операционной системы.
		"Sys": Gauge(0),

		//Объем памяти, используемый mcache структурами. mcache предоставляет кэш на уровне потока для выделений малого размера.
		"MCacheInuse": Gauge(0),

		//Общий объем памяти, зарезервированный для mcache.
		"MCacheSys": Gauge(0),

		//Объем памяти, используемый mspan структурами. mspan используется для управления метаданными фрагментов памяти в куче.
		"MSpanInuse": Gauge(0),

		//Общий объем памяти, зарезервированный для mspan.
		"MSpanSys": Gauge(0),

		//Объем памяти в стеке, используемый текущими горутинами.
		"StackInuse": Gauge(0),

		//Общий объем памяти, зарезервированный для стека горутин.
		"StackSys": Gauge(0),

		//Количество поисковых запросов к хеш-таблице, которые не сопровождались выделением памяти.
		"Lookups": Gauge(0),

		//Кастомное рандомное значение
		//"RandomValue": Gauge(0),

		//Значение для тестов
		//"testGauge": Gauge(0),
	}
}
