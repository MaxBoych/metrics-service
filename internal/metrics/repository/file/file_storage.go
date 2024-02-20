package file

import (
	"context"
	"encoding/json"
	"github.com/MaxBoych/MetricsService/internal/metrics/models"
	"github.com/MaxBoych/MetricsService/internal/metrics/repository/memory"
	"github.com/MaxBoych/MetricsService/pkg/logger"
	"go.uber.org/zap"
	"os"
	"sync"
)

type FileStorage struct {
	ms       *memory.MemStorage
	Mu       sync.RWMutex
	filePath string
	autoSave bool
}

func NewFileStorage(ms *memory.MemStorage) *FileStorage {
	return &FileStorage{
		ms: ms,
	}
}

func (o *FileStorage) SetConfigValues(filePath string, autoSave bool) {
	o.filePath = filePath
	o.autoSave = autoSave
}

func (o *FileStorage) CreateFileIfNotExists() error {
	o.Mu.Lock()
	defer o.Mu.Unlock()

	_, err := os.Stat(o.filePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if os.IsNotExist(err) {
		file, err := os.Create(o.filePath)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	return nil
}

func (o *FileStorage) LoadFromFile() error {
	o.Mu.Lock()
	defer o.Mu.Unlock()

	data, err := os.ReadFile(o.filePath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, o.ms.Data); err != nil {
		return err
	}
	return nil
}

func (o *FileStorage) saveOnChange() {
	err := o.StoreToFile()
	if err != nil {
		logger.Log.Info("ERROR store to file: saveOnChange()", zap.String("error", err.Error()))
	}
}

func (o *FileStorage) StoreToFile() error {
	o.Mu.Lock()
	defer o.Mu.Unlock()

	data, err := json.MarshalIndent(o.ms.Data, "", "   ")
	if err != nil {
		return err
	}
	return os.WriteFile(o.filePath, data, 0666)
}

func (o *FileStorage) UpdateGauge(ctx context.Context, m models.Metrics) (*models.Metrics, error) {
	_, err := o.ms.UpdateGauge(ctx, m)
	if o.autoSave {
		o.saveOnChange()
	}
	return &m, err
}

func (o *FileStorage) UpdateCounter(ctx context.Context, m models.Metrics) (*models.Metrics, error) {
	_, err := o.ms.UpdateCounter(ctx, m)
	if o.autoSave {
		o.saveOnChange()
	}
	return &m, err
}

func (o *FileStorage) GetGauge(ctx context.Context, name string) (*models.Gauge, error) {
	return o.ms.GetGauge(ctx, name)
}

func (o *FileStorage) GetCounter(ctx context.Context, name string) (*models.Counter, error) {
	return o.ms.GetCounter(ctx, name)
}

func (o *FileStorage) GetAll(ctx context.Context) (*models.Data, error) {
	return o.ms.GetAll(ctx)
}

func (o *FileStorage) UpdateMany(ctx context.Context, ms []models.Metrics) ([]models.Metrics, error) {
	for _, m := range ms {
		if m.MType == models.GaugeMetricName {
			_, _ = o.UpdateGauge(ctx, m)
		} else if m.MType == models.CounterMetricName {
			_, _ = o.UpdateCounter(ctx, m)
		}
	}
	return ms, nil
}
