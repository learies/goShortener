package contextutils

import (
	"context"

	"github.com/google/uuid"
)

// Contextual key for userID
type contextKey struct {
	name string
}

var userIDContextKey = &contextKey{"userID"}

func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDContextKey).(uuid.UUID)
	return userID, ok
}

func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}
