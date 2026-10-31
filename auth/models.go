package auth

import (
	"context"
	"time"
)

type contextKey string

const (
	UserContextKey contextKey = "anthology_live_user"
)

// User represents an account from anthology.db
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserFromContext retrieves the authenticated user from request context.
func UserFromContext(ctx context.Context) (*User, bool) {
	if ctx == nil {
		return nil, false
	}
	u, ok := ctx.Value(UserContextKey).(*User)
	return u, ok && u != nil
}

// WithUser returns a new context with the authenticated user attached.
func WithUser(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, UserContextKey, u)
}
