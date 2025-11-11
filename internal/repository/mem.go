package repository

import "sync"

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

type Storage interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, delta int64)
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetAllGauges() map[string]float64
	GetAllCounters() map[string]int64
}
type MemStorage struct {
	mu       sync.Mutex
	gauges   map[string]float64
	counters map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}
func (m *MemStorage) UpdateGauge(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauges[name] = value
}

func (m *MemStorage) UpdateCounter(name string, delta int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name] += delta
}

func (m *MemStorage) GetGauge(name string) (float64, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.gauges[name]
	return v, ok
}

func (m *MemStorage) GetCounter(name string) (int64, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.counters[name]
	return v, ok
}

func (m *MemStorage) GetAllGauges() map[string]float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make(map[string]float64)
	for k, v := range m.gauges {
		res[k] = v
	}
	return res
}

func (m *MemStorage) GetAllCounters() map[string]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make(map[string]int64)
	for k, v := range m.counters {
		res[k] = v
	}
	return res
}
