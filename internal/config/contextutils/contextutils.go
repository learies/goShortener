package contextutils

import (
	"context"

	"github.com/google/uuid"
)

// Contextual key for userID
type contextKey struct {
	name string
}

// userIDContextKey is a global variable that holds the context key for userID.
var userIDContextKey = &contextKey{"userID"}

// GetUserID is a function that retrieves the userID from the context.
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDContextKey).(uuid.UUID)
	return userID, ok
}

// WithUserID is a function that adds the userID to the context.
func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}
