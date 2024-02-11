package server

import (
	"github.com/MaxBoych/MetricsService/internal/metrics/delivery"
	"github.com/MaxBoych/MetricsService/internal/metrics/usecase"
	"github.com/go-chi/chi/v5"
)

func (o *Server) MapHandlers(router *chi.Mux) {
	useCase := usecase.NewMetricsUseCase(o.Repo)
	msHandler := delivery.NewMetricsHandler(useCase)

	delivery.SetupRoutes(router, msHandler)
}
