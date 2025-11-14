package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ilequeque/musthave-metrics-tpl/internal/handler"
	"github.com/ilequeque/musthave-metrics-tpl/internal/repository"
	"github.com/ilequeque/musthave-metrics-tpl/internal/service"
)

func main() {
	storage := repository.NewMemStorage()
	service := service.NewMetricService(storage)
	handler := handler.NewMetricHandler(service, storage)

	r := chi.NewRouter()

	r.Post("/update/{type}/{name}/{value}", handler.Update)
	r.Get("/value/{type}/{name}", handler.GetValue)
	r.Get("/", handler.GetAllMetrics)

	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
