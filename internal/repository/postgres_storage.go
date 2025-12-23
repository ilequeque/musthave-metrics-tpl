package repository

import (
	"database/sql"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) (*PostgresStorage, error) {
	s := &PostgresStorage{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}
func (s *PostgresStorage) migrate() error {
	const q = `
	CREATE TABLE IF NOT EXISTS metrics (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		gauge DOUBLE PRECISION,
		counter BIGINT
	);`
	_, err := s.db.Exec(q)
	return err
}
func (s *PostgresStorage) UpdateGauge(name string, value float64) {
	const q = `
	INSERT INTO metrics (id, type, gauge)
	VALUES ($1, 'gauge', $2)
	ON CONFLICT (id)
	DO UPDATE SET gauge = EXCLUDED.gauge;`

	_, _ = s.db.Exec(q, name, value)
}
func (s *PostgresStorage) UpdateCounter(name string, delta int64) {
	const q = `
	INSERT INTO metrics (id, type, counter)
	VALUES ($1, 'counter', $2)
	ON CONFLICT (id)
	DO UPDATE SET counter = metrics.counter + EXCLUDED.counter;`

	_, _ = s.db.Exec(q, name, delta)
}
func (s *PostgresStorage) GetGauge(name string) (float64, bool) {
	var v sql.NullFloat64
	err := s.db.QueryRow(
		`SELECT gauge FROM metrics WHERE id=$1 AND type='gauge'`,
		name,
	).Scan(&v)

	return v.Float64, err == nil && v.Valid
}
func (s *PostgresStorage) GetCounter(name string) (int64, bool) {
	var v sql.NullInt64
	err := s.db.QueryRow(
		`SELECT counter FROM metrics WHERE id=$1 AND type='counter'`,
		name,
	).Scan(&v)

	return v.Int64, err == nil && v.Valid
}
func (ps *PostgresStorage) GetAllGauges() map[string]float64 {
	rows, err := ps.db.Query(`SELECT id, value FROM gauges`)
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
	return result
}

func (ps *PostgresStorage) GetAllCounters() map[string]int64 {
	rows, err := ps.db.Query(`SELECT id, delta FROM counters`)
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
	return result
}
