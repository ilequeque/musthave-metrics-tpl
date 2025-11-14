package handler

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ilequeque/musthave-metrics-tpl/internal/repository"
	"github.com/ilequeque/musthave-metrics-tpl/internal/service"
)

type MetricHandler struct {
	svc service.MetricService
	st  repository.Storage
}

func NewMetricHandler(svc *service.MetricService, st *repository.MemStorage) *MetricHandler {
	return &MetricHandler{svc: *svc, st: st}
}

func (h *MetricHandler) Update(w http.ResponseWriter, r *http.Request) {
	mtype := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")

	err := h.svc.Update(mtype, name, value)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNoName):
			http.NotFound(w, r)
		case errors.Is(err, service.ErrUnknownType), errors.Is(err, service.ErrBadValue):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			log.Printf("internal error on Update: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *MetricHandler) GetValue(w http.ResponseWriter, r *http.Request) {
	mtype := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	switch repository.MetricType(mtype) {
	case repository.Gauge:
		if val, ok := h.st.GetGauge(name); ok {
			w.WriteHeader(http.StatusOK)
			str := strconv.FormatFloat(val, 'f', -1, 64)
			fmt.Fprint(w, str)
		} else {
			http.NotFound(w, r)
		}
	case repository.Counter:
		if val, ok := h.st.GetCounter(name); ok {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%d", val)
		} else {
			http.NotFound(w, r)
		}
	default:
		http.Error(w, "unknown metric type", http.StatusBadRequest)
		return
	}
}

func (h *MetricHandler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	type metricView struct {
		Name  string
		Type  string
		Value string
	}

	var data []metricView

	for name, value := range h.st.GetAllGauges() {
		data = append(data, metricView{
			Name:  name,
			Type:  "gauge",
			Value: strconv.FormatFloat(value, 'f', 6, 64),
		})
	}

	for name, value := range h.st.GetAllCounters() {
		data = append(data, metricView{
			Name:  name,
			Type:  "counter",
			Value: fmt.Sprintf("%d", value),
		})
	}

	tmpl := `
	<!DOCTYPE html>
	<html>
	<head><meta charset="UTF-8"><title>Metrics</title></head>
	<body>
	<h2>Known Metrics</h2>
	<table border="1" cellpadding="5">
	<tr><th>Type</th><th>Name</th><th>Value</th></tr>
	{{range .}}
	<tr><td>{{.Type}}</td><td>{{.Name}}</td><td>{{.Value}}</td></tr>
	{{end}}
	</table>
	</body>
	</html>
	`

	t, err := template.New("metrics").Parse(tmpl)
	if err != nil {
		log.Printf("template parse error: %v", err)
		http.Error(w, "internal template error", http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		log.Printf("template execute error: %v", err)
		http.Error(w, "internal template render error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
