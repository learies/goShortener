package contextutils

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUserIDContext(t *testing.T) {
	t.Run("WithUserID and GetUserID", func(t *testing.T) {
		// Создаем тестовый UUID
		expectedUserID := uuid.New()

		// Создаем контекст с userID
		ctx := WithUserID(context.Background(), expectedUserID)

		// Получаем userID из контекста
		actualUserID, ok := GetUserID(ctx)

		// Проверяем, что userID успешно получен
		assert.True(t, ok, "GetUserID should return true")
		assert.Equal(t, expectedUserID, actualUserID, "UserID should match")
	})

	t.Run("GetUserID with empty context", func(t *testing.T) {
		// Пытаемся получить userID из пустого контекста
		userID, ok := GetUserID(context.Background())

		// Проверяем, что получение не удалось
		assert.False(t, ok, "GetUserID should return false for empty context")
		assert.Equal(t, uuid.UUID{}, userID, "UserID should be empty for empty context")
	})

	t.Run("GetUserID with wrong type", func(t *testing.T) {
		// Создаем контекст с неправильным типом значения
		ctx := context.WithValue(context.Background(), userIDContextKey, "not a UUID")

		// Пытаемся получить userID
		userID, ok := GetUserID(ctx)

		// Проверяем, что получение не удалось
		assert.False(t, ok, "GetUserID should return false for wrong type")
		assert.Equal(t, uuid.UUID{}, userID, "UserID should be empty for wrong type")
	})
}
