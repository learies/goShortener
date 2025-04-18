// Package dbstore provides database-based storage implementation for the URL shortener service.
package dbstore

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/learies/goShortener/internal/config/logger"
	"github.com/learies/goShortener/internal/models"
	"github.com/learies/goShortener/internal/store/filestore"
)

// DBStore is a struct that represents the database store.
type DBStore struct {
	DB *sql.DB
}

// Add is a method that adds a new URL to the database.
func (d *DBStore) Add(ctx context.Context, shortURL, originalURL string, userID uuid.UUID) error {
	record := models.ShortenStore{
		UUID:        uuid.New(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UserID:      userID,
	}

	query := `INSERT INTO urls (uuid, short_url, original_url, user_id) VALUES ($1, $2, $3, $4)`
	_, err := d.DB.ExecContext(ctx, query, record.UUID, record.ShortURL, record.OriginalURL, record.UserID)
	if err != nil {
		return err
	}

	return nil
}

// Get is a method that retrieves the original URL from the database.
func (d *DBStore) Get(ctx context.Context, shortURL string) (models.ShortenStore, error) {
	query := `SELECT original_url, is_deleted FROM urls WHERE short_url = $1`

	shortenStore := models.ShortenStore{}

	err := d.DB.QueryRowContext(ctx, query, shortURL).Scan(&shortenStore.OriginalURL, &shortenStore.Deleted)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.ShortenStore{}, filestore.ErrURLNotFound
		}
		return models.ShortenStore{}, err
	}

	return shortenStore, nil
}

// Ping is a method that checks the database connection.
func (d *DBStore) Ping() error {
	return d.DB.Ping()
}

// AddBatch is a method that adds a batch of URLs to the database.
func (d *DBStore) AddBatch(ctx context.Context, batchRequest []models.ShortenBatchStore, userID uuid.UUID) error {
	tx, err := d.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO urls (uuid, short_url, original_url, user_id) VALUES ($1, $2, $3, $4)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, request := range batchRequest {
		_, err = stmt.ExecContext(ctx, request.CorrelationID, request.ShortURL, request.OriginalURL, userID)
		if err != nil {
			logger.Log.Error("Error adding batch request", "error", err)
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		logger.Log.Error("Error committing batch request", "error", err)
		return err
	}

	return nil
}

// GetUserURLs is a method that retrieves all URLs associated with the user ID.
func (d *DBStore) GetUserURLs(ctx context.Context, userID uuid.UUID) ([]models.UserURLResponse, error) {
	query := `SELECT short_url, original_url FROM urls WHERE user_id = $1`

	rows, err := d.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []models.UserURLResponse

	for rows.Next() {
		var url models.UserURLResponse
		if err := rows.Scan(&url.ShortURL, &url.OriginalURL); err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}

	// Проверяем наличие ошибок, которые могут возникнуть во время итерации
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

// DeleteUserURLs is a method that deletes URLs associated with the user ID.
func (d *DBStore) DeleteUserURLs(ctx context.Context, userShortURLs <-chan models.UserShortURL) error {
	tx, err := d.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `UPDATE urls SET is_deleted = true WHERE user_id = $1 AND short_url = $2`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for userShortURL := range userShortURLs {
		_, err = stmt.ExecContext(ctx, userShortURL.UserID, userShortURL.ShortURL)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// GetStats returns the number of URLs and unique users in the database
func (d *DBStore) GetStats(ctx context.Context) (int, int, error) {
	var urlsCount, usersCount int

	// Get total number of URLs
	err := d.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM urls WHERE is_deleted = false`).Scan(&urlsCount)
	if err != nil {
		return 0, 0, err
	}

	// Get number of unique users
	err = d.DB.QueryRowContext(ctx, `SELECT COUNT(DISTINCT user_id) FROM urls WHERE is_deleted = false`).Scan(&usersCount)
	if err != nil {
		return 0, 0, err
	}

	return urlsCount, usersCount, nil
}
