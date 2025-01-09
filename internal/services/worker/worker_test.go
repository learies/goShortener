package worker

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/learies/goShortener/internal/models"
)

func TestDeleteUserURLs(t *testing.T) {
	testData := []models.UserShortURL{
		{UserID: uuid.New(), ShortURL: "short1"},
		{UserID: uuid.New(), ShortURL: "short2"},
		{UserID: uuid.New(), ShortURL: "short3"},
	}

	ch := DeleteUserURLs(testData...)

	var result []models.UserShortURL
	for url := range ch {
		result = append(result, url)
	}

	assert.ElementsMatch(t, testData, result, "The input and output of DeleteUserURLs should match")
}
