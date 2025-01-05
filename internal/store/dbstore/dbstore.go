package dbstore

import (
	"database/sql"

	"github.com/google/uuid"

	"github.com/learies/goShortener/internal/models"
)

type DBStore struct {
	DB *sql.DB
}

func (d *DBStore) Add(shortURL, originalURL string) error {
	record := models.ShortenStore{
		UUID:        uuid.New(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	query := `INSERT INTO urls (uuid, short_url, original_url) VALUES ($1, $2, $3)`
	_, err := d.DB.Exec(query, record.UUID, record.ShortURL, record.OriginalURL)
	if err != nil {
		return err
	}

	return nil
}

func (d *DBStore) Get(shortURL string) (string, error) {
	query := `SELECT original_url FROM urls WHERE short_url = $1`

	var originalURL string
	err := d.DB.QueryRow(query, shortURL).Scan(&originalURL)
	if err != nil {
		return "", err
	}

	return originalURL, nil
}

func (d *DBStore) Ping() error {
	return d.DB.Ping()
}
