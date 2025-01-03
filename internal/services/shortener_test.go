package services

import (
	"testing"
)

func TestGenerateShortURL(t *testing.T) {
	shortener := NewURLShortener()
	t.Run("TestShortURLLength", func(t *testing.T) {
		inputURL := "https://example.com"
		expectedLength := 8

		shortURL := shortener.GenerateShortURL(inputURL)

		// Проверяем длину сгенерированного короткого URL
		if len(shortURL) != expectedLength {
			t.Errorf("Expected short URL of length %d, but got %d for URL '%s'",
				expectedLength, len(shortURL), inputURL)
		}
	})

	t.Run("TestUniqueShortURLs", func(t *testing.T) {
		// Тестовые примеры для проверки уникальности
		url1 := "https://example.com"
		url2 := "https://another-example.com"

		shortURL1 := shortener.GenerateShortURL(url1)
		shortURL2 := shortener.GenerateShortURL(url2)

		// Проверяем, что разные URL создают разные короткие URL
		if shortURL1 == shortURL2 {
			t.Errorf("Expected unique short URLs for '%s' and '%s', but got '%s'",
				url1, url2, shortURL1)
		}
	})

	t.Run("TestConsistentShortURL", func(t *testing.T) {
		inputURL := "https://consistent-url.com"

		shortURL1 := shortener.GenerateShortURL(inputURL)
		shortURL2 := shortener.GenerateShortURL(inputURL)

		// Проверяем, что одинаковые входные данные создают одинаковый короткий URL
		if shortURL1 != shortURL2 {
			t.Errorf("Short URLs for the same input '%s' should be consistent, but got '%s' and '%s'",
				inputURL, shortURL1, shortURL2)
		}
	})

	t.Run("TestEmptyURL", func(t *testing.T) {
		inputURL := ""

		shortURL := shortener.GenerateShortURL(inputURL)

		// Проверяем, что пустой URL не создаёт короткий URL
		if shortURL != "" {
			t.Errorf("Expected empty short URL for empty input, but got '%s'", shortURL)
		}
	})

	t.Run("TestLongURL", func(t *testing.T) {
		inputURL := "https://example.com/this-is-a-very-long-url"

		shortURL := shortener.GenerateShortURL(inputURL)

		// Проверяем, что длинный URL создаёт короткий URL
		if shortURL == "" {
			t.Errorf("Expected short URL for long input, but got empty short URL")
		}
	})

	t.Run("TestSpecialCharactersURL", func(t *testing.T) {
		inputURL := "https://example.com/!@#$%^&*()"

		shortURL := shortener.GenerateShortURL(inputURL)

		// Проверяем, что URL с специальными символами создаёт короткий URL
		if shortURL == "" {
			t.Errorf("Expected short URL for input with special characters, but got empty short URL")
		}
	})

	t.Run("TestURLWithSpaces", func(t *testing.T) {
		inputURL := "https://example.com/url with spaces"

		shortURL := shortener.GenerateShortURL(inputURL)

		// Проверяем, что URL с пробелами создаёт короткий URL
		if shortURL == "" {
			t.Errorf("Expected short URL for input with spaces, but got empty short URL")
		}
	})

	t.Run("TestURLWithQueryParams", func(t *testing.T) {
		inputURL := "https://example.com/?utm_source=google"

		shortURL := shortener.GenerateShortURL(inputURL)

		// Проверяем, что URL с параметрами запроса создаёт короткий URL
		if shortURL == "" {
			t.Errorf("Expected short URL for input with query params, but got empty short URL")
		}
	})

	t.Run("TestURLWithFragment", func(t *testing.T) {
		inputURL := "https://example.com/#section"

		shortURL := shortener.GenerateShortURL(inputURL)

		// Проверяем, что URL с фрагментом создаёт короткий URL
		if shortURL == "" {
			t.Errorf("Expected short URL for input with fragment, but got empty short URL")
		}
	})
}
