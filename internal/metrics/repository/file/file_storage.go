package file

import (
	"encoding/json"
	"github.com/MaxBoych/MetricsService/internal/repository/memory"
	"github.com/MaxBoych/MetricsService/pkg/logger"
	"go.uber.org/zap"
	"os"
	"sync"
)

type FileStorage struct {
	ms       *memory.MemStorage
	Mu       sync.RWMutex
	FilePath string
}

func NewFileStorage(ms *memory.MemStorage) *FileStorage {
	return &FileStorage{
		ms: ms,
	}
}

func (o *FileStorage) SetConfigValues(filePath string, autoSave bool) {
	o.FilePath = filePath

	if autoSave {
		o.ms.SetOnChange(o.saveOnChange)
	}
}

func (o *FileStorage) LoadFromFile() error {
	o.ms.Mu.Lock()
	defer o.ms.Mu.Unlock()

	data, err := os.ReadFile(o.FilePath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, o.ms.Data); err != nil {
		return err
	}
	return nil
}

func (o *FileStorage) saveOnChange() {
	err := o.storeToFile()
	if err != nil {
		logger.Log.Info("ERROR store to file", zap.String("error", err.Error()))
	}
}

func (o *FileStorage) StoreToFile() error {
	o.Mu.Lock()
	defer o.Mu.Unlock()
	return o.storeToFile()
}

func (o *FileStorage) storeToFile() error {
	data, err := json.MarshalIndent(o.ms.Data, "", "   ")
	if err != nil {
		return err
	}
	return os.WriteFile(o.FilePath, data, 0666)
}
