// Package store provides database persistence.
// Renamed from 'db' to be more idiomatic Go.
package store

import (
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

// DB wraps a SQL database connection.
type DB struct {
	conn *sql.DB
}

// Open opens a database at the given path.
func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set dialect for goose
	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, fmt.Errorf("failed to set goose dialect: %w", err)
	}

	baseFS, err := fs.Sub(sqlMigrations, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to get base FS: %w", err)
	}
	// Use embedded migrations
	goose.SetBaseFS(baseFS)

	// Run migrations
	if err := goose.Up(conn, "."); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}
