// Package model defines the data structures for n8n-compatible workflow definitions.
package model

import (
	"encoding/json"
)

// Workflow represents an n8n-compatible workflow definition.
type Workflow struct {
	ID          string                 `json:"id,omitempty"`
	Name        string                 `json:"name"`
	Active      bool                   `json:"active"`
	Nodes       []Node                 `json:"nodes"`
	Connections Connections            `json:"connections"`
	Settings    WorkflowSettings       `json:"settings,omitempty"`
	StaticData  map[string]interface{} `json:"staticData,omitempty"`
	PinnedData  map[string]interface{} `json:"pinnedData,omitempty"`
	Tags        []Tag                  `json:"tags,omitempty"`
	Meta        *WorkflowMeta          `json:"meta,omitempty"`
}

// WorkflowSettings contains workflow-level settings.
type WorkflowSettings struct {
	SaveDataSuccessExecution   string `json:"saveDataSuccessExecution,omitempty"`
	SaveDataErrorExecution     string `json:"saveDataErrorExecution,omitempty"`
	SaveManualExecutions       bool   `json:"saveManualExecutions,omitempty"`
	SaveExecutionProgress      bool   `json:"saveExecutionProgress,omitempty"`
	ExecutionTimeout           int    `json:"executionTimeout,omitempty"`
	ErrorWorkflow              string `json:"errorWorkflow,omitempty"`
	CallerPolicy               string `json:"callerPolicy,omitempty"`
	Timezone                   string `json:"timezone,omitempty"`
}

// WorkflowMeta contains metadata about the workflow.
type WorkflowMeta struct {
	InstanceID         string `json:"instanceId,omitempty"`
	TemplateID         string `json:"templateId,omitempty"`
	TemplateCredsSetup bool   `json:"templateCredsSetupCompleted,omitempty"`
}

// Tag represents a workflow tag.
type Tag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Connections maps source node names to their output connections.
// n8n format: map[sourceNodeName]map[outputType][][]ConnectionTarget
// The outer array is for output indices (e.g., true/false for IF)
// The inner array is for multiple connections from the same output
type Connections map[string]map[string][][]ConnectionTarget

// ConnectionTarget describes a single connection target.
type ConnectionTarget struct {
	Node  string `json:"node"`
	Type  string `json:"type"`
	Index int    `json:"index"`
}

// ParseWorkflow parses a JSON byte slice into a Workflow.
func ParseWorkflow(data []byte) (*Workflow, error) {
	var workflow Workflow
	if err := json.Unmarshal(data, &workflow); err != nil {
		return nil, err
	}
	return &workflow, nil
}

// ToJSON converts the workflow to JSON bytes.
func (w *Workflow) ToJSON() ([]byte, error) {
	return json.MarshalIndent(w, "", "  ")
}

// GetNodeByName finds a node by its name.
func (w *Workflow) GetNodeByName(name string) *Node {
	for i := range w.Nodes {
		if w.Nodes[i].Name == name {
			return &w.Nodes[i]
		}
	}
	return nil
}

// GetTriggerNodes returns all trigger nodes in the workflow.
func (w *Workflow) GetTriggerNodes() []Node {
	var triggers []Node
	for _, node := range w.Nodes {
		if node.IsTrigger() {
			triggers = append(triggers, node)
		}
	}
	return triggers
}

// GetConnectedNodes returns the names of nodes connected to the given node's output.
func (w *Workflow) GetConnectedNodes(nodeName string, outputType string) []string {
	nodeConns, ok := w.Connections[nodeName]
	if !ok {
		return nil
	}
	
	outputGroups, ok := nodeConns[outputType]
	if !ok {
		return nil
	}
	
	var nodeNames []string
	for _, targets := range outputGroups {
		for _, target := range targets {
			nodeNames = append(nodeNames, target.Node)
		}
	}
	return nodeNames
}

// GetAllTargets returns all connection targets from a source node.
func (w *Workflow) GetAllTargets(sourceName string) []ConnectionTarget {
	var targets []ConnectionTarget
	
	outputs, ok := w.Connections[sourceName]
	if !ok {
		return targets
	}
	
	for _, outputGroups := range outputs {
		for _, group := range outputGroups {
			targets = append(targets, group...)
		}
	}
	
	return targets
}
