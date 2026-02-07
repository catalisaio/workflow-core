// Package auth provides authentication and authorization functionality
// for workflow-core, integrating with the building-blocks IAM service.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// TokenPayload represents the JWT token payload structure from IAM.
type TokenPayload struct {
	// Subject (user ID)
	Sub string `json:"sub"`
	// User email
	Email string `json:"email,omitempty"`
	// Token type: "access" or "refresh"
	Type string `json:"type"`
	// Permissions granted to the user
	Permissions []string `json:"permissions,omitempty"`
	// Organization context (multi-tenant)
	OrganizationID string `json:"organizationId,omitempty"`
	// Issued at timestamp
	Iat int64 `json:"iat,omitempty"`
	// Expiration timestamp
	Exp int64 `json:"exp,omitempty"`
	// JWT ID (for refresh tokens)
	Jti string `json:"jti,omitempty"`
}

// JWTValidator validates JWT tokens issued by the IAM service.
type JWTValidator struct {
	secret []byte
}

// NewJWTValidator creates a new JWT validator with the given secret.
// The secret must match the JWT_SECRET configured in the IAM service.
func NewJWTValidator(secret string) (*JWTValidator, error) {
	if len(secret) < 44 {
		return nil, errors.New("JWT secret must be at least 44 characters for 256-bit security")
	}
	return &JWTValidator{
		secret: []byte(secret),
	}, nil
}

// Validate verifies a JWT token and returns its payload.
func (v *JWTValidator) Validate(token string) (*TokenPayload, error) {
	// Split the token into parts
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	headerB64, payloadB64, signatureB64 := parts[0], parts[1], parts[2]

	// Verify the signature
	if err := v.verifySignature(headerB64, payloadB64, signatureB64); err != nil {
		return nil, fmt.Errorf("invalid signature: %w", err)
	}

	// Verify header
	header, err := v.decodeHeader(headerB64)
	if err != nil {
		return nil, fmt.Errorf("invalid header: %w", err)
	}
	if header.Alg != "HS256" {
		return nil, fmt.Errorf("unsupported algorithm: %s", header.Alg)
	}

	// Decode payload
	payload, err := v.decodePayload(payloadB64)
	if err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	// Check expiration
	if payload.Exp > 0 {
		if time.Now().Unix() > payload.Exp {
			return nil, errors.New("token expired")
		}
	}

	// Check token type (we only accept access tokens)
	if payload.Type != "access" {
		return nil, errors.New("invalid token type: expected access token")
	}

	return payload, nil
}

// jwtHeader represents the JWT header.
type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ,omitempty"`
}

// verifySignature verifies the JWT signature using HMAC-SHA256.
func (v *JWTValidator) verifySignature(headerB64, payloadB64, signatureB64 string) error {
	// Decode signature
	signature, err := base64URLDecode(signatureB64)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	// Create signing input
	signingInput := headerB64 + "." + payloadB64

	// Compute expected signature
	h := hmac.New(sha256.New, v.secret)
	h.Write([]byte(signingInput))
	expectedSig := h.Sum(nil)

	// Compare signatures using constant-time comparison
	if !hmac.Equal(signature, expectedSig) {
		return errors.New("signature mismatch")
	}

	return nil
}

// decodeHeader decodes the JWT header.
func (v *JWTValidator) decodeHeader(headerB64 string) (*jwtHeader, error) {
	data, err := base64URLDecode(headerB64)
	if err != nil {
		return nil, err
	}

	var header jwtHeader
	if err := json.Unmarshal(data, &header); err != nil {
		return nil, err
	}

	return &header, nil
}

// decodePayload decodes the JWT payload.
func (v *JWTValidator) decodePayload(payloadB64 string) (*TokenPayload, error) {
	data, err := base64URLDecode(payloadB64)
	if err != nil {
		return nil, err
	}

	var payload TokenPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}

	return &payload, nil
}

// base64URLDecode decodes a base64url-encoded string without padding.
func base64URLDecode(s string) ([]byte, error) {
	// Add padding if necessary
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}

// HasPermission checks if the token has a specific permission.
func (p *TokenPayload) HasPermission(permission string) bool {
	for _, perm := range p.Permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if the token has any of the given permissions.
func (p *TokenPayload) HasAnyPermission(permissions ...string) bool {
	for _, required := range permissions {
		if p.HasPermission(required) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if the token has all of the given permissions.
func (p *TokenPayload) HasAllPermissions(permissions ...string) bool {
	for _, required := range permissions {
		if !p.HasPermission(required) {
			return false
		}
	}
	return true
}

// IsExpired checks if the token is expired.
func (p *TokenPayload) IsExpired() bool {
	if p.Exp <= 0 {
		return false
	}
	return time.Now().Unix() > p.Exp
}

// ExpiresIn returns the duration until the token expires.
func (p *TokenPayload) ExpiresIn() time.Duration {
	if p.Exp <= 0 {
		return 0
	}
	return time.Until(time.Unix(p.Exp, 0))
}
