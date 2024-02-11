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

func (o *FileStorage) UpdateGauge(ctx context.Context, name string, new models.Gauge) *models.Gauge {
	value := o.ms.UpdateGauge(ctx, name, new)
	if o.autoSave {
		o.saveOnChange()
	}
	return value
}

func (o *FileStorage) UpdateCounter(ctx context.Context, name string, new models.Counter) *models.Counter {
	value := o.ms.UpdateCounter(ctx, name, new)
	if o.autoSave {
		o.saveOnChange()
	}
	return value
}

func (o *FileStorage) GetGauge(ctx context.Context, name string) *models.Gauge {
	return o.ms.GetGauge(ctx, name)
}

func (o *FileStorage) GetCounter(ctx context.Context, name string) *models.Counter {
	return o.ms.GetCounter(ctx, name)
}

func (o *FileStorage) GetAllMetrics(ctx context.Context) *models.Data {
	return o.ms.GetAllMetrics(ctx)
}
