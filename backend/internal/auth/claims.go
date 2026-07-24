package auth

import (
	"context"
)

type contextKey string

const userKey contextKey = "user"

// Role constants must match the role names configured in the Keycloak realm.
const (
	RoleDeveloper = "PulseDeveloper"
	RoleManager   = "PulseManager"
)

// User holds the authenticated user's identity and roles, extracted from the JWT.
type User struct {
	Email       string   `json:"email"`
	DisplayName string   `json:"displayName"`
	Roles       []string `json:"roles"`
}

// MockManager is the synthetic user injected in LOCAL_DEV mode.
var MockManager = &User{
	Email:       "dev@solvas.local",
	DisplayName: "Dev User",
	Roles:       []string{RoleDeveloper, RoleManager},
}

// FromContext retrieves the authenticated User from the request context.
// Returns nil if no user is present (unauthenticated request).
func FromContext(ctx context.Context) *User {
	u, _ := ctx.Value(userKey).(*User)
	return u
}

// WithUser returns a new context carrying the given User.
func WithUser(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

// HasRole reports whether the user has the specified role.
func (u *User) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// IsManager reports whether the user has the PulseManager role.
func (u *User) IsManager() bool {
	return u.HasRole(RoleManager)
}
