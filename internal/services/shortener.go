// Package services provides core business logic for the URL shortener service.
package services

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

// ErrEmptyURL ошибка, возникающая при попытке сократить пустой URL
var ErrEmptyURL = errors.New("empty URL")

// Shortener определяет метод для генерации сокращённых URL
type Shortener interface {
	GenerateShortURL(url string) (string, error)
}

// URLShortener структура
type URLShortener struct{}

// NewURLShortener создаёт новую структуру URLShortener
func NewURLShortener() *URLShortener {
	return &URLShortener{}
}

// GenerateShortURL генерирует сокращённый URL, используя SHA256 и base64
func (us *URLShortener) GenerateShortURL(url string) (string, error) {
	if url == "" {
		return "", ErrEmptyURL
	}

	hasher := sha256.New()
	hasher.Write([]byte(url))
	hash := hasher.Sum(nil)

	shortURL := base64.URLEncoding.EncodeToString(hash)[:8]
	return shortURL, nil
}
