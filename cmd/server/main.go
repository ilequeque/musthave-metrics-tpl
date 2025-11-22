package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ilequeque/musthave-metrics-tpl/internal/config"
	"github.com/ilequeque/musthave-metrics-tpl/internal/handler"
	"github.com/ilequeque/musthave-metrics-tpl/internal/middleware"
	"github.com/ilequeque/musthave-metrics-tpl/internal/repository"
	"github.com/ilequeque/musthave-metrics-tpl/internal/service"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg := config.ParseServerFlags()

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.InfoLevel)

	st := repository.NewMemStorage()
	svc := service.NewMetricService(st)
	h := handler.NewMetricHandler(svc, st)

	r := chi.NewRouter()

	r.Use(middleware.LoggerMiddleware)

	r.Post("/update/{type}/{name}/{value}", h.Update)
	r.Get("/value/{type}/{name}", h.GetValue)
	r.Get("/", h.GetAllMetrics)
	r.Post("/update", h.UpdateJSON)
	r.Post("/value", h.GetValueJSON)

	logrus.Infof("server running on %s", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, r))
}
