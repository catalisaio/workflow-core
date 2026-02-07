// Package types provides shared types used across the workflow-core packages.
package types

import (
	"context"
	"sync"
	"time"

	"github.com/catalisaio/workflow-core/pkg/expression"
	"github.com/catalisaio/workflow-core/pkg/model"
)

// NodeData represents the data passed between nodes.
type NodeData struct {
	JSON   map[string]interface{} `json:"json"`
	Binary map[string]interface{} `json:"binary,omitempty"`
	Error  *NodeError             `json:"error,omitempty"`
	Paused bool                   `json:"paused,omitempty"`
}

// NodeError represents an error that occurred during node execution.
type NodeError struct {
	Message     string                 `json:"message"`
	Description string                 `json:"description,omitempty"`
	Stack       string                 `json:"stack,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// ExecutionStatus represents the status of an execution.
type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "pending"
	StatusRunning   ExecutionStatus = "running"
	StatusSuccess   ExecutionStatus = "success"
	StatusError     ExecutionStatus = "error"
	StatusCanceled  ExecutionStatus = "canceled"
	StatusWaiting   ExecutionStatus = "waiting"
)

// NodeExecutionResult holds the result of a single node execution.
type NodeExecutionResult struct {
	NodeName    string          `json:"nodeName"`
	NodeType    string          `json:"nodeType"`
	Status      ExecutionStatus `json:"status"`
	StartTime   time.Time       `json:"startTime"`
	EndTime     time.Time       `json:"endTime"`
	Duration    time.Duration   `json:"duration"`
	OutputData  []NodeData      `json:"outputData,omitempty"`
	Error       *NodeError      `json:"error,omitempty"`
	OutputIndex int             `json:"outputIndex,omitempty"`
}

// ExecutionContext holds the state for a workflow execution.
type ExecutionContext struct {
	ctx          context.Context
	cancel       context.CancelFunc
	workflow     *model.Workflow
	evaluator    *expression.Evaluator
	
	// Node results
	nodeResults  map[string]*NodeExecutionResult
	nodeOutputs  map[string][]NodeData
	
	// Execution state
	status       ExecutionStatus
	startTime    time.Time
	endTime      time.Time
	
	// Environment and config
	env          map[string]string
	variables    map[string]interface{}
	
	// Webhook response handling
	webhookResponse     interface{}
	webhookResponseCode int
	webhookResponseSet  bool
	
	mu           sync.RWMutex
}

// NewExecutionContext creates a new execution context.
func NewExecutionContext(ctx context.Context, workflow *model.Workflow) *ExecutionContext {
	ctx, cancel := context.WithCancel(ctx)
	return &ExecutionContext{
		ctx:         ctx,
		cancel:      cancel,
		workflow:    workflow,
		evaluator:   expression.NewEvaluator(),
		nodeResults: make(map[string]*NodeExecutionResult),
		nodeOutputs: make(map[string][]NodeData),
		status:      StatusPending,
		env:         make(map[string]string),
		variables:   make(map[string]interface{}),
	}
}

// Context returns the underlying context.
func (ec *ExecutionContext) Context() context.Context {
	return ec.ctx
}

// Cancel cancels the execution.
func (ec *ExecutionContext) Cancel() {
	ec.cancel()
	ec.SetStatus(StatusCanceled)
}

// SetStatus sets the execution status.
func (ec *ExecutionContext) SetStatus(status ExecutionStatus) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.status = status
}

// Status returns the current execution status.
func (ec *ExecutionContext) Status() ExecutionStatus {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	return ec.status
}

// SetEnvironment sets environment variables.
func (ec *ExecutionContext) SetEnvironment(env map[string]string) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.env = env
	ec.evaluator.SetEnv(env)
}

// GetEnvironment returns a copy of environment variables.
func (ec *ExecutionContext) GetEnvironment() map[string]string {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	result := make(map[string]string)
	for k, v := range ec.env {
		result[k] = v
	}
	return result
}

// SetVariable sets a workflow variable.
func (ec *ExecutionContext) SetVariable(key string, value interface{}) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.variables[key] = value
}

// GetVariable returns a workflow variable.
func (ec *ExecutionContext) GetVariable(key string) (interface{}, bool) {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	val, ok := ec.variables[key]
	return val, ok
}

// SetNodeResult stores the result of a node execution.
func (ec *ExecutionContext) SetNodeResult(nodeName string, result *NodeExecutionResult) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.nodeResults[nodeName] = result
}

// GetNodeResult returns the result of a node execution.
func (ec *ExecutionContext) GetNodeResult(nodeName string) (*NodeExecutionResult, bool) {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	result, ok := ec.nodeResults[nodeName]
	return result, ok
}

// SetNodeOutput stores the output data for a node.
func (ec *ExecutionContext) SetNodeOutput(nodeName string, data []NodeData) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.nodeOutputs[nodeName] = data
	
	// Update evaluator with node data
	if len(data) > 0 {
		ec.evaluator.SetNodeData(nodeName, data[0].JSON)
	}
}

// GetNodeOutput returns the output data for a node.
func (ec *ExecutionContext) GetNodeOutput(nodeName string) ([]NodeData, bool) {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	data, ok := ec.nodeOutputs[nodeName]
	return data, ok
}

// Evaluator returns the expression evaluator.
func (ec *ExecutionContext) Evaluator() *expression.Evaluator {
	return ec.evaluator
}

// Workflow returns the workflow being executed.
func (ec *ExecutionContext) Workflow() *model.Workflow {
	return ec.workflow
}

// SetStartTime records the start time.
func (ec *ExecutionContext) SetStartTime(t time.Time) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.startTime = t
}

// SetEndTime records the end time.
func (ec *ExecutionContext) SetEndTime(t time.Time) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.endTime = t
}

// Duration returns the execution duration.
func (ec *ExecutionContext) Duration() time.Duration {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	if ec.endTime.IsZero() {
		return time.Since(ec.startTime)
	}
	return ec.endTime.Sub(ec.startTime)
}

// SetWebhookResponse sets the response to be sent for a webhook.
func (ec *ExecutionContext) SetWebhookResponse(data interface{}, statusCode int) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.webhookResponse = data
	ec.webhookResponseCode = statusCode
	ec.webhookResponseSet = true
}

// GetWebhookResponse returns the webhook response if set.
func (ec *ExecutionContext) GetWebhookResponse() (interface{}, int, bool) {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	return ec.webhookResponse, ec.webhookResponseCode, ec.webhookResponseSet
}

// GetAllNodeResults returns all node execution results.
func (ec *ExecutionContext) GetAllNodeResults() map[string]*NodeExecutionResult {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	result := make(map[string]*NodeExecutionResult)
	for k, v := range ec.nodeResults {
		result[k] = v
	}
	return result
}

// MergeInputData combines input data from multiple source nodes.
func MergeInputData(inputs ...[]NodeData) []NodeData {
	var result []NodeData
	for _, input := range inputs {
		result = append(result, input...)
	}
	return result
}

// EmptyNodeData returns empty node data for triggers.
func EmptyNodeData() []NodeData {
	return []NodeData{{
		JSON: make(map[string]interface{}),
	}}
}

// NodeExecutor defines the interface for node execution.
type NodeExecutor interface {
	// Execute runs the node with the given parameters and input data.
	Execute(ctx *ExecutionContext, params map[string]interface{}, input []NodeData) ([]NodeData, error)
	
	// Type returns the n8n node type identifier.
	Type() string
}
