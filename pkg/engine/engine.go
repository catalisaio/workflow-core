package engine

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/catalisaio/workflow-core/pkg/model"
	"github.com/catalisaio/workflow-core/pkg/types"
)

// Engine is the workflow execution engine.
type Engine struct {
	registry NodeRegistry
	options  EngineOptions
}

// NodeRegistry interface for node lookup.
type NodeRegistry interface {
	Get(nodeType string) (types.NodeExecutor, bool)
	Has(nodeType string) bool
	List() []string
}

// EngineOptions configures the engine behavior.
type EngineOptions struct {
	MaxConcurrentNodes int
	DefaultTimeout     time.Duration
	ContinueOnError    bool
	Environment        map[string]string
}

// DefaultEngineOptions returns sensible default options.
func DefaultEngineOptions() EngineOptions {
	return EngineOptions{
		MaxConcurrentNodes: 10,
		DefaultTimeout:     30 * time.Minute,
		ContinueOnError:    false,
		Environment:        make(map[string]string),
	}
}

// NewEngine creates a new workflow engine.
func NewEngine(options EngineOptions) *Engine {
	return &Engine{
		options: options,
	}
}

// SetRegistry sets the node registry.
func (e *Engine) SetRegistry(registry NodeRegistry) {
	e.registry = registry
}

// Registry returns the node registry.
func (e *Engine) Registry() NodeRegistry {
	return e.registry
}

// ExecutionResult holds the complete result of a workflow execution.
type ExecutionResult struct {
	WorkflowName string                          `json:"workflowName"`
	Status       ExecutionStatus                 `json:"status"`
	StartTime    time.Time                       `json:"startTime"`
	EndTime      time.Time                       `json:"endTime"`
	Duration     time.Duration                   `json:"duration"`
	NodeResults  map[string]*NodeExecutionResult `json:"nodeResults"`
	FinalOutput  []NodeData                      `json:"finalOutput,omitempty"`
	Error        error                           `json:"error,omitempty"`
}

// Execute runs a workflow with optional initial input data.
func (e *Engine) Execute(ctx context.Context, workflow *model.Workflow, inputData []NodeData) (*ExecutionResult, error) {
	execCtx := NewExecutionContext(ctx, workflow)
	execCtx.SetEnvironment(e.options.Environment)
	execCtx.SetStartTime(time.Now())
	execCtx.SetStatus(StatusRunning)

	result := &ExecutionResult{
		WorkflowName: workflow.Name,
		StartTime:    time.Now(),
		Status:       StatusRunning,
	}

	// Build execution order
	order, err := e.buildExecutionOrder(workflow)
	if err != nil {
		execCtx.SetStatus(StatusError)
		result.Status = StatusError
		result.Error = err
		return result, err
	}

	// Find trigger nodes
	triggers := workflow.GetTriggerNodes()
	if len(triggers) == 0 {
		// No trigger, use first node in order
		if len(order) > 0 {
			triggers = []model.Node{*workflow.GetNodeByName(order[0])}
		}
	}

	// Set initial data for trigger nodes
	if len(inputData) == 0 {
		inputData = EmptyNodeData()
	}

	for _, trigger := range triggers {
		execCtx.SetNodeOutput(trigger.Name, inputData)
	}

	// Execute nodes in order
	var lastOutput []NodeData
	for _, nodeName := range order {
		select {
		case <-ctx.Done():
			execCtx.SetStatus(StatusCanceled)
			result.Status = StatusCanceled
			result.Error = ctx.Err()
			return result, ctx.Err()
		default:
		}

		node := workflow.GetNodeByName(nodeName)
		if node == nil {
			continue
		}

		// Skip disabled nodes
		if node.Disabled {
			continue
		}

		// Get input data from connected nodes
		nodeInputData := e.getNodeInputData(execCtx, workflow, node)

		// Execute the node
		nodeResult, err := e.executeNode(execCtx, node, nodeInputData)
		execCtx.SetNodeResult(nodeName, nodeResult)

		if err != nil {
			if !e.options.ContinueOnError && !node.ContinueOnFail {
				execCtx.SetStatus(StatusError)
				result.Status = StatusError
				result.Error = fmt.Errorf("node %s failed: %w", nodeName, err)
				result.EndTime = time.Now()
				result.Duration = result.EndTime.Sub(result.StartTime)
				result.NodeResults = execCtx.GetAllNodeResults()
				return result, result.Error
			}
		}

		if nodeResult.OutputData != nil {
			lastOutput = nodeResult.OutputData
			execCtx.SetNodeOutput(nodeName, nodeResult.OutputData)
		}
	}

	execCtx.SetStatus(StatusSuccess)
	execCtx.SetEndTime(time.Now())

	result.Status = StatusSuccess
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.NodeResults = execCtx.GetAllNodeResults()
	result.FinalOutput = lastOutput

	return result, nil
}

// buildExecutionOrder creates a topological sort of nodes.
func (e *Engine) buildExecutionOrder(workflow *model.Workflow) ([]string, error) {
	// Build adjacency list and in-degree map
	inDegree := make(map[string]int)
	adjacency := make(map[string][]string)

	// Initialize all nodes
	for _, node := range workflow.Nodes {
		inDegree[node.Name] = 0
		adjacency[node.Name] = []string{}
	}

	// Build edges from connections
	for sourceName, outputs := range workflow.Connections {
		for _, outputGroups := range outputs {
			for _, targets := range outputGroups {
				for _, target := range targets {
					adjacency[sourceName] = append(adjacency[sourceName], target.Node)
					inDegree[target.Node]++
				}
			}
		}
	}

	// Find all nodes with no incoming edges (starting points)
	var queue []string
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	// Sort the initial queue for deterministic order
	sort.Strings(queue)

	// Kahn's algorithm for topological sort
	var order []string
	for len(queue) > 0 {
		// Take the first node
		node := queue[0]
		queue = queue[1:]
		order = append(order, node)

		// Sort adjacent nodes for deterministic order
		adjacent := adjacency[node]
		sort.Strings(adjacent)

		for _, next := range adjacent {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
				sort.Strings(queue)
			}
		}
	}

	// Check for cycles
	if len(order) != len(workflow.Nodes) {
		return nil, fmt.Errorf("workflow contains cycles")
	}

	return order, nil
}

// getNodeInputData collects input data from nodes connected to this node.
func (e *Engine) getNodeInputData(execCtx *ExecutionContext, workflow *model.Workflow, node *model.Node) []NodeData {
	var inputs []NodeData

	// Find all nodes that connect to this node
	for sourceName, outputs := range workflow.Connections {
		for outputType, outputGroups := range outputs {
			for _, targets := range outputGroups {
				for _, target := range targets {
					if target.Node == node.Name {
						// Get output from the source node
						if sourceOutput, ok := execCtx.GetNodeOutput(sourceName); ok {
							// Check if this is from a specific output index (for branching nodes)
							if outputType == "main" {
								inputs = append(inputs, sourceOutput...)
							} else {
								// Handle named outputs (e.g., "true", "false" for IF nodes)
								inputs = append(inputs, sourceOutput...)
							}
						}
					}
				}
			}
		}
	}

	if len(inputs) == 0 {
		// Check if this is a trigger node
		if node.IsTrigger() {
			if output, ok := execCtx.GetNodeOutput(node.Name); ok {
				return output
			}
		}
		return EmptyNodeData()
	}

	return inputs
}

// executeNode executes a single node.
func (e *Engine) executeNode(execCtx *ExecutionContext, node *model.Node, inputData []NodeData) (*NodeExecutionResult, error) {
	result := &NodeExecutionResult{
		NodeName:  node.Name,
		NodeType:  node.Type,
		Status:    StatusRunning,
		StartTime: time.Now(),
	}

	// Get the executor for this node type
	if e.registry == nil {
		result.Status = StatusError
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		result.Error = &NodeError{
			Message: "no registry configured",
		}
		return result, fmt.Errorf("no registry configured")
	}

	executor, ok := e.registry.Get(node.Type)
	if !ok {
		result.Status = StatusError
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		result.Error = &NodeError{
			Message: fmt.Sprintf("unknown node type: %s", node.Type),
		}
		return result, fmt.Errorf("unknown node type: %s", node.Type)
	}

	// Set up evaluator with input data
	if len(inputData) > 0 && inputData[0].JSON != nil {
		execCtx.Evaluator().SetData(inputData[0].JSON)
	}

	// Execute the node
	output, err := executor.Execute(execCtx, node.Parameters, inputData)

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	if err != nil {
		result.Status = StatusError
		result.Error = &NodeError{
			Message: err.Error(),
		}
		return result, err
	}

	result.Status = StatusSuccess
	result.OutputData = output

	return result, nil
}

// ExecuteWithWebhook starts a workflow execution triggered by a webhook.
func (e *Engine) ExecuteWithWebhook(ctx context.Context, workflow *model.Workflow, webhookData map[string]interface{}) (*ExecutionResult, error) {
	inputData := []NodeData{{
		JSON: webhookData,
	}}
	return e.Execute(ctx, workflow, inputData)
}
