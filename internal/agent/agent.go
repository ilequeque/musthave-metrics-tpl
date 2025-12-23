package agent

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"
)

type Runner struct {
	client       *http.Client
	addr         string
	pollDelta    int64 // сколько опросов с прошлого успешного репорта
	pollTicker   *time.Ticker
	reportTicker *time.Ticker
}

func NewRunner(addr string, poll, report time.Duration) *Runner {
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}
	return &Runner{
		client:       &http.Client{Timeout: 5 * time.Second},
		addr:         addr,
		pollTicker:   time.NewTicker(poll),
		reportTicker: time.NewTicker(report),
	}
}

func (r *Runner) Close() {
	r.pollTicker.Stop()
	r.reportTicker.Stop()
}

func (r *Runner) Run() {
	metrics := make(map[string]float64)

	for {
		select {
		case <-r.pollTicker.C:
			collectRuntime(metrics)
			metrics["RandomValue"] = randomGauge()
			r.pollDelta++

		case <-r.reportTicker.C:
			for name, value := range metrics {
				if err := r.sendMetric("gauge", name, value); err != nil {
					log.Printf("send gauge %s error: %v", name, err)
				}
			}

			if r.pollDelta > 0 {
				if err := r.sendMetric("counter", "PollCount", r.pollDelta); err != nil {
					log.Printf("send counter PollCount error: %v", err)
				} else {
					r.pollDelta = 0
				}
			}
		}
	}
}

func collectRuntime(dst map[string]float64) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	dst["Alloc"] = float64(ms.Alloc)
	dst["BuckHashSys"] = float64(ms.BuckHashSys)
	dst["Frees"] = float64(ms.Frees)
	dst["GCCPUFraction"] = ms.GCCPUFraction
	dst["GCSys"] = float64(ms.GCSys)
	dst["HeapAlloc"] = float64(ms.HeapAlloc)
	dst["HeapIdle"] = float64(ms.HeapIdle)
	dst["HeapInuse"] = float64(ms.HeapInuse)
	dst["HeapObjects"] = float64(ms.HeapObjects)
	dst["HeapReleased"] = float64(ms.HeapReleased)
	dst["HeapSys"] = float64(ms.HeapSys)
	dst["LastGC"] = float64(ms.LastGC)
	dst["Lookups"] = float64(ms.Lookups)
	dst["MCacheInuse"] = float64(ms.MCacheInuse)
	dst["MCacheSys"] = float64(ms.MCacheSys)
	dst["MSpanInuse"] = float64(ms.MSpanInuse)
	dst["MSpanSys"] = float64(ms.MSpanSys)
	dst["Mallocs"] = float64(ms.Mallocs)
	dst["NextGC"] = float64(ms.NextGC)
	dst["NumForcedGC"] = float64(ms.NumForcedGC)
	dst["NumGC"] = float64(ms.NumGC)
	dst["OtherSys"] = float64(ms.OtherSys)
	dst["PauseTotalNs"] = float64(ms.PauseTotalNs)
	dst["StackInuse"] = float64(ms.StackInuse)
	dst["StackSys"] = float64(ms.StackSys)
	dst["Sys"] = float64(ms.Sys)
	dst["TotalAlloc"] = float64(ms.TotalAlloc)
}

func randomGauge() float64 { return float64(time.Now().UnixNano()%1_000_000) / 1_000_000 }

func (r *Runner) sendMetric(metricType, name string, value any) error {
	url := fmt.Sprintf("%s/update", r.addr)

	var metric struct {
		ID    string   `json:"id"`
		MType string   `json:"type"`
		Delta *int64   `json:"delta,omitempty"`
		Value *float64 `json:"value,omitempty"`
	}

	metric.ID = name
	metric.MType = metricType

	switch metricType {
	case "gauge":
		v := toFloat64(value)
		metric.Value = &v
	case "counter":
		d := toInt64(value)
		metric.Delta = &d
	default:
		return fmt.Errorf("unknown metric type: %s", metricType)
	}

	if err := sendJSONGzip(r.client, url, metric); err != nil {
		return fmt.Errorf("send gzip json: %w", err)
	}

	return nil
}

func toFloat64(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int64:
		return float64(t)
	case int:
		return float64(t)
	default:
		return 0
	}
}

func toInt64(v any) int64 {
	switch t := v.(type) {
	case int64:
		return t
	case int:
		return int64(t)
	case float64:
		return int64(t)
	default:
		return 0
	}
}
