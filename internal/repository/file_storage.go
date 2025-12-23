package repository

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/ilequeque/musthave-metrics-tpl/internal/model"
)

type FileStorage struct {
	path     string
	interval time.Duration
	mem      *MemStorage
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

func NewFileStorage(path string, interval time.Duration, restore bool) (*FileStorage, error) {
	mem := NewMemStorage()

	fs := &FileStorage{
		path:     path,
		interval: interval,
		mem:      mem,
		stopCh:   make(chan struct{}),
	}

	if restore {
		if err := fs.LoadFromFile(); err != nil {
			return nil, err
		}
	}

	fs.RunAutosave()
	return fs, nil
}

func (fs *FileStorage) LoadFromFile() error {
	data, err := os.ReadFile(fs.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var metrics []model.Metrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return err
	}

	for _, m := range metrics {
		switch m.MType {
		case model.Gauge:
			if m.Value != nil {
				fs.mem.UpdateGauge(m.ID, *m.Value)
			}
		case model.Counter:
			if m.Delta != nil {
				fs.mem.UpdateCounter(m.ID, *m.Delta)
			}
		}
	}
	return nil
}

func (fs *FileStorage) SaveToFile() error {
	var metrics []model.Metrics

	for name, val := range fs.mem.GetAllGauges() {
		v := val
		metrics = append(metrics, model.Metrics{
			ID:    name,
			MType: model.Gauge,
			Value: &v,
		})
	}

	for name, val := range fs.mem.GetAllCounters() {
		d := val
		metrics = append(metrics, model.Metrics{
			ID:    name,
			MType: model.Counter,
			Delta: &d,
		})
	}

	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(fs.path, data, 0644)
}

func (fs *FileStorage) RunAutosave() {
	if fs.interval == 0 {
		return
	}

	fs.wg.Add(1)
	ticker := time.NewTicker(fs.interval)
	go func() {
		defer fs.wg.Done()
		for {
			select {
			case <-ticker.C:
				_ = fs.SaveToFile()
			case <-fs.stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func (fs *FileStorage) Stop() {
	close(fs.stopCh)
	fs.wg.Wait()
}
func (fs *FileStorage) UpdateGauge(name string, value float64) {
	fs.mem.UpdateGauge(name, value)
	if fs.interval == 0 {
		_ = fs.SaveToFile()
	}
}

func (fs *FileStorage) UpdateCounter(name string, delta int64) {
	fs.mem.UpdateCounter(name, delta)
	if fs.interval == 0 {
		_ = fs.SaveToFile()
	}
}

func (fs *FileStorage) GetGauge(name string) (float64, bool) {
	return fs.mem.GetGauge(name)
}

func (fs *FileStorage) GetCounter(name string) (int64, bool) {
	return fs.mem.GetCounter(name)
}

func (fs *FileStorage) GetAllGauges() map[string]float64 {
	return fs.mem.GetAllGauges()
}

func (fs *FileStorage) GetAllCounters() map[string]int64 {
	return fs.mem.GetAllCounters()
}
