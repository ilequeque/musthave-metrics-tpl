package service

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/ilequeque/musthave-metrics-tpl/internal/repository"
)

var (
	ErrUnknownType = errors.New("unknown metric type")
	ErrBadValue    = errors.New("bad metric value")
	ErrNoName      = errors.New("no metric name")
)

type MetricService struct {
	st repository.Storage
}

func NewMetricService(st repository.Storage) *MetricService {
	return &MetricService{st: st}
}

func (s *MetricService) Update(mtype, name, rawValue string) error {
	if name == "" {
		return ErrNoName
	}

	switch repository.MetricType(mtype) {
	case repository.Gauge:
		v, err := strconv.ParseFloat(rawValue, 64)
		if err != nil {
			return fmt.Errorf("invalid gauge value %q: %w", rawValue, ErrBadValue)
		}
		s.st.UpdateGauge(name, v)
		return nil

	case repository.Counter:
		delta, err := strconv.ParseInt(rawValue, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid counter value %q: %w", rawValue, ErrBadValue)
		}
		s.st.UpdateCounter(name, delta)
		return nil

	default:
		return ErrUnknownType
	}
}
