// Package engine provides the workflow execution engine.
package engine

import (
	"github.com/catalisaio/workflow-core/pkg/types"
)

// Re-export types for convenience
type (
	NodeData            = types.NodeData
	NodeError           = types.NodeError
	ExecutionStatus     = types.ExecutionStatus
	NodeExecutionResult = types.NodeExecutionResult
	ExecutionContext    = types.ExecutionContext
)

// Re-export constants
const (
	StatusPending  = types.StatusPending
	StatusRunning  = types.StatusRunning
	StatusSuccess  = types.StatusSuccess
	StatusError    = types.StatusError
	StatusCanceled = types.StatusCanceled
	StatusWaiting  = types.StatusWaiting
)

// Re-export functions
var (
	NewExecutionContext = types.NewExecutionContext
	MergeInputData      = types.MergeInputData
	EmptyNodeData       = types.EmptyNodeData
)
