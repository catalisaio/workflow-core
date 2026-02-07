package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

// Context keys for storing auth information.
type contextKey string

const (
	// ContextKeyToken is the context key for the validated token payload.
	ContextKeyToken contextKey = "auth_token"
	// ContextKeyUserID is the context key for the user ID.
	ContextKeyUserID contextKey = "auth_user_id"
	// ContextKeyOrgID is the context key for the organization ID.
	ContextKeyOrgID contextKey = "auth_org_id"
	// ContextKeyPermissions is the context key for the user's permissions.
	ContextKeyPermissions contextKey = "auth_permissions"
)

// AuthMiddleware provides HTTP middleware for authentication.
type AuthMiddleware struct {
	validator *JWTValidator
	optional  bool
}

// AuthMiddlewareConfig holds configuration for the auth middleware.
type AuthMiddlewareConfig struct {
	// JWTSecret is the secret used to validate JWT tokens.
	// Must match the JWT_SECRET configured in the IAM service.
	JWTSecret string
	// Optional makes authentication optional (for public endpoints).
	Optional bool
}

// NewAuthMiddleware creates a new authentication middleware.
func NewAuthMiddleware(config AuthMiddlewareConfig) (*AuthMiddleware, error) {
	validator, err := NewJWTValidator(config.JWTSecret)
	if err != nil {
		return nil, err
	}

	return &AuthMiddleware{
		validator: validator,
		optional:  config.Optional,
	}, nil
}

// Handler returns an HTTP handler that validates JWT tokens.
func (m *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := m.extractToken(r)
		if err != nil {
			if m.optional {
				// For optional auth, continue without authentication
				next.ServeHTTP(w, r)
				return
			}
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		payload, err := m.validator.Validate(token)
		if err != nil {
			if m.optional {
				next.ServeHTTP(w, r)
				return
			}
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Add token payload to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextKeyToken, payload)
		ctx = context.WithValue(ctx, ContextKeyUserID, payload.Sub)
		ctx = context.WithValue(ctx, ContextKeyOrgID, payload.OrganizationID)
		ctx = context.WithValue(ctx, ContextKeyPermissions, payload.Permissions)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePermission returns middleware that checks for a specific permission.
func (m *AuthMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			payload := GetTokenFromContext(r.Context())
			if payload == nil {
				http.Error(w, "Unauthorized: no authentication", http.StatusUnauthorized)
				return
			}

			if !payload.HasPermission(permission) {
				http.Error(w, "Forbidden: missing required permission", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission returns middleware that checks for any of the given permissions.
func (m *AuthMiddleware) RequireAnyPermission(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			payload := GetTokenFromContext(r.Context())
			if payload == nil {
				http.Error(w, "Unauthorized: no authentication", http.StatusUnauthorized)
				return
			}

			if !payload.HasAnyPermission(permissions...) {
				http.Error(w, "Forbidden: missing required permission", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAllPermissions returns middleware that checks for all given permissions.
func (m *AuthMiddleware) RequireAllPermissions(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			payload := GetTokenFromContext(r.Context())
			if payload == nil {
				http.Error(w, "Unauthorized: no authentication", http.StatusUnauthorized)
				return
			}

			if !payload.HasAllPermissions(permissions...) {
				http.Error(w, "Forbidden: missing required permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireOrganization returns middleware that ensures an organization context is present.
func (m *AuthMiddleware) RequireOrganization() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			orgID := GetOrganizationIDFromContext(r.Context())
			if orgID == "" {
				http.Error(w, "Forbidden: organization context required", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractToken extracts the JWT token from the Authorization header.
func (m *AuthMiddleware) extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing Authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid Authorization header format")
	}

	return parts[1], nil
}

// GetTokenFromContext retrieves the token payload from the context.
func GetTokenFromContext(ctx context.Context) *TokenPayload {
	if payload, ok := ctx.Value(ContextKeyToken).(*TokenPayload); ok {
		return payload
	}
	return nil
}

// GetUserIDFromContext retrieves the user ID from the context.
func GetUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(ContextKeyUserID).(string); ok {
		return userID
	}
	return ""
}

// GetOrganizationIDFromContext retrieves the organization ID from the context.
func GetOrganizationIDFromContext(ctx context.Context) string {
	if orgID, ok := ctx.Value(ContextKeyOrgID).(string); ok {
		return orgID
	}
	return ""
}

// GetPermissionsFromContext retrieves the permissions from the context.
func GetPermissionsFromContext(ctx context.Context) []string {
	if perms, ok := ctx.Value(ContextKeyPermissions).([]string); ok {
		return perms
	}
	return nil
}

// MustGetToken retrieves the token from context or panics.
// Use this only when you're certain authentication has occurred.
func MustGetToken(ctx context.Context) *TokenPayload {
	payload := GetTokenFromContext(ctx)
	if payload == nil {
		panic("auth: no token in context (was authentication middleware applied?)")
	}
	return payload
}
