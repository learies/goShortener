package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateShortURL(t *testing.T) {
	shortener := NewURLShortener()
	t.Run("TestShortURLLength", func(t *testing.T) {
		inputURL := "https://example.com"
		expectedLength := 8

		shortURL, err := shortener.GenerateShortURL(inputURL)
		assert.NoError(t, err)

		// Проверяем длину сгенерированного короткого URL
		assert.Equal(t, expectedLength, len(shortURL))
	})

	t.Run("TestUniqueShortURLs", func(t *testing.T) {
		// Тестовые примеры для проверки уникальности
		url1 := "https://example.com"
		url2 := "https://another-example.com"

		shortURL1, err := shortener.GenerateShortURL(url1)
		assert.NoError(t, err)

		shortURL2, err := shortener.GenerateShortURL(url2)
		assert.NoError(t, err)

		// Проверяем, что разные URL создают разные короткие URL
		if shortURL1 == shortURL2 {
			t.Errorf("Expected unique short URLs for '%s' and '%s', but got '%s'",
				url1, url2, shortURL1)
		}
	})

	t.Run("TestConsistentShortURL", func(t *testing.T) {
		inputURL := "https://consistent-url.com"

		shortURL1, err := shortener.GenerateShortURL(inputURL)
		assert.NoError(t, err)

		shortURL2, err := shortener.GenerateShortURL(inputURL)
		assert.NoError(t, err)

		// Проверяем, что одинаковые входные данные создают одинаковый короткий URL
		if shortURL1 != shortURL2 {
			t.Errorf("Short URLs for the same input '%s' should be consistent, but got '%s' and '%s'",
				inputURL, shortURL1, shortURL2)
		}
	})

	t.Run("TestEmptyURL", func(t *testing.T) {
		// Arrange
		inputURL := ""

		// Act
		_, err := shortener.GenerateShortURL(inputURL)

		// Assert
		if assert.ErrorAs(t, err, &ErrEmptyURL) {
			assert.Equal(t, ErrEmptyURL, err, "Error should be ErrEmptyURL")
		}
	})

	t.Run("TestLongURL", func(t *testing.T) {
		// Arrange
		inputURL := "https://example.com/this-is-a-very-long-url"

		// Act
		shortURL, err := shortener.GenerateShortURL(inputURL)

		// Assert
		if assert.NoError(t, err) {
			assert.LessOrEqual(t, len(shortURL), 8, "Short URL should not exceed 8 characters")
		}
	})

	t.Run("TestSpecialCharactersURL", func(t *testing.T) {
		// Arrange
		inputURL := "https://example.com/!@#$%^&*()"

		// Act
		shortURL, err := shortener.GenerateShortURL(inputURL)

		// Assert
		if assert.NoError(t, err) {
			assert.NotContains(t, shortURL, "!@#$%^&*()", "Short URL should not contain special characters")
		}
	})

	t.Run("TestURLWithSpaces", func(t *testing.T) {
		// Arrange
		inputURL := "https://example.com/url with spaces"

		// Act
		shortURL, err := shortener.GenerateShortURL(inputURL)

		// Assert
		if assert.NoError(t, err) {
			assert.NotContains(t, shortURL, " ", "Short URL should not contain spaces")
		}
	})

	t.Run("TestURLWithQueryParams", func(t *testing.T) {
		// Arrange
		inputURL := "https://example.com/?utm_source=google"

		// Act
		shortURL, err := shortener.GenerateShortURL(inputURL)

		// Assert
		if assert.NoError(t, err) {
			assert.NotContains(t, shortURL, "?", "Short URL should not contain query parameters")
		}
	})

	t.Run("TestURLWithFragment", func(t *testing.T) {
		// Arrange
		inputURL := "https://example.com/#section"

		// Act
		shortURL, err := shortener.GenerateShortURL(inputURL)

		// Assert
		if assert.NoError(t, err) {
			assert.NotContains(t, shortURL, "#", "Short URL should not contain fragments")
		}
	})
}
