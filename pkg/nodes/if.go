package nodes

import (
	"fmt"
	"strings"

	"github.com/catalisaio/workflow-core/pkg/types"
)

// IfNode implements the IF conditional node.
type IfNode struct{}

// Type returns the n8n node type.
func (n *IfNode) Type() string {
	return "n8n-nodes-base.if"
}

// Execute evaluates conditions and routes data to true or false output.
func (n *IfNode) Execute(ctx *types.ExecutionContext, params map[string]interface{}, input []types.NodeData) ([]types.NodeData, error) {
	if len(input) == 0 {
		return types.EmptyNodeData(), nil
	}

	var trueOutput []types.NodeData
	var falseOutput []types.NodeData

	for _, item := range input {
		// Set up evaluator with current item data
		ctx.Evaluator().SetData(item.JSON)

		// Evaluate conditions
		result, err := n.evaluateConditions(ctx, params)
		if err != nil {
			// On error, send to false output
			falseOutput = append(falseOutput, item)
			continue
		}

		if result {
			trueOutput = append(trueOutput, item)
		} else {
			falseOutput = append(falseOutput, item)
		}
	}

	// For IF node, we return both outputs
	// The engine will handle routing based on connection types
	// Convention: index 0 = true, index 1 = false
	if len(trueOutput) > 0 {
		return trueOutput, nil
	}
	return falseOutput, nil
}

// evaluateConditions evaluates the IF conditions.
func (n *IfNode) evaluateConditions(ctx *types.ExecutionContext, params map[string]interface{}) (bool, error) {
	// Handle v2 format (conditions)
	if conditions, ok := params["conditions"].(map[string]interface{}); ok {
		return n.evaluateConditionsV2(ctx, conditions)
	}

	// Handle v1 format (conditions as array)
	if conditionsArray, ok := params["conditions"].([]interface{}); ok {
		return n.evaluateConditionsV1(ctx, conditionsArray, params)
	}

	// Handle simple boolean condition
	if condition, ok := params["condition"].(map[string]interface{}); ok {
		return n.evaluateSingleCondition(ctx, condition)
	}

	return false, fmt.Errorf("no valid conditions found")
}

// evaluateConditionsV2 evaluates v2 format conditions.
func (n *IfNode) evaluateConditionsV2(ctx *types.ExecutionContext, conditions map[string]interface{}) (bool, error) {
	// Get the options (AND/OR)
	combineOperation := "and"
	if options, ok := conditions["options"].(map[string]interface{}); ok {
		if op, ok := options["combineOperation"].(string); ok {
			combineOperation = strings.ToLower(op)
		}
	}

	// Get condition groups
	conditionGroups, ok := conditions["conditions"].([]interface{})
	if !ok {
		return false, fmt.Errorf("invalid conditions structure")
	}

	results := make([]bool, 0)

	for _, group := range conditionGroups {
		groupMap, ok := group.(map[string]interface{})
		if !ok {
			continue
		}

		result, err := n.evaluateSingleCondition(ctx, groupMap)
		if err != nil {
			results = append(results, false)
			continue
		}
		results = append(results, result)
	}

	if len(results) == 0 {
		return false, nil
	}

	// Combine results
	if combineOperation == "or" {
		for _, r := range results {
			if r {
				return true, nil
			}
		}
		return false, nil
	}

	// Default: AND
	for _, r := range results {
		if !r {
			return false, nil
		}
	}
	return true, nil
}

// evaluateConditionsV1 evaluates v1 format conditions.
func (n *IfNode) evaluateConditionsV1(ctx *types.ExecutionContext, conditions []interface{}, params map[string]interface{}) (bool, error) {
	// Get combine operation
	combineOperation := getStringParam(params, "combineOperation", "all")

	results := make([]bool, 0)

	for _, cond := range conditions {
		condMap, ok := cond.(map[string]interface{})
		if !ok {
			continue
		}

		result, err := n.evaluateSingleConditionV1(ctx, condMap)
		if err != nil {
			results = append(results, false)
			continue
		}
		results = append(results, result)
	}

	if len(results) == 0 {
		return false, nil
	}

	// Combine results
	if combineOperation == "any" {
		for _, r := range results {
			if r {
				return true, nil
			}
		}
		return false, nil
	}

	// Default: all (AND)
	for _, r := range results {
		if !r {
			return false, nil
		}
	}
	return true, nil
}

// evaluateSingleCondition evaluates a single v2 condition.
func (n *IfNode) evaluateSingleCondition(ctx *types.ExecutionContext, condition map[string]interface{}) (bool, error) {
	leftValue := condition["leftValue"]
	rightValue := condition["rightValue"]
	operator := getStringParam(condition, "operator", "equals")

	// Evaluate expressions
	left, err := ctx.Evaluator().Evaluate(leftValue)
	if err != nil {
		left = leftValue
	}

	right, err := ctx.Evaluator().Evaluate(rightValue)
	if err != nil {
		right = rightValue
	}

	return ctx.Evaluator().EvaluateCondition(left, operator, right)
}

// evaluateSingleConditionV1 evaluates a single v1 condition.
func (n *IfNode) evaluateSingleConditionV1(ctx *types.ExecutionContext, condition map[string]interface{}) (bool, error) {
	value1 := condition["value1"]
	value2 := condition["value2"]
	operation := getStringParam(condition, "operation", "equals")

	// Evaluate expressions
	left, err := ctx.Evaluator().Evaluate(value1)
	if err != nil {
		left = value1
	}

	right, err := ctx.Evaluator().Evaluate(value2)
	if err != nil {
		right = value2
	}

	return ctx.Evaluator().EvaluateCondition(left, operation, right)
}
