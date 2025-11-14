package service

import (
	"errors"
	"strconv"

	"github.com/ilequeque/musthave-metrics-tpl/internal/repository"
)

var (
	ErrUnknownType = errors.New("unknown metric type")
	ErrBadValue    = errors.New("bad metric value")
	ErrNoName      = errors.New("empty metric name")
)

type MetricService struct {
	st repository.Storage
}

func NewMetricService(st *repository.MemStorage) *MetricService {
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
			return ErrBadValue
		}
		s.st.UpdateGauge(name, v)
		return nil
	case repository.Counter:
		delta, err := strconv.ParseInt(rawValue, 10, 64)
		if err != nil {
			return ErrBadValue
		}
		s.st.UpdateCounter(name, delta)
		return nil
	default:
		return ErrUnknownType
	}
}
