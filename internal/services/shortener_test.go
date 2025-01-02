package services

import (
	"testing"
)

func TestGenerateShortURL(t *testing.T) {
	shortener := NewURLShortener()

	// Test case 1: Ensure the short URL length is 8
	t.Run("Check short URL length", func(t *testing.T) {
		url := "https://example.com"
		shortURL := shortener.GenerateShortURL(url)
		if len(shortURL) != 8 {
			t.Errorf("expected length 8, got %d", len(shortURL))
		}
	})

	// Test case 2: Ensure determinism
	t.Run("Check determinism", func(t *testing.T) {
		url := "https://example.com"
		shortURL1 := shortener.GenerateShortURL(url)
		shortURL2 := shortener.GenerateShortURL(url)
		if shortURL1 != shortURL2 {
			t.Errorf("expected the same short URL, got %s and %s", shortURL1, shortURL2)
		}
	})

	// Test case 3: Check collision
	t.Run("Check collision", func(t *testing.T) {
		url1 := "https://example.com/1"
		url2 := "https://example.com/2"
		shortURL1 := shortener.GenerateShortURL(url1)
		shortURL2 := shortener.GenerateShortURL(url2)
		if shortURL1 == shortURL2 {
			t.Errorf("expected different short URLs for different inputs, got %s and %s", shortURL1, shortURL2)
		}
	})
}
