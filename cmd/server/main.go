package main

import (
	"fmt"
	"github.com/MaxBoych/MetricsService/config"
	"github.com/MaxBoych/MetricsService/internal/metrics"
	"github.com/MaxBoych/MetricsService/internal/metrics/delivery"
	"github.com/MaxBoych/MetricsService/internal/metrics/repository/memory"
	"github.com/MaxBoych/MetricsService/internal/server"
	"github.com/MaxBoych/MetricsService/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	setupLogger()
	//defer db.Close()
	setupServer()
}

func setupLogger() (ms *memory.MemStorage, msHandler *delivery.MetricsHandler) {
	if err := logger.Initialize("INFO"); err != nil {
		fmt.Printf("logger init error: %v\n", err)
	}
	return
}

func setupServer() {
	// setup config
	cfg := config.NewConfig()
	cfg.ParseConfig()
	logger.Log.Info("PARSE DATA", zap.String("data", cfg.String()))

	var repo metrics.Repository

	// 3-rd priority
	ms := cfg.ConfigureMS()
	repo = ms

	// 2-nd priority
	fs := cfg.ConfigureFS(ms)
	if fs != nil {
		repo = fs
	}

	// 1-st priority
	db := cfg.ConfigureDB()
	if db != nil {
		defer db.Close()
		repo = db
	}

	// setup server
	srv := server.NewServer(repo)
	srv.Run(cfg.RunAddr)
}
