package worker

import (
	"github.com/learies/goShortener/internal/models"
)

// DeleteUserURLs is a function that creates a channel to delete user URLs.
func DeleteUserURLs(deleteUserUrls ...models.UserShortURL) chan models.UserShortURL {
	ch := make(chan models.UserShortURL, len(deleteUserUrls))
	go func() {
		defer close(ch)
		for _, userURL := range deleteUserUrls {
			ch <- userURL
		}
	}()
	return ch
}
