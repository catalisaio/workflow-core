package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthMiddleware_Handler(t *testing.T) {
	secret := "this-is-a-very-long-secret-key-that-is-at-least-44-characters-long"
	middleware, err := NewAuthMiddleware(AuthMiddlewareConfig{
		JWTSecret: secret,
		Optional:  false,
	})
	if err != nil {
		t.Fatalf("Failed to create middleware: %v", err)
	}

	// Create a test handler that checks the context
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := GetTokenFromContext(r.Context())
		if token == nil {
			t.Error("Token should be in context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if token.Sub != "user-123" {
			t.Errorf("Expected sub=user-123, got %s", token.Sub)
		}
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.Handler(handler)

	t.Run("valid token", func(t *testing.T) {
		token := createTestToken(secret, TokenPayload{
			Sub:  "user-123",
			Type: "access",
			Exp:  time.Now().Add(time.Hour).Unix(),
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
	})

	t.Run("missing authorization header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", rec.Code)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", rec.Code)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		token := createTestToken(secret, TokenPayload{
			Sub:  "user-123",
			Type: "access",
			Exp:  time.Now().Add(-time.Hour).Unix(),
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", rec.Code)
		}
	})
}

func TestAuthMiddleware_Optional(t *testing.T) {
	secret := "this-is-a-very-long-secret-key-that-is-at-least-44-characters-long"
	middleware, _ := NewAuthMiddleware(AuthMiddlewareConfig{
		JWTSecret: secret,
		Optional:  true,
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.Handler(handler)

	t.Run("missing auth header allowed", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
	})
}

func TestAuthMiddleware_RequirePermission(t *testing.T) {
	secret := "this-is-a-very-long-secret-key-that-is-at-least-44-characters-long"
	middleware, _ := NewAuthMiddleware(AuthMiddlewareConfig{
		JWTSecret: secret,
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.Handler(
		middleware.RequirePermission(PermissionWorkflowsExecute)(handler),
	)

	t.Run("has required permission", func(t *testing.T) {
		token := createTestToken(secret, TokenPayload{
			Sub:         "user-123",
			Type:        "access",
			Permissions: []string{PermissionWorkflowsExecute, PermissionWorkflowsRead},
			Exp:         time.Now().Add(time.Hour).Unix(),
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
	})

	t.Run("missing required permission", func(t *testing.T) {
		token := createTestToken(secret, TokenPayload{
			Sub:         "user-123",
			Type:        "access",
			Permissions: []string{PermissionWorkflowsRead},
			Exp:         time.Now().Add(time.Hour).Unix(),
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", rec.Code)
		}
	})
}

func TestAuthMiddleware_RequireOrganization(t *testing.T) {
	secret := "this-is-a-very-long-secret-key-that-is-at-least-44-characters-long"
	middleware, _ := NewAuthMiddleware(AuthMiddlewareConfig{
		JWTSecret: secret,
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.Handler(
		middleware.RequireOrganization()(handler),
	)

	t.Run("has organization", func(t *testing.T) {
		token := createTestToken(secret, TokenPayload{
			Sub:            "user-123",
			Type:           "access",
			OrganizationID: "org-456",
			Exp:            time.Now().Add(time.Hour).Unix(),
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
	})

	t.Run("missing organization", func(t *testing.T) {
		token := createTestToken(secret, TokenPayload{
			Sub:  "user-123",
			Type: "access",
			Exp:  time.Now().Add(time.Hour).Unix(),
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", rec.Code)
		}
	})
}

func TestGetContextFunctions(t *testing.T) {
	payload := &TokenPayload{
		Sub:            "user-123",
		OrganizationID: "org-456",
		Permissions:    []string{"WORKFLOWS_READ"},
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, ContextKeyToken, payload)
	ctx = context.WithValue(ctx, ContextKeyUserID, payload.Sub)
	ctx = context.WithValue(ctx, ContextKeyOrgID, payload.OrganizationID)
	ctx = context.WithValue(ctx, ContextKeyPermissions, payload.Permissions)

	t.Run("GetTokenFromContext", func(t *testing.T) {
		got := GetTokenFromContext(ctx)
		if got == nil {
			t.Fatal("Expected token, got nil")
		}
		if got.Sub != payload.Sub {
			t.Errorf("Expected sub=%s, got %s", payload.Sub, got.Sub)
		}
	})

	t.Run("GetUserIDFromContext", func(t *testing.T) {
		got := GetUserIDFromContext(ctx)
		if got != payload.Sub {
			t.Errorf("Expected %s, got %s", payload.Sub, got)
		}
	})

	t.Run("GetOrganizationIDFromContext", func(t *testing.T) {
		got := GetOrganizationIDFromContext(ctx)
		if got != payload.OrganizationID {
			t.Errorf("Expected %s, got %s", payload.OrganizationID, got)
		}
	})

	t.Run("GetPermissionsFromContext", func(t *testing.T) {
		got := GetPermissionsFromContext(ctx)
		if len(got) != 1 || got[0] != "WORKFLOWS_READ" {
			t.Errorf("Expected [WORKFLOWS_READ], got %v", got)
		}
	})

	t.Run("empty context", func(t *testing.T) {
		emptyCtx := context.Background()
		if GetTokenFromContext(emptyCtx) != nil {
			t.Error("Expected nil token from empty context")
		}
		if GetUserIDFromContext(emptyCtx) != "" {
			t.Error("Expected empty user ID from empty context")
		}
		if GetOrganizationIDFromContext(emptyCtx) != "" {
			t.Error("Expected empty org ID from empty context")
		}
		if GetPermissionsFromContext(emptyCtx) != nil {
			t.Error("Expected nil permissions from empty context")
		}
	})
}
