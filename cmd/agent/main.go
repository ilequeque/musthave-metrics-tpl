package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strings"
	"time"
)

func main() {
	addr := flag.String("a", "http://localhost:8080", "server address")
	reportInt := flag.Int("r", 10, "report interval in seconds")
	pollInt := flag.Int("p", 2, "poll interval in seconds")
	flag.Parse()

	if !strings.HasPrefix(*addr, "http://") && !strings.HasPrefix(*addr, "https://") {
		*addr = "http://" + *addr
	}

	client := &http.Client{Timeout: 5 * time.Second}

	pollInterval := time.Duration(*pollInt) * time.Second
	reportInterval := time.Duration(*reportInt) * time.Second

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
				sendMetric(client, *addr, "gauge", name, value)
			}
			sendMetric(client, *addr, "counter", "PollCount", pollCount)
			fmt.Println("metrics sent to", *addr)
		}
	}
}

func collectMetrics() map[string]float64 {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	return map[string]float64{
		"Alloc":         float64(ms.Alloc),
		"TotalAlloc":    float64(ms.TotalAlloc),
		"Sys":           float64(ms.Sys),
		"Lookups":       float64(ms.Lookups),
		"Mallocs":       float64(ms.Mallocs),
		"Frees":         float64(ms.Frees),
		"HeapAlloc":     float64(ms.HeapAlloc),
		"HeapSys":       float64(ms.HeapSys),
		"HeapIdle":      float64(ms.HeapIdle),
		"HeapInuse":     float64(ms.HeapInuse),
		"HeapReleased":  float64(ms.HeapReleased),
		"HeapObjects":   float64(ms.HeapObjects),
		"StackInuse":    float64(ms.StackInuse),
		"StackSys":      float64(ms.StackSys),
		"MSpanInuse":    float64(ms.MSpanInuse),
		"MSpanSys":      float64(ms.MSpanSys),
		"MCacheInuse":   float64(ms.MCacheInuse),
		"MCacheSys":     float64(ms.MCacheSys),
		"BuckHashSys":   float64(ms.BuckHashSys),
		"GCSys":         float64(ms.GCSys),
		"OtherSys":      float64(ms.OtherSys),
		"NextGC":        float64(ms.NextGC),
		"LastGC":        float64(ms.LastGC),
		"PauseTotalNs":  float64(ms.PauseTotalNs),
		"NumGC":         float64(ms.NumGC),
		"NumForcedGC":   float64(ms.NumForcedGC),
		"GCCPUFraction": ms.GCCPUFraction,
	}
}

func sendMetric(client *http.Client, addr, metricType, name string, value interface{}) {
	url := fmt.Sprintf("%s/update/%s/%s/%v", addr, metricType, name, value)
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
