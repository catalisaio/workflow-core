package auth

import (
	"errors"
	"os"
	"strconv"
	"time"
)

// Config holds authentication configuration.
type Config struct {
	// IAMBaseURL is the base URL of the IAM service.
	// Example: "http://localhost:3000/iam"
	IAMBaseURL string

	// JWTSecret is the secret used to validate JWT tokens.
	// Must match the JWT_SECRET configured in the IAM service.
	JWTSecret string

	// RequestTimeout is the timeout for requests to the IAM service.
	RequestTimeout time.Duration

	// AuthEnabled determines if authentication is required.
	// When false, all requests are allowed (for development).
	AuthEnabled bool

	// RequireOrganization determines if organization context is required.
	RequireOrganization bool
}

// LoadConfigFromEnv loads authentication configuration from environment variables.
func LoadConfigFromEnv() (*Config, error) {
	config := &Config{
		IAMBaseURL:          os.Getenv("IAM_BASE_URL"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		RequestTimeout:      30 * time.Second,
		AuthEnabled:         true,
		RequireOrganization: false,
	}

	// Parse request timeout
	if timeout := os.Getenv("IAM_REQUEST_TIMEOUT"); timeout != "" {
		d, err := time.ParseDuration(timeout)
		if err != nil {
			// Try parsing as seconds
			if secs, err := strconv.Atoi(timeout); err == nil {
				d = time.Duration(secs) * time.Second
			}
		}
		if d > 0 {
			config.RequestTimeout = d
		}
	}

	// Parse auth enabled
	if enabled := os.Getenv("AUTH_ENABLED"); enabled != "" {
		config.AuthEnabled = enabled == "true" || enabled == "1"
	}

	// Parse require organization
	if required := os.Getenv("REQUIRE_ORGANIZATION"); required != "" {
		config.RequireOrganization = required == "true" || required == "1"
	}

	return config, nil
}

// Validate checks the configuration for required fields.
func (c *Config) Validate() error {
	if !c.AuthEnabled {
		// No validation needed if auth is disabled
		return nil
	}

	if c.JWTSecret == "" {
		return errors.New("JWT_SECRET is required when authentication is enabled")
	}

	if len(c.JWTSecret) < 44 {
		return errors.New("JWT_SECRET must be at least 44 characters for 256-bit security")
	}

	return nil
}

// NewAuthComponents creates the authentication components from config.
func (c *Config) NewAuthComponents() (*JWTValidator, *IAMClient, error) {
	if !c.AuthEnabled {
		return nil, nil, nil
	}

	validator, err := NewJWTValidator(c.JWTSecret)
	if err != nil {
		return nil, nil, err
	}

	var client *IAMClient
	if c.IAMBaseURL != "" {
		client = NewIAMClient(IAMClientConfig{
			BaseURL: c.IAMBaseURL,
			Timeout: c.RequestTimeout,
		})
	}

	return validator, client, nil
}

// NewMiddleware creates an AuthMiddleware from the config.
func (c *Config) NewMiddleware() (*AuthMiddleware, error) {
	if !c.AuthEnabled {
		// Return nil middleware when auth is disabled
		return nil, nil
	}

	return NewAuthMiddleware(AuthMiddlewareConfig{
		JWTSecret: c.JWTSecret,
		Optional:  false,
	})
}
