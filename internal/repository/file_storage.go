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

func NewFileStorage(mem *MemStorage, path string, interval time.Duration) *FileStorage {
	return &FileStorage{
		path:     path,
		interval: interval,
		mem:      mem,
		stopCh:   make(chan struct{}),
	}
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
