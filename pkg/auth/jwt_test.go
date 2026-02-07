package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

// Helper to create a valid JWT for testing
func createTestToken(secret string, payload TokenPayload) string {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerBytes, _ := json.Marshal(header)
	payloadBytes, _ := json.Marshal(payload)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)

	signingInput := headerB64 + "." + payloadB64

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(signingInput))
	signature := h.Sum(nil)
	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

	return headerB64 + "." + payloadB64 + "." + signatureB64
}

func TestNewJWTValidator(t *testing.T) {
	tests := []struct {
		name    string
		secret  string
		wantErr bool
	}{
		{
			name:    "valid secret",
			secret:  "this-is-a-very-long-secret-key-that-is-at-least-44-characters-long",
			wantErr: false,
		},
		{
			name:    "secret too short",
			secret:  "short-secret",
			wantErr: true,
		},
		{
			name:    "exactly 44 characters",
			secret:  "12345678901234567890123456789012345678901234",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewJWTValidator(tt.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewJWTValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJWTValidator_Validate(t *testing.T) {
	secret := "this-is-a-very-long-secret-key-that-is-at-least-44-characters-long"
	validator, _ := NewJWTValidator(secret)

	tests := []struct {
		name    string
		payload TokenPayload
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid token",
			payload: TokenPayload{
				Sub:            "user-123",
				Email:          "test@example.com",
				Type:           "access",
				Permissions:    []string{"WORKFLOWS_READ", "WORKFLOWS_EXECUTE"},
				OrganizationID: "org-456",
				Exp:            time.Now().Add(time.Hour).Unix(),
			},
			wantErr: false,
		},
		{
			name: "expired token",
			payload: TokenPayload{
				Sub:  "user-123",
				Type: "access",
				Exp:  time.Now().Add(-time.Hour).Unix(),
			},
			wantErr: true,
			errMsg:  "expired",
		},
		{
			name: "refresh token (invalid type)",
			payload: TokenPayload{
				Sub:  "user-123",
				Type: "refresh",
				Exp:  time.Now().Add(time.Hour).Unix(),
			},
			wantErr: true,
			errMsg:  "invalid token type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := createTestToken(secret, tt.payload)
			result, err := validator.Validate(token)

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
				return
			}

			if !tt.wantErr {
				if result.Sub != tt.payload.Sub {
					t.Errorf("Subject = %v, want %v", result.Sub, tt.payload.Sub)
				}
				if result.Email != tt.payload.Email {
					t.Errorf("Email = %v, want %v", result.Email, tt.payload.Email)
				}
				if result.OrganizationID != tt.payload.OrganizationID {
					t.Errorf("OrganizationID = %v, want %v", result.OrganizationID, tt.payload.OrganizationID)
				}
			}
		})
	}
}

func TestJWTValidator_InvalidSignature(t *testing.T) {
	secret := "this-is-a-very-long-secret-key-that-is-at-least-44-characters-long"
	wrongSecret := "this-is-a-different-secret-key-that-is-at-least-44-characters"

	validator, _ := NewJWTValidator(secret)

	payload := TokenPayload{
		Sub:  "user-123",
		Type: "access",
		Exp:  time.Now().Add(time.Hour).Unix(),
	}

	// Create token with wrong secret
	token := createTestToken(wrongSecret, payload)

	_, err := validator.Validate(token)
	if err == nil {
		t.Error("Validate() should fail with invalid signature")
	}
	if !contains(err.Error(), "signature") {
		t.Errorf("Validate() error = %v, want error about signature", err)
	}
}

func TestJWTValidator_MalformedToken(t *testing.T) {
	secret := "this-is-a-very-long-secret-key-that-is-at-least-44-characters-long"
	validator, _ := NewJWTValidator(secret)

	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"single part", "header"},
		{"two parts", "header.payload"},
		{"invalid base64", "!!!.!!!.!!!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validator.Validate(tt.token)
			if err == nil {
				t.Error("Validate() should fail with malformed token")
			}
		})
	}
}

func TestTokenPayload_HasPermission(t *testing.T) {
	payload := TokenPayload{
		Permissions: []string{"WORKFLOWS_READ", "WORKFLOWS_EXECUTE", "EXECUTIONS_READ"},
	}

	tests := []struct {
		permission string
		want       bool
	}{
		{"WORKFLOWS_READ", true},
		{"WORKFLOWS_EXECUTE", true},
		{"EXECUTIONS_READ", true},
		{"WORKFLOWS_DELETE", false},
		{"ADMIN", false},
	}

	for _, tt := range tests {
		t.Run(tt.permission, func(t *testing.T) {
			if got := payload.HasPermission(tt.permission); got != tt.want {
				t.Errorf("HasPermission(%q) = %v, want %v", tt.permission, got, tt.want)
			}
		})
	}
}

func TestTokenPayload_HasAnyPermission(t *testing.T) {
	payload := TokenPayload{
		Permissions: []string{"WORKFLOWS_READ", "EXECUTIONS_READ"},
	}

	tests := []struct {
		name        string
		permissions []string
		want        bool
	}{
		{"has one", []string{"WORKFLOWS_READ"}, true},
		{"has one of many", []string{"WORKFLOWS_DELETE", "WORKFLOWS_READ"}, true},
		{"has none", []string{"WORKFLOWS_DELETE", "ADMIN"}, false},
		{"empty permissions", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := payload.HasAnyPermission(tt.permissions...); got != tt.want {
				t.Errorf("HasAnyPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokenPayload_HasAllPermissions(t *testing.T) {
	payload := TokenPayload{
		Permissions: []string{"WORKFLOWS_READ", "WORKFLOWS_EXECUTE", "EXECUTIONS_READ"},
	}

	tests := []struct {
		name        string
		permissions []string
		want        bool
	}{
		{"has all", []string{"WORKFLOWS_READ", "EXECUTIONS_READ"}, true},
		{"has some", []string{"WORKFLOWS_READ", "ADMIN"}, false},
		{"has none", []string{"ADMIN", "WORKFLOWS_DELETE"}, false},
		{"empty permissions", []string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := payload.HasAllPermissions(tt.permissions...); got != tt.want {
				t.Errorf("HasAllPermissions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokenPayload_IsExpired(t *testing.T) {
	tests := []struct {
		name    string
		exp     int64
		expired bool
	}{
		{"future expiration", time.Now().Add(time.Hour).Unix(), false},
		{"past expiration", time.Now().Add(-time.Hour).Unix(), true},
		{"no expiration", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := TokenPayload{Exp: tt.exp}
			if got := payload.IsExpired(); got != tt.expired {
				t.Errorf("IsExpired() = %v, want %v", got, tt.expired)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
