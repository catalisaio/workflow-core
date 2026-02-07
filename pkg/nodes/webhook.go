package nodes

import (
	"github.com/catalisaio/workflow-core/pkg/types"
)

// WebhookNode implements the webhook trigger node.
type WebhookNode struct{}

// Type returns the n8n node type.
func (n *WebhookNode) Type() string {
	return "n8n-nodes-base.webhook"
}

// Execute processes webhook data - it passes through the webhook input data.
func (n *WebhookNode) Execute(ctx *types.ExecutionContext, params map[string]interface{}, input []types.NodeData) ([]types.NodeData, error) {
	// Webhook node receives data from HTTP request
	// The input data should contain the webhook payload
	if len(input) == 0 {
		return types.EmptyNodeData(), nil
	}

	// Process webhook parameters
	httpMethod := getStringParam(params, "httpMethod", "GET")
	path := getStringParam(params, "path", "")
	responseMode := getStringParam(params, "responseMode", "onReceived")

	// Add webhook metadata to the output
	for i := range input {
		if input[i].JSON == nil {
			input[i].JSON = make(map[string]interface{})
		}
		input[i].JSON["$webhook"] = map[string]interface{}{
			"httpMethod":   httpMethod,
			"path":         path,
			"responseMode": responseMode,
		}
	}

	return input, nil
}

// WebhookConfig represents webhook configuration.
type WebhookConfig struct {
	HTTPMethod         string
	Path               string
	ResponseMode       string
	ResponseCode       int
	ResponseData       string
	ResponseContentType string
}

// GetWebhookConfig extracts webhook configuration from parameters.
func GetWebhookConfig(params map[string]interface{}) WebhookConfig {
	return WebhookConfig{
		HTTPMethod:          getStringParam(params, "httpMethod", "GET"),
		Path:                getStringParam(params, "path", ""),
		ResponseMode:        getStringParam(params, "responseMode", "onReceived"),
		ResponseCode:        getIntParam(params, "responseCode", 200),
		ResponseData:        getStringParam(params, "responseData", ""),
		ResponseContentType: getStringParam(params, "responseContentType", "application/json"),
	}
}
