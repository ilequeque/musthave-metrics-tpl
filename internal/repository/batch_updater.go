package repository

import "github.com/ilequeque/musthave-metrics-tpl/internal/model"

type BatchUpdater interface {
	UpdateBatch([]model.Metrics) error
}
