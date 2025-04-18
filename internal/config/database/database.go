// Package database provides database connection and configuration functionality.
package database

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Connect is a function that connects to the database.
func Connect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := initialize(db); err != nil {
		return nil, err
	}

	return db, nil
}

// initialize is a function that initializes the database.
func initialize(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		uuid UUID PRIMARY KEY,
		short_url VARCHAR(8) NOT NULL UNIQUE,
		original_url TEXT NOT NULL UNIQUE,
		user_id UUID NOT NULL,
		is_deleted BOOLEAN NOT NULL DEFAULT FALSE
	);`

	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}
