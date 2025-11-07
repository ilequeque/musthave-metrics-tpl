package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

const (
	serverAddr     = "http://localhost:8080"
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
)

func main() {
	client := &http.Client{Timeout: 5 * time.Second}
	rand.Seed(time.Now().UnixNano())

	var pollCount int64
	metrics := make(map[string]float64)

	tPoll := time.NewTicker(pollInterval)
	tReport := time.NewTicker(reportInterval)
	defer tPoll.Stop()
	defer tReport.Stop()

	for {
		select {
		case <-tPoll.C:
			collected := collectMetrics()
			for k, v := range collected {
				metrics[k] = v
			}
			metrics["RandomValue"] = rand.Float64()
			pollCount++

		case <-tReport.C:
			for name, value := range metrics {
				sendMetric(client, "gauge", name, value)
			}
			sendMetric(client, "counter", "PollCount", pollCount)
			fmt.Println("metrics sent to server")
		}
	}
}

func collectMetrics() map[string]float64 {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	return map[string]float64{
		"Alloc":         float64(ms.Alloc),
		"BuckHashSys":   float64(ms.BuckHashSys),
		"Frees":         float64(ms.Frees),
		"GCCPUFraction": ms.GCCPUFraction,
		"GCSys":         float64(ms.GCSys),
		"HeapAlloc":     float64(ms.HeapAlloc),
		"HeapIdle":      float64(ms.HeapIdle),
		"HeapInuse":     float64(ms.HeapInuse),
		"HeapObjects":   float64(ms.HeapObjects),
		"HeapReleased":  float64(ms.HeapReleased),
		"HeapSys":       float64(ms.HeapSys),
		"LastGC":        float64(ms.LastGC),
		"Lookups":       float64(ms.Lookups),
		"MCacheInuse":   float64(ms.MCacheInuse),
		"MCacheSys":     float64(ms.MCacheSys),
		"MSpanInuse":    float64(ms.MSpanInuse),
		"MSpanSys":      float64(ms.MSpanSys),
		"Mallocs":       float64(ms.Mallocs),
		"NextGC":        float64(ms.NextGC),
		"NumForcedGC":   float64(ms.NumForcedGC),
		"NumGC":         float64(ms.NumGC),
		"OtherSys":      float64(ms.OtherSys),
		"PauseTotalNs":  float64(ms.PauseTotalNs),
		"StackInuse":    float64(ms.StackInuse),
		"StackSys":      float64(ms.StackSys),
		"Sys":           float64(ms.Sys),
		"TotalAlloc":    float64(ms.TotalAlloc),
	}
}

func sendMetric(client *http.Client, metricType, name string, value interface{}) {
	url := fmt.Sprintf("%s/update/%s/%s/%v", serverAddr, metricType, name, value)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		fmt.Println("request create error:", err)
		return
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("send error:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("server returned %d for %s\n", resp.StatusCode, name)
	}
}
