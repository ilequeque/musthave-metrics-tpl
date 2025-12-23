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

	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logrus.SetLevel(logrus.InfoLevel)

	var pdb *repository.PostgresDB
	if cfg.DatabaseDSN != "" {
		db, err := repository.NewPostgresDB(cfg.DatabaseDSN)
		if err != nil {
			logrus.Warnf("database not available: %v", err)
		} else {
			pdb = db
			logrus.Info("database connected")
			defer func() { _ = pdb.Close() }()
		}
	}

	var st repository.Storage

	if pdb != nil {
		ps, err := repository.NewPostgresStorage(pdb.DB)
		if err != nil {
			logrus.Fatalf("failed to init postgres storage: %v", err)
		}
		st = ps
		logrus.Info("using postgres storage")

	} else if cfg.FileStorage != "" {
		fs, err := repository.NewFileStorage(
			cfg.FileStorage,
			time.Duration(cfg.StoreInterval)*time.Second,
			cfg.Restore,
		)
		if err != nil {
			logrus.Fatalf("failed to init file storage: %v", err)
		}
		st = fs
		logrus.Info("using file storage")

	} else {
		st = repository.NewMemStorage()
		logrus.Info("using memory storage")
	}

	svc := service.NewMetricService(st)
	h := handler.NewMetricHandler(svc, st, pdb)

	r := chi.NewRouter()
	r.Use(middleware.LoggerMiddleware)
	r.Use(middleware.GzipMiddleware)

	r.Post("/update/{type}/{name}/{value}", h.Update)
	r.Get("/value/{type}/{name}", h.GetValue)
	r.Get("/", h.GetAllMetrics)
	r.Post("/update", h.UpdateJSON)
	r.Post("/update/", h.UpdateJSON)

	r.Post("/value", h.GetValueJSON)
	r.Post("/value/", h.GetValueJSON)

	r.Get("/ping", h.PingDB)

	logrus.Infof("server running on %s", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, r))
}
