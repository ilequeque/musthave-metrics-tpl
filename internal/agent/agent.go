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
			collectRuntime(metrics)                // gauge
			metrics["RandomValue"] = randomGauge() // gauge
			r.pollDelta++                          // считаем ДЕЛЬТУ опросов с прошлого отчёта

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
			log.Printf("metrics sent to %s", r.addr)
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
	url := fmt.Sprintf("%s/update/%s/%s/%v", r.addr, metricType, name, value)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("request create: %w", err)
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}
