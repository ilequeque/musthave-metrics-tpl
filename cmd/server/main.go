package main

import (
	"log"
	"net/http"
	"time"

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
	fs := repository.NewFileStorage(st, cfg.FileStorage, time.Duration(cfg.StoreInterval)*time.Second)

	if cfg.Restore {
		if err := fs.LoadFromFile(); err != nil {
			logrus.Warnf("restore failed: %v", err)
		}
	}

	if cfg.StoreInterval > 0 {
		fs.RunAutosave()
		defer fs.Stop()
	}

	svc := service.NewMetricService(st)
	h := handler.NewMetricHandler(svc, st)
	
	r := chi.NewRouter()

	r.Use(middleware.LoggerMiddleware)
	r.Use(middleware.StripTrailingSlash)
	r.Use(middleware.GzipMiddleware)

	r.Post("/update/{type}/{name}/{value}", h.Update)
	r.Get("/value/{type}/{name}", h.GetValue)
	r.Get("/", h.GetAllMetrics)
	r.Post("/update", h.UpdateJSON)
	r.Post("/value", h.GetValueJSON)

	logrus.Infof("server running on %s", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, r))
}
