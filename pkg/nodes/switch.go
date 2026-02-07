package nodes

import (
	"fmt"
	"strings"

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
	rules := n.extractRules(params)
	if len(rules) == 0 {
		return types.EmptyNodeData(), nil
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

			conditions := n.extractRuleConditions(ruleMap)

			if n.evaluateRuleConditions(ctx, conditions) {
				matchedOutput = append(matchedOutput, item)
				matched = true
				break // First match wins
			}
		}

		// If no rule matched, check for fallback
		if !matched {
			fallbackOutput := n.extractFallbackOutput(params)
			if fallbackOutput != "none" {
				matchedOutput = append(matchedOutput, item)
			}
		}
	}

	if len(matchedOutput) == 0 {
		return []types.NodeData{}, nil
	}
	return matchedOutput, nil
}

func (n *SwitchNode) extractRules(params map[string]interface{}) []interface{} {
	if rulesObj, ok := params["rules"].(map[string]interface{}); ok {
		if values, ok := rulesObj["values"].([]interface{}); ok {
			return values
		}
		if rulesList, ok := rulesObj["rules"].([]interface{}); ok {
			return rulesList
		}
	}
	if rules, ok := params["rules"].([]interface{}); ok {
		return rules
	}
	return nil
}

func (n *SwitchNode) extractRuleConditions(rule map[string]interface{}) []interface{} {
	if conditionsObj, ok := rule["conditions"].(map[string]interface{}); ok {
		if values, ok := conditionsObj["values"].([]interface{}); ok {
			return values
		}
		if condList, ok := conditionsObj["conditions"].([]interface{}); ok {
			return condList
		}
	}
	if conditions, ok := rule["conditions"].([]interface{}); ok {
		return conditions
	}
	return nil
}

func (n *SwitchNode) extractFallbackOutput(params map[string]interface{}) string {
	if options, ok := params["options"].(map[string]interface{}); ok {
		if fallback, ok := options["fallbackOutput"]; ok {
			return strings.TrimSpace(fmt.Sprintf("%v", fallback))
		}
	}
	if fallback, ok := params["fallbackOutput"]; ok {
		return strings.TrimSpace(fmt.Sprintf("%v", fallback))
	}
	return "none"
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
		return []types.NodeData{}, nil
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
		operator := "equals"
		if op, ok := condMap["operator"].(string); ok {
			operator = op
		}
		if opObj, ok := condMap["operator"].(map[string]interface{}); ok {
			operator = getStringParam(opObj, "operation", operator)
		}

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
