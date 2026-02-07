package nodes

import (
	"encoding/json"

	"github.com/catalisaio/workflow-core/pkg/types"
)

// RespondToWebhookNode implements the Respond to Webhook node.
type RespondToWebhookNode struct{}

// Type returns the n8n node type.
func (n *RespondToWebhookNode) Type() string {
	return "n8n-nodes-base.respondToWebhook"
}

// Execute sets the webhook response.
func (n *RespondToWebhookNode) Execute(ctx *types.ExecutionContext, params map[string]interface{}, input []types.NodeData) ([]types.NodeData, error) {
	if len(input) == 0 {
		input = types.EmptyNodeData()
	}

	// Get response configuration
	respondWith := getStringParam(params, "respondWith", "firstIncomingItem")
	responseCode := getIntParam(params, "responseCode", 200)
	responseHeaders := getMapParam(params, "responseHeaders")
	
	var responseData interface{}

	switch respondWith {
	case "firstIncomingItem":
		if len(input) > 0 {
			responseData = input[0].JSON
		}
	case "allIncomingItems":
		items := make([]map[string]interface{}, len(input))
		for i, item := range input {
			items[i] = item.JSON
		}
		responseData = items
	case "text":
		responseBody := getStringParam(params, "responseBody", "")
		// Evaluate expressions
		ctx.Evaluator().SetData(input[0].JSON)
		evaluated, err := ctx.Evaluator().Evaluate(responseBody)
		if err == nil {
			responseData = evaluated
		} else {
			responseData = responseBody
		}
	case "json":
		responseBody := getStringParam(params, "responseBody", "{}")
		// Try to parse as JSON
		var jsonData interface{}
		if err := json.Unmarshal([]byte(responseBody), &jsonData); err == nil {
			responseData = jsonData
		} else {
			responseData = responseBody
		}
	case "noData":
		responseData = nil
	default:
		if len(input) > 0 {
			responseData = input[0].JSON
		}
	}

	// Set the webhook response in context
	ctx.SetWebhookResponse(map[string]interface{}{
		"statusCode": responseCode,
		"headers":    responseHeaders,
		"body":       responseData,
	}, responseCode)

	// Pass through input
	return input, nil
}
