package services

import (
	"crypto/sha256"
	"encoding/base64"
)

// Shortener определяет метод для генерации сокращённых URL
type Shortener interface {
	GenerateShortURL(url string) string
}

// URLShortener структура
type URLShortener struct{}

// NewURLShortener создаёт новую структуру URLShortener
func NewURLShortener() *URLShortener {
	return &URLShortener{}
}

// GenerateShortURL генерирует сокращённый URL, используя SHA256 и base64
func (us *URLShortener) GenerateShortURL(url string) string {
	if url == "" {
		return ""
	}

	hasher := sha256.New()
	hasher.Write([]byte(url))
	hash := hasher.Sum(nil)

	shortURL := base64.URLEncoding.EncodeToString(hash)[:8]
	return shortURL
}
