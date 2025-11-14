package main

import (
	"log"
	"net/http"

	"github.com/ilequeque/musthave-metrics-tpl/internal/handler"
	"github.com/ilequeque/musthave-metrics-tpl/internal/repository"
	"github.com/ilequeque/musthave-metrics-tpl/internal/service"
)

func main() {
	storage := repository.NewMemStorage()
	service := service.NewMetricService(storage)
	handler := handler.NewMetricHandler(service)

	mux := http.NewServeMux()
	mux.HandleFunc("/update/", handler.Update)
	mux.HandleFunc("/update", handler.Update)

	log.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
