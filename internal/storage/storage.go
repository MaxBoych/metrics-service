package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/MaxBoych/MetricsService/internal/logger"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"os"
	"sync"
)

type Gauge float64
type Counter int64

type MemStorageData struct {
	Gauges   map[string]Gauge
	Counters map[string]Counter
}

type MemStorage struct {
	MemStorageData
	Mu       sync.Mutex
	FilePath string
	AutoSave bool
	db       *pgx.Conn
}

func NewMemStorage() (storage *MemStorage) {
	storage = &MemStorage{}
	// Оставил функцию init, так как в ней содержится информация о метриках, что занимает слишком много места.
	// Пусть лучше это будет отдельной вспомогательной функцией.//
	storage.Init()
	return
}

func (ms *MemStorage) GetAllMetrics() (metrics []string) {
	ms.Mu.Lock()
	defer ms.Mu.Unlock()

	for k, v := range ms.Gauges {
		metrics = append(metrics, fmt.Sprintf("%s: %v", k, v))
	}
	for k, v := range ms.Counters {
		metrics = append(metrics, fmt.Sprintf("%s: %v", k, v))
	}
	return
}

func (ms *MemStorage) GetGauge(name string) (string, bool) {
	ms.Mu.Lock()
	defer ms.Mu.Unlock()

	if value, ok := ms.Gauges[name]; ok {
		value := fmt.Sprintf("%v", value)
		return value, true
	}
	return "", false
}

func (ms *MemStorage) GetCounter(name string) (string, bool) {
	ms.Mu.Lock()
	defer ms.Mu.Unlock()

	if value, ok := ms.Counters[name]; ok {
		value := fmt.Sprintf("%v", value)
		return value, true
	}
	return "", false
}

func (ms *MemStorage) UpdateGauge(name string, new Gauge) Gauge {
	ms.Mu.Lock()
	defer ms.Mu.Unlock()

	ms.Gauges[name] = new
	ms.Count()

	if ms.AutoSave {
		ms.saveOnChange()
	}

	return ms.Gauges[name]
}

func (ms *MemStorage) UpdateCounter(name string, new Counter) Counter {
	ms.Mu.Lock()
	defer ms.Mu.Unlock()

	ms.Counters[name] += new
	ms.Count()

	if ms.AutoSave {
		ms.saveOnChange()
	}

	return ms.Counters[name]
}

func (ms *MemStorage) Count() {
	ms.Counters["PollCount"]++
}

func (ms *MemStorage) LoadFromFile() error {
	ms.Mu.Lock()
	defer ms.Mu.Unlock()

	data, err := os.ReadFile(ms.FilePath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, ms); err != nil {
		return err
	}
	return nil
}

func (ms *MemStorage) saveOnChange() {
	err := ms.storeToFile()
	if err != nil {
		logger.Log.Info("ERROR store to file", zap.String("error", err.Error()))
	}
}

func (ms *MemStorage) StoreToFile() error {
	ms.Mu.Lock()
	defer ms.Mu.Unlock()
	return ms.storeToFile()
}

func (ms *MemStorage) storeToFile() error {
	data, err := json.MarshalIndent(ms, "", "   ")
	if err != nil {
		return err
	}
	return os.WriteFile(ms.FilePath, data, 0666)
}

func (ms *MemStorage) SetDB(db *pgx.Conn) {
	ms.db = db
}

func (ms *MemStorage) PingDB() error {
	err := ms.db.Ping(context.Background())
	if err != nil {
		logger.Log.Error("Unable to ping database", zap.String("err", err.Error()))
		return err
	}
	logger.Log.Error("successfully pinged to database")
	return nil
}

func (ms *MemStorage) CloseDB() {
	if ms.db != nil {
		ms.db.Close(context.Background())
	}
}

func (ms *MemStorage) Init() {
	ms.Mu.Lock()
	defer ms.Mu.Unlock()

	ms.Counters = map[string]Counter{
		"PollCount": Counter(0),
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
	}
}
