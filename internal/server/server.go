package server

import (
	"github.com/MaxBoych/MetricsService/internal/metrics"
	"github.com/MaxBoych/MetricsService/pkg/logger"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
)

type Server struct {
	Repo metrics.Repository
}

func NewServer(repo metrics.Repository) *Server {
	return &Server{
		Repo: repo,
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
