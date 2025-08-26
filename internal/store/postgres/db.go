package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type DBOptions struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

// Open connects to Postgres using lib/pq and returns *sql.DB.
func Open(opts DBOptions) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		opts.Host, opts.Port, opts.User, opts.Password, opts.Name, opts.SSLMode,
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql open: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}
	return db, nil
}

// MigrateUp runs all up migrations from database/migrations using golang-migrate.
func MigrateUp(dbOpts DBOptions) error {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		dbOpts.User, dbOpts.Password, dbOpts.Host, dbOpts.Port, dbOpts.Name, dbOpts.SSLMode,
	)
	m, err := migrate.New(
		"file://database/migrations",
		dsn,
	)
	if err != nil {
		return fmt.Errorf("migrate new: %w", err)
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}
