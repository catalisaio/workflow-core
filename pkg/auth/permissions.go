package auth

// Workflow-related permissions.
// These permissions align with the building-blocks IAM permission model.
const (
	// Workflow permissions
	PermissionWorkflowsCreate  = "WORKFLOWS_CREATE"
	PermissionWorkflowsRead    = "WORKFLOWS_READ"
	PermissionWorkflowsUpdate  = "WORKFLOWS_UPDATE"
	PermissionWorkflowsDelete  = "WORKFLOWS_DELETE"
	PermissionWorkflowsExecute = "WORKFLOWS_EXECUTE"
	PermissionWorkflowsAdmin   = "WORKFLOWS_ADMIN"

	// Execution permissions
	PermissionExecutionsRead   = "EXECUTIONS_READ"
	PermissionExecutionsCancel = "EXECUTIONS_CANCEL"
	PermissionExecutionsRetry  = "EXECUTIONS_RETRY"

	// Webhook permissions
	PermissionWebhooksCreate = "WEBHOOKS_SUBSCRIPTIONS_CREATE"
	PermissionWebhooksRead   = "WEBHOOKS_SUBSCRIPTIONS_READ"
	PermissionWebhooksUpdate = "WEBHOOKS_SUBSCRIPTIONS_UPDATE"
	PermissionWebhooksDelete = "WEBHOOKS_SUBSCRIPTIONS_DELETE"
)

// Role represents a predefined role from IAM.
type Role string

const (
	RoleAdmin           Role = "ADMIN"
	RoleWorkflowManager Role = "WORKFLOW_MANAGER"
	RoleViewer          Role = "VIEWER"
)

// DefaultRolePermissions maps roles to their default permissions.
// This should be kept in sync with the IAM service configuration.
var DefaultRolePermissions = map[Role][]string{
	RoleAdmin: {
		PermissionWorkflowsCreate,
		PermissionWorkflowsRead,
		PermissionWorkflowsUpdate,
		PermissionWorkflowsDelete,
		PermissionWorkflowsExecute,
		PermissionWorkflowsAdmin,
		PermissionExecutionsRead,
		PermissionExecutionsCancel,
		PermissionExecutionsRetry,
		PermissionWebhooksCreate,
		PermissionWebhooksRead,
		PermissionWebhooksUpdate,
		PermissionWebhooksDelete,
	},
	RoleWorkflowManager: {
		PermissionWorkflowsCreate,
		PermissionWorkflowsRead,
		PermissionWorkflowsUpdate,
		PermissionWorkflowsExecute,
		PermissionExecutionsRead,
		PermissionExecutionsCancel,
		PermissionExecutionsRetry,
	},
	RoleViewer: {
		PermissionWorkflowsRead,
		PermissionExecutionsRead,
	},
}

// PermissionChecker provides methods for checking permissions.
type PermissionChecker struct {
	permissions []string
}

// NewPermissionChecker creates a new permission checker from a token payload.
func NewPermissionChecker(payload *TokenPayload) *PermissionChecker {
	return &PermissionChecker{
		permissions: payload.Permissions,
	}
}

// CanCreateWorkflow checks if the user can create workflows.
func (c *PermissionChecker) CanCreateWorkflow() bool {
	return c.hasAny(PermissionWorkflowsCreate, PermissionWorkflowsAdmin)
}

// CanReadWorkflow checks if the user can read workflows.
func (c *PermissionChecker) CanReadWorkflow() bool {
	return c.hasAny(PermissionWorkflowsRead, PermissionWorkflowsAdmin)
}

// CanUpdateWorkflow checks if the user can update workflows.
func (c *PermissionChecker) CanUpdateWorkflow() bool {
	return c.hasAny(PermissionWorkflowsUpdate, PermissionWorkflowsAdmin)
}

// CanDeleteWorkflow checks if the user can delete workflows.
func (c *PermissionChecker) CanDeleteWorkflow() bool {
	return c.hasAny(PermissionWorkflowsDelete, PermissionWorkflowsAdmin)
}

// CanExecuteWorkflow checks if the user can execute workflows.
func (c *PermissionChecker) CanExecuteWorkflow() bool {
	return c.hasAny(PermissionWorkflowsExecute, PermissionWorkflowsAdmin)
}

// CanReadExecutions checks if the user can view execution history.
func (c *PermissionChecker) CanReadExecutions() bool {
	return c.hasAny(PermissionExecutionsRead, PermissionWorkflowsAdmin)
}

// CanCancelExecution checks if the user can cancel running executions.
func (c *PermissionChecker) CanCancelExecution() bool {
	return c.hasAny(PermissionExecutionsCancel, PermissionWorkflowsAdmin)
}

// CanRetryExecution checks if the user can retry failed executions.
func (c *PermissionChecker) CanRetryExecution() bool {
	return c.hasAny(PermissionExecutionsRetry, PermissionWorkflowsAdmin)
}

// CanManageWebhooks checks if the user can manage webhooks.
func (c *PermissionChecker) CanManageWebhooks() bool {
	return c.hasAny(PermissionWebhooksCreate, PermissionWebhooksUpdate, PermissionWebhooksDelete, PermissionWorkflowsAdmin)
}

// hasAny checks if any of the given permissions is present.
func (c *PermissionChecker) hasAny(perms ...string) bool {
	for _, required := range perms {
		for _, have := range c.permissions {
			if have == required {
				return true
			}
		}
	}
	return false
}
