package webcontext

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const (
	userIDKey        contextKey = "user_id"
	rolesKey         contextKey = "roles"
	correlationIDKey contextKey = "correlation_id"
)

func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}

func WithRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, rolesKey, roles)
}

func RolesFromContext(ctx context.Context) ([]string, bool) {
	roles, ok := ctx.Value(rolesKey).([]string)
	return roles, ok
}

func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey, correlationID)
}

func CorrelationIDFromContext(ctx context.Context) (string, bool) {
	correlationID, ok := ctx.Value(correlationIDKey).(string)
	return correlationID, ok
}
