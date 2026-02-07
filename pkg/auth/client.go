package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// IAMClient provides a client for the building-blocks IAM service.
type IAMClient struct {
	baseURL    string
	httpClient *http.Client
}

// IAMClientConfig holds configuration for the IAM client.
type IAMClientConfig struct {
	// BaseURL is the base URL of the IAM service (e.g., "http://localhost:3000/iam")
	BaseURL string
	// Timeout for HTTP requests
	Timeout time.Duration
}

// NewIAMClient creates a new IAM service client.
func NewIAMClient(config IAMClientConfig) *IAMClient {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &IAMClient{
		baseURL: config.BaseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// LoginRequest represents a login request to IAM.
type LoginRequest struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	OrganizationID string `json:"organizationId,omitempty"`
}

// LoginResponse represents a login response from IAM.
type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken,omitempty"`
	TokenType    string `json:"tokenType"`
	ExpiresIn    int    `json:"expiresIn"`
	UserID       string `json:"userId,omitempty"`
	Email        string `json:"email,omitempty"`
}

// TokenRequest represents an OAuth token request.
type TokenRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// TokenResponse represents an OAuth token response.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// UserInfo represents user information from IAM.
type UserInfo struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	Status      string   `json:"status"`
	Permissions []string `json:"permissions,omitempty"`
}

// IAMError represents an error from the IAM service.
type IAMError struct {
	StatusCode int
	Message    string
	Details    []ErrorDetail
}

// ErrorDetail provides details about an error.
type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

func (e *IAMError) Error() string {
	return fmt.Sprintf("IAM error (status %d): %s", e.StatusCode, e.Message)
}

// Login authenticates a user with the IAM service.
func (c *IAMClient) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/users/login", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &loginResp, nil
}

// GetToken exchanges credentials for an OAuth token.
func (c *IAMClient) GetToken(ctx context.Context, req TokenRequest) (*TokenResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/oauth/token", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &tokenResp, nil
}

// RefreshToken refreshes an access token using a refresh token.
func (c *IAMClient) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	return c.GetToken(ctx, TokenRequest{
		GrantType:    "refresh_token",
		RefreshToken: refreshToken,
	})
}

// ValidateToken validates a token with the IAM service.
// Note: For performance, use local JWT validation (JWTValidator) when possible.
// This method makes a round-trip to IAM and should only be used when
// additional validation is needed (e.g., checking if token was revoked).
func (c *IAMClient) ValidateToken(ctx context.Context, token string) (*UserInfo, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/v1/users/me", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, &IAMError{
			StatusCode: resp.StatusCode,
			Message:    "invalid or expired token",
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &userInfo, nil
}

// parseError parses an error response from the IAM service.
func (c *IAMClient) parseError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	iamErr := &IAMError{
		StatusCode: resp.StatusCode,
		Message:    "unknown error",
	}

	// Try to parse as JSON error
	var errResp struct {
		Message string        `json:"message"`
		Error   string        `json:"error"`
		Details []ErrorDetail `json:"details,omitempty"`
	}
	if err := json.Unmarshal(body, &errResp); err == nil {
		if errResp.Message != "" {
			iamErr.Message = errResp.Message
		} else if errResp.Error != "" {
			iamErr.Message = errResp.Error
		}
		iamErr.Details = errResp.Details
	} else if len(body) > 0 {
		iamErr.Message = string(body)
	}

	return iamErr
}
