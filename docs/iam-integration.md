# IAM Integration

workflow-core integrates with the building-blocks IAM service for authentication and authorization. This document describes how to configure and use IAM integration.

## Overview

The IAM integration provides:

- **JWT Token Validation**: Verify tokens issued by the IAM service
- **Permission-Based Access Control**: Check permissions for workflow operations
- **Multi-Tenant Support**: Organization context from JWT tokens
- **HTTP Middleware**: Easy integration with REST APIs

## Configuration

### Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `IAM_BASE_URL` | Base URL of the IAM service | No* | - |
| `JWT_SECRET` | Secret for JWT validation | Yes** | - |
| `IAM_REQUEST_TIMEOUT` | Timeout for IAM requests | No | 30s |
| `AUTH_ENABLED` | Enable/disable authentication | No | true |
| `REQUIRE_ORGANIZATION` | Require organization context | No | false |

*Required if using IAM client features (login, token refresh)
**Required when `AUTH_ENABLED=true`

### Example Configuration

```bash
export IAM_BASE_URL="http://localhost:3000/iam"
export JWT_SECRET="your-256-bit-secret-key-at-least-44-chars-long"
export AUTH_ENABLED="true"
export REQUIRE_ORGANIZATION="true"
```

## JWT Token Structure

Tokens issued by IAM contain the following claims:

```json
{
  "sub": "user-uuid",
  "email": "user@example.com",
  "type": "access",
  "permissions": ["WORKFLOWS_READ", "WORKFLOWS_EXECUTE"],
  "organizationId": "org-uuid",
  "iat": 1704067200,
  "exp": 1704070800
}
```

## Workflow Permissions

The following permissions control access to workflow operations:

| Permission | Description |
|------------|-------------|
| `WORKFLOWS_CREATE` | Create new workflows |
| `WORKFLOWS_READ` | View workflows |
| `WORKFLOWS_UPDATE` | Modify workflows |
| `WORKFLOWS_DELETE` | Delete workflows |
| `WORKFLOWS_EXECUTE` | Execute workflows |
| `WORKFLOWS_ADMIN` | Full workflow access |
| `EXECUTIONS_READ` | View execution history |
| `EXECUTIONS_CANCEL` | Cancel running executions |
| `EXECUTIONS_RETRY` | Retry failed executions |

## Usage

### Loading Configuration

```go
import "github.com/catalisaio/workflow-core/pkg/auth"

// Load from environment
config, err := auth.LoadConfigFromEnv()
if err != nil {
    log.Fatal(err)
}

// Validate configuration
if err := config.Validate(); err != nil {
    log.Fatal(err)
}
```

### JWT Validation

```go
// Create validator
validator, err := auth.NewJWTValidator(os.Getenv("JWT_SECRET"))
if err != nil {
    log.Fatal(err)
}

// Validate a token
token := "eyJhbG..."
payload, err := validator.Validate(token)
if err != nil {
    log.Printf("Invalid token: %v", err)
    return
}

// Access token claims
fmt.Printf("User: %s\n", payload.Sub)
fmt.Printf("Organization: %s\n", payload.OrganizationID)
fmt.Printf("Permissions: %v\n", payload.Permissions)
```

### HTTP Middleware

```go
import (
    "net/http"
    "github.com/catalisaio/workflow-core/pkg/auth"
)

// Create middleware
middleware, err := auth.NewAuthMiddleware(auth.AuthMiddlewareConfig{
    JWTSecret: os.Getenv("JWT_SECRET"),
})
if err != nil {
    log.Fatal(err)
}

// Protect all routes
http.Handle("/api/", middleware.Handler(apiHandler))

// Require specific permission
executeHandler := middleware.RequirePermission(auth.PermissionWorkflowsExecute)(
    http.HandlerFunc(handleExecute),
)
http.Handle("/api/workflows/execute", middleware.Handler(executeHandler))

// Require organization context
orgHandler := middleware.RequireOrganization()(
    http.HandlerFunc(handleOrgRoute),
)
http.Handle("/api/org/", middleware.Handler(orgHandler))
```

### Accessing Auth Context

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Get full token payload
    token := auth.GetTokenFromContext(ctx)
    if token == nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    // Get specific values
    userID := auth.GetUserIDFromContext(ctx)
    orgID := auth.GetOrganizationIDFromContext(ctx)
    permissions := auth.GetPermissionsFromContext(ctx)
    
    // Check permissions
    checker := auth.NewPermissionChecker(token)
    if !checker.CanExecuteWorkflow() {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
    
    // Process request...
}
```

### IAM Client

For advanced use cases, you can use the IAM client to communicate with the IAM service:

```go
// Create client
client := auth.NewIAMClient(auth.IAMClientConfig{
    BaseURL: os.Getenv("IAM_BASE_URL"),
    Timeout: 30 * time.Second,
})

// Login
resp, err := client.Login(ctx, auth.LoginRequest{
    Email:          "user@example.com",
    Password:       "password",
    OrganizationID: "org-uuid",
})
if err != nil {
    log.Printf("Login failed: %v", err)
    return
}
fmt.Printf("Access Token: %s\n", resp.AccessToken)

// Refresh token
newToken, err := client.RefreshToken(ctx, resp.RefreshToken)
if err != nil {
    log.Printf("Refresh failed: %v", err)
    return
}
```

## Multi-Tenancy

When `REQUIRE_ORGANIZATION=true`, all requests must include an organization context in the JWT token. This enables multi-tenant isolation:

```go
// In your handler
func handleWorkflow(w http.ResponseWriter, r *http.Request) {
    orgID := auth.GetOrganizationIDFromContext(r.Context())
    
    // Use orgID to filter/scope data
    workflows, err := store.ListWorkflows(ctx, orgID)
    // ...
}
```

## Security Considerations

1. **JWT Secret**: Use a cryptographically random secret of at least 44 characters (256-bit). Generate with:
   ```bash
   openssl rand -base64 32
   ```

2. **Token Expiration**: Access tokens should have short expiration times (default: 1 hour in IAM)

3. **HTTPS**: Always use HTTPS in production to protect tokens in transit

4. **Permission Granularity**: Use the most specific permission required for each operation

5. **Organization Isolation**: When using multi-tenancy, always filter data by organization ID

## Error Handling

The middleware returns standard HTTP error codes:

| Code | Description |
|------|-------------|
| 401 Unauthorized | Missing or invalid token |
| 403 Forbidden | Valid token but missing required permission |

Error responses include a message in the response body:

```
Unauthorized: missing Authorization header
Unauthorized: token expired
Forbidden: missing required permission
Forbidden: organization context required
```

## Testing

For testing, you can disable authentication:

```bash
export AUTH_ENABLED=false
```

Or create test tokens using the `auth` package test helpers (see `jwt_test.go`).
