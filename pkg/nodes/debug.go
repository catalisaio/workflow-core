package nodes

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/catalisaio/workflow-core/pkg/types"
)

// DebugNode implements a debug/logging node.
type DebugNode struct{}

// Type returns the n8n node type.
func (n *DebugNode) Type() string {
	return "n8n-nodes-base.debug"
}

// Execute logs the input data for debugging purposes.
func (n *DebugNode) Execute(ctx *types.ExecutionContext, params map[string]interface{}, input []types.NodeData) ([]types.NodeData, error) {
	if len(input) == 0 {
		input = types.EmptyNodeData()
	}

	// Get options
	logOutput := getBoolParam(params, "logOutput", true)
	logLevel := getStringParam(params, "logLevel", "info")
	message := getStringParam(params, "message", "")

	if logOutput {
		for i, item := range input {
			// Format the data
			jsonData, err := json.MarshalIndent(item.JSON, "", "  ")
			if err != nil {
				jsonData = []byte(fmt.Sprintf("%v", item.JSON))
			}

			// Build log message
			logMsg := fmt.Sprintf("[DEBUG] Item %d", i)
			if message != "" {
				// Evaluate message expression
				ctx.Evaluator().SetData(item.JSON)
				evaluated, err := ctx.Evaluator().Evaluate(message)
				if err == nil {
					logMsg = fmt.Sprintf("[DEBUG] %v - Item %d", evaluated, i)
				} else {
					logMsg = fmt.Sprintf("[DEBUG] %s - Item %d", message, i)
				}
			}

			// Log based on level
			switch logLevel {
			case "error":
				log.Printf("[ERROR] %s: %s", logMsg, string(jsonData))
			case "warn", "warning":
				log.Printf("[WARN] %s: %s", logMsg, string(jsonData))
			case "debug":
				log.Printf("[DEBUG] %s: %s", logMsg, string(jsonData))
			default:
				log.Printf("[INFO] %s: %s", logMsg, string(jsonData))
			}
		}
	}

	// Pass through input unchanged
	return input, nil
}
