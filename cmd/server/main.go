package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ilequeque/musthave-metrics-tpl/internal/handler"
	"github.com/ilequeque/musthave-metrics-tpl/internal/repository"
	"github.com/ilequeque/musthave-metrics-tpl/internal/service"
)

func main() {
	addr := flag.String("a", "localhost:8080", "address for HTTP server")
	flag.Parse()

	storage := repository.NewMemStorage()
	service := service.NewMetricService(storage)
	handler := handler.NewMetricHandler(service, storage)

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", handler.Update)
	r.Get("/value/{type}/{name}", handler.GetValue)
	r.Get("/", handler.GetAllMetrics)

	fmt.Printf("Server running on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, r))
}
