package server

import (
	"github.com/MaxBoych/MetricsService/config"
	"github.com/MaxBoych/MetricsService/internal/metrics"
	"github.com/MaxBoych/MetricsService/pkg/logger"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
)

type Server struct {
	Repo metrics.Repository
	Cfg  *config.Config
}

func NewServer(repo metrics.Repository, cfg *config.Config) *Server {
	return &Server{
		Repo: repo,
		Cfg:  cfg,
	}
}

func (o *Server) Run(address string) {
	router := chi.NewRouter()
	o.MapHandlers(router)

	logger.Log.Info("Server running", zap.String("address", address))
	if err := http.ListenAndServe(address, router); err != nil {
		panic(err)
	}
}
