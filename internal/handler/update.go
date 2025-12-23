package handler

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ilequeque/musthave-metrics-tpl/internal/model"
	"github.com/ilequeque/musthave-metrics-tpl/internal/repository"
	"github.com/ilequeque/musthave-metrics-tpl/internal/service"
)

type MetricHandler struct {
	svc service.MetricService
	st  repository.Storage
	db  DBPinger
}

func NewMetricHandler(
	svc *service.MetricService,
	st repository.Storage,
	db ...DBPinger,
) *MetricHandler {
	var pinger DBPinger
	if len(db) > 0 {
		pinger = db[0]
	}

	return &MetricHandler{
		svc: *svc,
		st:  st,
		db:  pinger,
	}
}

type DBPinger interface {
	Ping() error
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

func (h *MetricHandler) PingDB(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		http.Error(w, "database not configured", http.StatusInternalServerError)
		return
	}

	if err := h.db.Ping(); err != nil {
		http.Error(w, "database not reachable", http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if err := t.Execute(w, data); err != nil {
		log.Printf("template execute error: %v", err)
		http.Error(w, "internal template render error", http.StatusInternalServerError)
		return
	}
}

func (h *MetricHandler) UpdateJSON(w http.ResponseWriter, r *http.Request) {
	var reader io.Reader = r.Body

	if r.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, "invalid gzip", http.StatusBadRequest)
			return
		}
		defer gz.Close()
		reader = gz
	}

	var m model.Metrics
	if err := json.NewDecoder(reader).Decode(&m); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	switch m.MType {
	case model.Gauge:
		if m.Value == nil {
			http.Error(w, "missing value for gauge", http.StatusBadRequest)
			return
		}
		h.st.UpdateGauge(m.ID, *m.Value)

	case model.Counter:
		if m.Delta == nil {
			http.Error(w, "missing delta for counter", http.StatusBadRequest)
			return
		}
		h.st.UpdateCounter(m.ID, *m.Delta)

	default:
		http.Error(w, "unknown metric type", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *MetricHandler) GetValueJSON(w http.ResponseWriter, r *http.Request) {
	var m model.Metrics
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	switch m.MType {
	case model.Gauge:
		if val, ok := h.st.GetGauge(m.ID); ok {
			m.Value = &val
		} else {
			http.NotFound(w, r)
			return
		}
	case model.Counter:
		if val, ok := h.st.GetCounter(m.ID); ok {
			m.Delta = &val
		} else {
			http.NotFound(w, r)
			return
		}
	default:
		http.Error(w, "unknown metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(m)
}
