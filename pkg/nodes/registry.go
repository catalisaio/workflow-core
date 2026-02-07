// Package nodes provides node implementations for the workflow engine.
package nodes

import (
	"sync"

	"github.com/catalisaio/workflow-core/pkg/model"
	"github.com/catalisaio/workflow-core/pkg/types"
)

// Registry holds all registered node executors.
type Registry struct {
	executors map[string]types.NodeExecutor
	mu        sync.RWMutex
}

// NewRegistry creates a new node registry with default nodes registered.
func NewRegistry() *Registry {
	r := &Registry{
		executors: make(map[string]types.NodeExecutor),
	}
	
	// Register default nodes
	r.RegisterDefaults()
	
	return r
}

// Register adds a node executor to the registry.
func (r *Registry) Register(executor types.NodeExecutor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.executors[executor.Type()] = executor
}

// Get retrieves a node executor by type.
func (r *Registry) Get(nodeType string) (types.NodeExecutor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	executor, ok := r.executors[nodeType]
	return executor, ok
}

// Has checks if a node type is registered.
func (r *Registry) Has(nodeType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.executors[nodeType]
	return ok
}

// List returns all registered node types.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	types := make([]string, 0, len(r.executors))
	for t := range r.executors {
		types = append(types, t)
	}
	return types
}

// RegisterDefaults registers all built-in node types.
func (r *Registry) RegisterDefaults() {
	// Trigger nodes
	r.Register(&ManualTriggerNode{})
	r.Register(&WebhookNode{})
	r.Register(&CronNode{})
	r.Register(&ScheduleTriggerNode{})
	
	// Data manipulation nodes
	r.Register(&SetNode{})
	r.Register(&MergeNode{})
	r.Register(&SplitInBatchesNode{})
	
	// Control flow nodes
	r.Register(&IfNode{})
	r.Register(&SwitchNode{})
	r.Register(&NoOpNode{})
	
	// HTTP nodes
	r.Register(&HTTPRequestNode{})
	r.Register(&RespondToWebhookNode{})
	
	// Code execution
	r.Register(&CodeNode{})
	
	// Debug
	r.Register(&DebugNode{})
}

// Unregister removes a node executor from the registry.
func (r *Registry) Unregister(nodeType string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.executors, nodeType)
}

// GetSupportedNodes returns information about all supported nodes.
func (r *Registry) GetSupportedNodes() []model.Node {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var nodes []model.Node
	for nodeType := range r.executors {
		nodes = append(nodes, model.Node{
			Type: nodeType,
		})
	}
	return nodes
}
