package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/ilequeque/musthave-metrics-tpl/internal/repository"
	"github.com/ilequeque/musthave-metrics-tpl/internal/service"
)

func addChiURLParams(r *http.Request, params map[string]string) *http.Request {
	routeCtx := chi.NewRouteContext()
	for k, v := range params {
		routeCtx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
}

func TestUpdateHandler_OK(t *testing.T) {
	st := repository.NewMemStorage()
	svc := service.NewMetricService(st)
	h := NewMetricHandler(svc, st)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/testMetric/42.5", nil)
	req = addChiURLParams(req, map[string]string{
		"type":  "gauge",
		"name":  "testMetric",
		"value": "42.5",
	})

	w := httptest.NewRecorder()
	h.Update(w, req)

	res := w.Result()
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", res.StatusCode)
	}

	if val, ok := st.GetGauge("testMetric"); !ok || val != 42.5 {
		t.Fatalf("expected testMetric=42.5, got %v (ok=%v)", val, ok)
	}
}

func TestGetValueHandler_OK(t *testing.T) {
	st := repository.NewMemStorage()
	svc := service.NewMetricService(st)
	h := NewMetricHandler(svc, st)

	st.UpdateGauge("HeapAlloc", 123.45)

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/HeapAlloc", nil)
	req = addChiURLParams(req, map[string]string{
		"type": "gauge",
		"name": "HeapAlloc",
	})

	w := httptest.NewRecorder()
	h.GetValue(w, req)

	res := w.Result()
	defer func() { _ = res.Body.Close() }()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", res.StatusCode, string(body))
	}

	if !strings.Contains(string(body), "123.45") {
		t.Fatalf("unexpected body: %s", body)
	}
}
