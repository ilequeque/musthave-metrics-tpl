package repository

import (
	"database/sql"
	"errors"

	"github.com/ilequeque/musthave-metrics-tpl/internal/model"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) (*PostgresStorage, error) {
	ps := &PostgresStorage{db: db}
	if err := ps.migrate(); err != nil {
		return nil, err
	}
	return ps, nil
}

func (ps *PostgresStorage) migrate() error {
	const q = `
	CREATE TABLE IF NOT EXISTS metrics (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		gauge DOUBLE PRECISION,
		counter BIGINT
	);`
	_, err := ps.db.Exec(q)
	return err
}

func (ps *PostgresStorage) UpdateGauge(name string, value float64) {
	const q = `
	INSERT INTO metrics (id, type, gauge)
	VALUES ($1, 'gauge', $2)
	ON CONFLICT (id)
	DO UPDATE SET gauge = EXCLUDED.gauge;`

	_ = retry(func() error {
		_, err := ps.db.Exec(q, name, value)
		if isRetriablePGError(err) {
			return err
		}
		return nil
	})
}

func (ps *PostgresStorage) UpdateCounter(name string, delta int64) {
	const q = `
	INSERT INTO metrics (id, type, counter)
	VALUES ($1, 'counter', $2)
	ON CONFLICT (id)
	DO UPDATE SET counter = metrics.counter + EXCLUDED.counter;`
	_ = retry(func() error {
		_, err := ps.db.Exec(q, name, delta)
		if isRetriablePGError(err) {
			return err
		}
		return nil
	})
}

func (ps *PostgresStorage) GetGauge(name string) (float64, bool) {
	var v sql.NullFloat64
	err := ps.db.QueryRow(
		`SELECT gauge FROM metrics WHERE id=$1 AND type='gauge'`,
		name,
	).Scan(&v)

	return v.Float64, err == nil && v.Valid
}

func (ps *PostgresStorage) GetCounter(name string) (int64, bool) {
	var v sql.NullInt64
	err := ps.db.QueryRow(
		`SELECT counter FROM metrics WHERE id=$1 AND type='counter'`,
		name,
	).Scan(&v)

	return v.Int64, err == nil && v.Valid
}

func (ps *PostgresStorage) GetAllGauges() map[string]float64 {
	rows, err := ps.db.Query(
		`SELECT id, gauge FROM metrics WHERE type='gauge'`,
	)
	if err != nil {
		return map[string]float64{}
	}
	defer rows.Close()

	result := make(map[string]float64)
	for rows.Next() {
		var id string
		var val float64
		if err := rows.Scan(&id, &val); err == nil {
			result[id] = val
		}
	}

	if err := rows.Err(); err != nil {
		return map[string]float64{}
	}

	return result
}

func (ps *PostgresStorage) GetAllCounters() map[string]int64 {
	rows, err := ps.db.Query(
		`SELECT id, counter FROM metrics WHERE type='counter'`,
	)
	if err != nil {
		return map[string]int64{}
	}
	defer rows.Close()

	result := make(map[string]int64)
	for rows.Next() {
		var id string
		var val int64
		if err := rows.Scan(&id, &val); err == nil {
			result[id] = val
		}
	}

	if err := rows.Err(); err != nil {
		return map[string]int64{}
	}

	return result
}

func (s *PostgresStorage) UpdateBatch(metrics []model.Metrics) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, m := range metrics {
		switch m.MType {
		case model.Gauge:
			if m.Value == nil {
				continue
			}
			_, err = tx.Exec(`
				INSERT INTO metrics (id, type, gauge)
				VALUES ($1, 'gauge', $2)
				ON CONFLICT (id)
				DO UPDATE SET gauge = EXCLUDED.gauge
			`, m.ID, *m.Value)

		case model.Counter:
			if m.Delta == nil {
				continue
			}
			_, err = tx.Exec(`
				INSERT INTO metrics (id, type, counter)
				VALUES ($1, 'counter', $2)
				ON CONFLICT (id)
				DO UPDATE SET counter = metrics.counter + EXCLUDED.counter
			`, m.ID, *m.Delta)
		}

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
func isRetriablePGError(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pgerrcode.IsConnectionException(pqErr.Code)
	}
	return false
}
