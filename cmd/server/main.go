package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ilequeque/musthave-metrics-tpl/internal/config"
	"github.com/ilequeque/musthave-metrics-tpl/internal/handler"
	"github.com/ilequeque/musthave-metrics-tpl/internal/repository"
	"github.com/ilequeque/musthave-metrics-tpl/internal/service"
)

func main() {
	cfg := config.ParseServerFlags()

	st := repository.NewMemStorage()
	svc := service.NewMetricService(st)
	h := handler.NewMetricHandler(svc, st)

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", h.Update)
	r.Get("/value/{type}/{name}", h.GetValue)
	r.Get("/", h.GetAllMetrics)

	log.Printf("server running on %s", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, r))
}
