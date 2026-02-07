package nodes

import (
	"fmt"
	"strings"

	"github.com/catalisaio/workflow-core/pkg/types"
)

// SetNode implements the Set node for data manipulation.
type SetNode struct{}

// Type returns the n8n node type.
func (n *SetNode) Type() string {
	return "n8n-nodes-base.set"
}

// Execute sets values on the data.
func (n *SetNode) Execute(ctx *types.ExecutionContext, params map[string]interface{}, input []types.NodeData) ([]types.NodeData, error) {
	if len(input) == 0 {
		input = types.EmptyNodeData()
	}

	mode := getStringParam(params, "mode", "manual")
	keepOnlySet := getBoolParam(params, "keepOnlySet", false)

	var results []types.NodeData

	for _, item := range input {
		ctx.Evaluator().SetData(item.JSON)

		var newJSON map[string]interface{}

		if keepOnlySet {
			newJSON = make(map[string]interface{})
		} else {
			// Clone existing data
			newJSON = make(map[string]interface{})
			for k, v := range item.JSON {
				newJSON[k] = v
			}
		}

		// Handle different modes
		switch mode {
		case "manual":
			n.processManualMode(ctx, params, newJSON)
		case "raw":
			n.processRawMode(ctx, params, newJSON)
		default:
			// Try to process as manual mode by default
			n.processManualMode(ctx, params, newJSON)
		}

		results = append(results, types.NodeData{
			JSON:   newJSON,
			Binary: item.Binary,
		})
	}

	return results, nil
}

// processManualMode handles the manual assignment mode.
func (n *SetNode) processManualMode(ctx *types.ExecutionContext, params map[string]interface{}, data map[string]interface{}) {
	// Handle v3 format (assignments)
	if assignments, ok := params["assignments"].(map[string]interface{}); ok {
		if assignmentList, ok := assignments["assignments"].([]interface{}); ok {
			for _, a := range assignmentList {
				assignment, ok := a.(map[string]interface{})
				if !ok {
					continue
				}

				name := getStringParam(assignment, "name", "")
				if name == "" {
					continue
				}

				value := assignment["value"]

				if strVal, ok := value.(string); ok && isExplicitExpression(strVal) {
					evaluated, err := ctx.Evaluator().Evaluate(strVal)
					if err == nil {
						value = evaluated
					}
				}

				data[name] = value
			}
		}
	}

	// Handle v1/v2 format (values)
	if values, ok := params["values"].(map[string]interface{}); ok {
		// String values
		if stringVals, ok := values["string"].([]interface{}); ok {
			for _, sv := range stringVals {
				strVal, ok := sv.(map[string]interface{})
				if !ok {
					continue
				}
				name := getStringParam(strVal, "name", "")
				if name == "" {
					continue
				}
				value := strVal["value"]
				if strValue, ok := value.(string); ok && isExplicitExpression(strValue) {
					evaluated, err := ctx.Evaluator().Evaluate(strValue)
					if err == nil {
						value = evaluated
					}
				}
				data[name] = value
			}
		}

		// Number values
		if numberVals, ok := values["number"].([]interface{}); ok {
			for _, nv := range numberVals {
				numVal, ok := nv.(map[string]interface{})
				if !ok {
					continue
				}
				name := getStringParam(numVal, "name", "")
				if name == "" {
					continue
				}
				data[name] = numVal["value"]
			}
		}

		// Boolean values
		if boolVals, ok := values["boolean"].([]interface{}); ok {
			for _, bv := range boolVals {
				boolVal, ok := bv.(map[string]interface{})
				if !ok {
					continue
				}
				name := getStringParam(boolVal, "name", "")
				if name == "" {
					continue
				}
				data[name] = boolVal["value"]
			}
		}
	}

	// Handle options
	if options, ok := params["options"].(map[string]interface{}); ok {
		if dotNotation, ok := options["dotNotation"].(bool); ok && dotNotation {
			// TODO: Support dot notation for nested fields
			_ = dotNotation
		}
	}
}

// processRawMode handles raw JSON mode.
func (n *SetNode) processRawMode(ctx *types.ExecutionContext, params map[string]interface{}, data map[string]interface{}) {
	if jsonOutput, ok := params["jsonOutput"].(string); ok {
		// Evaluate as expression
		result, err := ctx.Evaluator().Evaluate(jsonOutput)
		if err == nil {
			if resultMap, ok := result.(map[string]interface{}); ok {
				for k, v := range resultMap {
					data[k] = v
				}
			}
		}
	}
}

// Helper functions

func getStringParam(params map[string]interface{}, key, defaultVal string) string {
	if val, ok := params[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultVal
}

func getBoolParam(params map[string]interface{}, key string, defaultVal bool) bool {
	if val, ok := params[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultVal
}

func getIntParam(params map[string]interface{}, key string, defaultVal int) int {
	if val, ok := params[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return defaultVal
}

func getFloatParam(params map[string]interface{}, key string, defaultVal float64) float64 {
	if val, ok := params[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return defaultVal
}

func getMapParam(params map[string]interface{}, key string) map[string]interface{} {
	if val, ok := params[key]; ok {
		if m, ok := val.(map[string]interface{}); ok {
			return m
		}
	}
	return nil
}

func getSliceParam(params map[string]interface{}, key string) []interface{} {
	if val, ok := params[key]; ok {
		if s, ok := val.([]interface{}); ok {
			return s
		}
	}
	return nil
}

func interfaceToString(val interface{}) string {
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%v", val)
}

func isExplicitExpression(value string) bool {
	trimmed := strings.TrimSpace(value)
	return strings.HasPrefix(trimmed, "={{")
}
