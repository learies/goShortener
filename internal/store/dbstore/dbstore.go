package dbstore

import (
	"database/sql"
)

type DBStore struct {
	DB *sql.DB
}

func (d *DBStore) Add(shortURL, originalURL string) error {
	return nil
}

func (d *DBStore) Get(shortURL string) (string, error) {
	return "", nil
}

func (d *DBStore) Ping() error {
	return d.DB.Ping()
}
