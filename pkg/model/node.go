package model

import (
	"strings"
)

// Node represents a single node in the workflow.
type Node struct {
	ID           string                 `json:"id,omitempty"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	TypeVersion  float64                `json:"typeVersion,omitempty"`
	Position     []float64              `json:"position"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	Credentials  map[string]Credential  `json:"credentials,omitempty"`
	Disabled     bool                   `json:"disabled,omitempty"`
	Notes        string                 `json:"notes,omitempty"`
	NotesInFlow  bool                   `json:"notesInFlow,omitempty"`
	RetryOnFail  bool                   `json:"retryOnFail,omitempty"`
	MaxTries     int                    `json:"maxTries,omitempty"`
	WaitBetweenTries int                `json:"waitBetweenTries,omitempty"`
	AlwaysOutputData bool               `json:"alwaysOutputData,omitempty"`
	ExecuteOnce  bool                   `json:"executeOnce,omitempty"`
	ContinueOnFail bool                 `json:"continueOnFail,omitempty"`
	Webhooks     []WebhookDescription   `json:"webhooks,omitempty"`
	WebhookId    string                 `json:"webhookId,omitempty"`
}

// Credential represents node credentials reference.
type Credential struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// WebhookDescription describes a webhook for a node.
type WebhookDescription struct {
	Name               string `json:"name"`
	HTTPMethod         string `json:"httpMethod"`
	Path               string `json:"path"`
	ResponseMode       string `json:"responseMode,omitempty"`
	IsFullPath         bool   `json:"isFullPath,omitempty"`
	RestartWebhook     bool   `json:"restartWebhook,omitempty"`
	ResponseCode       int    `json:"responseCode,omitempty"`
	ResponseData       string `json:"responseData,omitempty"`
	ResponseContentType string `json:"responseContentType,omitempty"`
}

// TriggerTypes lists known trigger node types.
var TriggerTypes = []string{
	"n8n-nodes-base.manualTrigger",
	"n8n-nodes-base.webhook",
	"n8n-nodes-base.scheduleTrigger",
	"n8n-nodes-base.cronTrigger",
	"n8n-nodes-base.cron",
	"n8n-nodes-base.start",
	"n8n-nodes-base.executeWorkflowTrigger",
}

// IsTrigger returns true if the node is a trigger node.
func (n *Node) IsTrigger() bool {
	nodeType := strings.ToLower(n.Type)
	for _, triggerType := range TriggerTypes {
		if strings.ToLower(triggerType) == nodeType {
			return true
		}
	}
	// Also check for trigger suffix
	return strings.HasSuffix(nodeType, "trigger")
}

// GetParameter returns a parameter value by key.
func (n *Node) GetParameter(key string) (interface{}, bool) {
	if n.Parameters == nil {
		return nil, false
	}
	val, ok := n.Parameters[key]
	return val, ok
}

// GetStringParameter returns a string parameter or default value.
func (n *Node) GetStringParameter(key string, defaultVal string) string {
	val, ok := n.GetParameter(key)
	if !ok {
		return defaultVal
	}
	if str, ok := val.(string); ok {
		return str
	}
	return defaultVal
}

// GetBoolParameter returns a bool parameter or default value.
func (n *Node) GetBoolParameter(key string, defaultVal bool) bool {
	val, ok := n.GetParameter(key)
	if !ok {
		return defaultVal
	}
	if b, ok := val.(bool); ok {
		return b
	}
	return defaultVal
}

// GetIntParameter returns an int parameter or default value.
func (n *Node) GetIntParameter(key string, defaultVal int) int {
	val, ok := n.GetParameter(key)
	if !ok {
		return defaultVal
	}
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	}
	return defaultVal
}

// Clone creates a deep copy of the node.
func (n *Node) Clone() *Node {
	clone := *n
	
	// Clone position
	if n.Position != nil {
		clone.Position = make([]float64, len(n.Position))
		copy(clone.Position, n.Position)
	}
	
	// Clone parameters
	if n.Parameters != nil {
		clone.Parameters = make(map[string]interface{})
		for k, v := range n.Parameters {
			clone.Parameters[k] = v
		}
	}
	
	// Clone credentials
	if n.Credentials != nil {
		clone.Credentials = make(map[string]Credential)
		for k, v := range n.Credentials {
			clone.Credentials[k] = v
		}
	}
	
	return &clone
}
