package nodes

import (
	"fmt"

	"github.com/catalisaio/workflow-core/pkg/types"
)

// SwitchNode implements the Switch node for multi-way branching.
type SwitchNode struct{}

// Type returns the n8n node type.
func (n *SwitchNode) Type() string {
	return "n8n-nodes-base.switch"
}

// Execute evaluates conditions and routes data to matching output.
func (n *SwitchNode) Execute(ctx *types.ExecutionContext, params map[string]interface{}, input []types.NodeData) ([]types.NodeData, error) {
	if len(input) == 0 {
		return types.EmptyNodeData(), nil
	}

	mode := getStringParam(params, "mode", "rules")
	
	switch mode {
	case "rules":
		return n.executeRulesMode(ctx, params, input)
	case "expression":
		return n.executeExpressionMode(ctx, params, input)
	default:
		return n.executeRulesMode(ctx, params, input)
	}
}

// executeRulesMode executes switch with rule-based routing.
func (n *SwitchNode) executeRulesMode(ctx *types.ExecutionContext, params map[string]interface{}, input []types.NodeData) ([]types.NodeData, error) {
	rules := getSliceParam(params, "rules")
	if rules == nil {
		// Try v2 format
		if rulesObj, ok := params["rules"].(map[string]interface{}); ok {
			if rulesList, ok := rulesObj["rules"].([]interface{}); ok {
				rules = rulesList
			}
		}
	}

	// Process each input item
	var matchedOutput []types.NodeData
	
	for _, item := range input {
		ctx.Evaluator().SetData(item.JSON)
		
		matched := false
		for _, rule := range rules {
			ruleMap, ok := rule.(map[string]interface{})
			if !ok {
				continue
			}

			// Evaluate rule conditions
			conditions := getSliceParam(ruleMap, "conditions")
			if conditions == nil {
				if condObj, ok := ruleMap["conditions"].(map[string]interface{}); ok {
					if condList, ok := condObj["conditions"].([]interface{}); ok {
						conditions = condList
					}
				}
			}

			if n.evaluateRuleConditions(ctx, conditions) {
				matchedOutput = append(matchedOutput, item)
				matched = true
				break // First match wins
			}
		}

		// If no rule matched, check for fallback
		if !matched {
			fallbackOutput := getStringParam(params, "fallbackOutput", "none")
			if fallbackOutput != "none" {
				matchedOutput = append(matchedOutput, item)
			}
		}
	}

	if len(matchedOutput) == 0 {
		return types.EmptyNodeData(), nil
	}
	return matchedOutput, nil
}

// executeExpressionMode executes switch with expression-based routing.
func (n *SwitchNode) executeExpressionMode(ctx *types.ExecutionContext, params map[string]interface{}, input []types.NodeData) ([]types.NodeData, error) {
	output := getStringParam(params, "output", "")
	
	var matchedOutput []types.NodeData
	
	for _, item := range input {
		ctx.Evaluator().SetData(item.JSON)
		
		// Evaluate the output expression
		result, err := ctx.Evaluator().Evaluate(output)
		if err != nil {
			continue
		}
		
		// Result should be an output index or name
		outputStr := fmt.Sprintf("%v", result)
		if outputStr != "" && outputStr != "0" {
			matchedOutput = append(matchedOutput, item)
		}
	}

	if len(matchedOutput) == 0 {
		return types.EmptyNodeData(), nil
	}
	return matchedOutput, nil
}

// evaluateRuleConditions evaluates all conditions in a rule.
func (n *SwitchNode) evaluateRuleConditions(ctx *types.ExecutionContext, conditions []interface{}) bool {
	if len(conditions) == 0 {
		return false
	}

	for _, cond := range conditions {
		condMap, ok := cond.(map[string]interface{})
		if !ok {
			continue
		}

		leftValue := condMap["leftValue"]
		rightValue := condMap["rightValue"]
		operator := getStringParam(condMap, "operator", "equals")

		// Evaluate expressions
		left, err := ctx.Evaluator().Evaluate(leftValue)
		if err != nil {
			left = leftValue
		}

		right, err := ctx.Evaluator().Evaluate(rightValue)
		if err != nil {
			right = rightValue
		}

		result, err := ctx.Evaluator().EvaluateCondition(left, operator, right)
		if err != nil || !result {
			return false
		}
	}

	return true
}
