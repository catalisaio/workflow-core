package nodes

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/catalisaio/workflow-core/pkg/types"
)

// CodeNode implements the Code/Function node.
// This is a simplified implementation that handles basic JavaScript-like expressions.
type CodeNode struct{}

// Type returns the n8n node type.
func (n *CodeNode) Type() string {
	return "n8n-nodes-base.code"
}

// Execute runs the code and returns the result.
func (n *CodeNode) Execute(ctx *types.ExecutionContext, params map[string]interface{}, input []types.NodeData) ([]types.NodeData, error) {
	if len(input) == 0 {
		input = types.EmptyNodeData()
	}

	mode := getStringParam(params, "mode", "runOnceForAllItems")
	jsCode := getStringParam(params, "jsCode", "")
	
	if jsCode == "" {
		// Try alternative parameter names
		jsCode = getStringParam(params, "code", "")
	}

	if jsCode == "" {
		return input, nil
	}

	switch mode {
	case "runOnceForAllItems":
		return n.executeForAllItems(ctx, jsCode, input)
	case "runOnceForEachItem":
		return n.executeForEachItem(ctx, jsCode, input)
	default:
		return n.executeForAllItems(ctx, jsCode, input)
	}
}

// executeForAllItems runs code once with all items.
func (n *CodeNode) executeForAllItems(ctx *types.ExecutionContext, code string, input []types.NodeData) ([]types.NodeData, error) {
	// Create items array for the code
	items := make([]map[string]interface{}, len(input))
	for i, item := range input {
		items[i] = map[string]interface{}{
			"json": item.JSON,
		}
	}

	// Execute the simplified code interpreter
	result, err := n.interpretCode(ctx, code, items, input)
	if err != nil {
		return nil, fmt.Errorf("code execution failed: %w", err)
	}

	return result, nil
}

// executeForEachItem runs code once for each item.
func (n *CodeNode) executeForEachItem(ctx *types.ExecutionContext, code string, input []types.NodeData) ([]types.NodeData, error) {
	var results []types.NodeData

	for i, item := range input {
		items := []map[string]interface{}{
			{"json": item.JSON},
		}

		result, err := n.interpretCode(ctx, code, items, []types.NodeData{item})
		if err != nil {
			return nil, fmt.Errorf("code execution failed for item %d: %w", i, err)
		}

		results = append(results, result...)
	}

	return results, nil
}

// interpretCode provides a simple code interpretation.
// This handles basic patterns common in n8n code nodes.
func (n *CodeNode) interpretCode(ctx *types.ExecutionContext, code string, items []map[string]interface{}, input []types.NodeData) ([]types.NodeData, error) {
	// Set up evaluator
	if len(input) > 0 && input[0].JSON != nil {
		ctx.Evaluator().SetData(input[0].JSON)
	}

	// Check for common patterns

	// Pattern: return items;
	if strings.TrimSpace(code) == "return items;" || strings.TrimSpace(code) == "return items" {
		return input, nil
	}

	// Pattern: return $input.all();
	if strings.Contains(code, "return $input.all()") {
		return input, nil
	}

	// Pattern: return [{ json: { ... } }];
	returnPattern := regexp.MustCompile(`return\s*\[\s*\{\s*json\s*:\s*(\{[^}]+\})\s*\}\s*\]`)
	if matches := returnPattern.FindStringSubmatch(code); len(matches) > 1 {
		jsonStr := matches[1]
		// Replace JavaScript-style object notation with JSON
		jsonStr = n.jsObjectToJSON(jsonStr)
		
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &result); err == nil {
			return []types.NodeData{{JSON: result}}, nil
		}
	}

	// Pattern: items.map(...)
	if strings.Contains(code, "items.map") {
		// Simple map operation - try to extract the mapping
		return n.handleMapOperation(ctx, code, input)
	}

	// Pattern: items.filter(...)
	if strings.Contains(code, "items.filter") {
		return n.handleFilterOperation(ctx, code, input)
	}

	// Pattern: for loop modifications
	if strings.Contains(code, "for (") || strings.Contains(code, "for(") {
		return n.handleForLoop(ctx, code, input)
	}

	// Default: return input unchanged
	return input, nil
}

// jsObjectToJSON converts JavaScript object notation to JSON.
func (n *CodeNode) jsObjectToJSON(js string) string {
	// Add quotes around unquoted keys
	keyPattern := regexp.MustCompile(`(\w+)\s*:`)
	js = keyPattern.ReplaceAllString(js, `"$1":`)
	
	// Replace single quotes with double quotes for values
	js = strings.ReplaceAll(js, "'", "\"")
	
	return js
}

// handleMapOperation handles simple map operations.
func (n *CodeNode) handleMapOperation(ctx *types.ExecutionContext, code string, input []types.NodeData) ([]types.NodeData, error) {
	// Look for pattern: items.map(item => ({ json: { ... } }))
	// or: items.map(item => { return { json: item.json }; })
	
	// For now, apply a simple transformation based on detected patterns
	var results []types.NodeData

	for _, item := range input {
		// Look for field assignments in the map
		if strings.Contains(code, "item.json") {
			results = append(results, item)
		} else {
			results = append(results, item)
		}
	}

	if len(results) == 0 {
		return input, nil
	}
	return results, nil
}

// handleFilterOperation handles simple filter operations.
func (n *CodeNode) handleFilterOperation(ctx *types.ExecutionContext, code string, input []types.NodeData) ([]types.NodeData, error) {
	// Look for filter condition
	// For now, return all items
	return input, nil
}

// handleForLoop handles for loop modifications.
func (n *CodeNode) handleForLoop(ctx *types.ExecutionContext, code string, input []types.NodeData) ([]types.NodeData, error) {
	// Look for modifications in the loop
	// Pattern: item.json.newField = value
	fieldPattern := regexp.MustCompile(`item\.json\.(\w+)\s*=\s*["']?([^"';\n]+)["']?`)
	matches := fieldPattern.FindAllStringSubmatch(code, -1)

	var results []types.NodeData
	for _, item := range input {
		newJSON := make(map[string]interface{})
		for k, v := range item.JSON {
			newJSON[k] = v
		}

		// Apply field assignments
		for _, match := range matches {
			if len(match) >= 3 {
				fieldName := match[1]
				value := strings.TrimSpace(match[2])
				
				// Try to parse value
				if strings.HasPrefix(value, "item.json.") {
					// Reference to another field
					refField := strings.TrimPrefix(value, "item.json.")
					if refVal, ok := item.JSON[refField]; ok {
						newJSON[fieldName] = refVal
					}
				} else if num, err := strconv.ParseFloat(value, 64); err == nil {
					newJSON[fieldName] = num
				} else if value == "true" {
					newJSON[fieldName] = true
				} else if value == "false" {
					newJSON[fieldName] = false
				} else {
					newJSON[fieldName] = value
				}
			}
		}

		results = append(results, types.NodeData{JSON: newJSON})
	}

	if len(results) == 0 {
		return input, nil
	}
	return results, nil
}
