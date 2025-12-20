package repository

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
)

type PostgresDB struct {
	DB *sql.DB
}

func NewPostgresDB(dsn string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("db.Ping: %w", err)
	}

	return &PostgresDB{DB: db}, nil
}
func (p *PostgresDB) Ping() error {
	if p == nil || p.DB == nil {
		return errors.New("database not initialized")
	}
	return p.DB.Ping()
}

func (p *PostgresDB) Close() error {
	return p.DB.Close()
}
