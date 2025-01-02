package services

import (
	"crypto/sha256"
	"encoding/base64"
)

type Shortener interface {
	GenerateShortURL(url string) string
}

type URLShortener struct{}

func NewURLShortener() *URLShortener {
	return &URLShortener{}
}

func (us *URLShortener) GenerateShortURL(url string) string {
	hasher := sha256.New()

	hasher.Write([]byte(url))
	hash := hasher.Sum(nil)
	shortURL := base64.URLEncoding.EncodeToString(hash)[:8]

	return shortURL
}
