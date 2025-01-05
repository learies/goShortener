package dbstore

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/learies/goShortener/internal/models"
)

type DBStore struct {
	DB *sql.DB
}

func (d *DBStore) Add(ctx context.Context, shortURL, originalURL string) error {
	record := models.ShortenStore{
		UUID:        uuid.New(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	query := `INSERT INTO urls (uuid, short_url, original_url) VALUES ($1, $2, $3)`
	_, err := d.DB.ExecContext(ctx, query, record.UUID, record.ShortURL, record.OriginalURL)
	if err != nil {
		return err
	}

	return nil
}

func (d *DBStore) Get(ctx context.Context, shortURL string) (string, error) {
	query := `SELECT original_url FROM urls WHERE short_url = $1`

	var originalURL string
	err := d.DB.QueryRowContext(ctx, query, shortURL).Scan(&originalURL)
	if err != nil {
		return "", err
	}

	return originalURL, nil
}

func (d *DBStore) Ping() error {
	return d.DB.Ping()
}
