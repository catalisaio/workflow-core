// Package expression provides expression evaluation for n8n-style expressions.
package expression

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Evaluator evaluates n8n-style expressions.
type Evaluator struct {
	data      map[string]interface{}
	nodeData  map[string]interface{}
	env       map[string]string
}

// NewEvaluator creates a new expression evaluator.
func NewEvaluator() *Evaluator {
	return &Evaluator{
		data:     make(map[string]interface{}),
		nodeData: make(map[string]interface{}),
		env:      make(map[string]string),
	}
}

// SetData sets the input data ($json).
func (e *Evaluator) SetData(data map[string]interface{}) {
	e.data = data
}

// SetNodeData sets the node data for $node references.
func (e *Evaluator) SetNodeData(nodeName string, data interface{}) {
	e.nodeData[nodeName] = data
}

// SetEnv sets environment variables.
func (e *Evaluator) SetEnv(env map[string]string) {
	e.env = env
}

// expressionPattern matches {{ expression }} syntax.
var expressionPattern = regexp.MustCompile(`\{\{\s*(.+?)\s*\}\}`)

// Evaluate evaluates expressions in a string value.
func (e *Evaluator) Evaluate(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		return e.evaluateString(v)
	case map[string]interface{}:
		return e.evaluateMap(v)
	case []interface{}:
		return e.evaluateSlice(v)
	default:
		return value, nil
	}
}

// evaluateString evaluates expressions in a string.
func (e *Evaluator) evaluateString(s string) (interface{}, error) {
	// Check if the entire string is a single expression
	if strings.HasPrefix(strings.TrimSpace(s), "={{") || strings.HasPrefix(strings.TrimSpace(s), "{{") {
		trimmed := strings.TrimSpace(s)
		trimmed = strings.TrimPrefix(trimmed, "=")
		
		matches := expressionPattern.FindStringSubmatch(trimmed)
		if len(matches) == 2 && "{{"+matches[1]+"}}" == trimmed || "{{ "+matches[1]+" }}" == trimmed {
			// Single expression, return the actual value type
			return e.evaluateExpression(matches[1])
		}
	}

	// Replace all expressions in the string
	result := expressionPattern.ReplaceAllStringFunc(s, func(match string) string {
		matches := expressionPattern.FindStringSubmatch(match)
		if len(matches) != 2 {
			return match
		}
		
		val, err := e.evaluateExpression(matches[1])
		if err != nil {
			return match
		}
		
		return e.toString(val)
	})

	return result, nil
}

// evaluateMap evaluates expressions in a map.
func (e *Evaluator) evaluateMap(m map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for k, v := range m {
		evaluated, err := e.Evaluate(v)
		if err != nil {
			return nil, fmt.Errorf("error evaluating key %s: %w", k, err)
		}
		result[k] = evaluated
	}
	return result, nil
}

// evaluateSlice evaluates expressions in a slice.
func (e *Evaluator) evaluateSlice(s []interface{}) ([]interface{}, error) {
	result := make([]interface{}, len(s))
	for i, v := range s {
		evaluated, err := e.Evaluate(v)
		if err != nil {
			return nil, fmt.Errorf("error evaluating index %d: %w", i, err)
		}
		result[i] = evaluated
	}
	return result, nil
}

// evaluateExpression evaluates a single expression.
func (e *Evaluator) evaluateExpression(expr string) (interface{}, error) {
	expr = strings.TrimSpace(expr)

	// Handle $now
	if expr == "$now" {
		return time.Now().Format(time.RFC3339), nil
	}

	// Handle $timestamp
	if expr == "$timestamp" {
		return time.Now().Unix(), nil
	}

	// Handle $env
	if strings.HasPrefix(expr, "$env.") {
		key := strings.TrimPrefix(expr, "$env.")
		if val, ok := e.env[key]; ok {
			return val, nil
		}
		return "", nil
	}

	// Handle $env["key"]
	if strings.HasPrefix(expr, "$env[") {
		key := e.extractBracketKey(expr, "$env")
		if key != "" {
			if val, ok := e.env[key]; ok {
				return val, nil
			}
		}
		return "", nil
	}

	// Handle $json
	if strings.HasPrefix(expr, "$json") {
		return e.evaluateJSONPath(expr, e.data)
	}

	// Handle $input
	if strings.HasPrefix(expr, "$input") {
		// $input refers to the same as $json in most contexts
		path := strings.TrimPrefix(expr, "$input")
		if strings.HasPrefix(path, ".item") {
			path = strings.TrimPrefix(path, ".item")
		}
		if strings.HasPrefix(path, ".all") {
			// Return all items
			return e.data, nil
		}
		return e.evaluateJSONPath("$json"+path, e.data)
	}

	// Handle $node
	if strings.HasPrefix(expr, "$node[") {
		return e.evaluateNodeReference(expr)
	}

	// Handle simple literals
	if val, err := strconv.ParseFloat(expr, 64); err == nil {
		return val, nil
	}
	if val, err := strconv.ParseBool(expr); err == nil {
		return val, nil
	}
	if strings.HasPrefix(expr, "'") && strings.HasSuffix(expr, "'") {
		return strings.Trim(expr, "'"), nil
	}
	if strings.HasPrefix(expr, "\"") && strings.HasSuffix(expr, "\"") {
		return strings.Trim(expr, "\""), nil
	}

	// Return as-is if we can't evaluate
	return expr, nil
}

// evaluateJSONPath evaluates a $json.path or $json["path"] expression.
func (e *Evaluator) evaluateJSONPath(expr string, data map[string]interface{}) (interface{}, error) {
	path := strings.TrimPrefix(expr, "$json")
	
	if path == "" {
		return data, nil
	}

	current := interface{}(data)

	for path != "" {
		if strings.HasPrefix(path, ".") {
			// Dot notation: .fieldName
			path = strings.TrimPrefix(path, ".")
			idx := strings.IndexAny(path, ".[")
			var key string
			if idx == -1 {
				key = path
				path = ""
			} else {
				key = path[:idx]
				path = path[idx:]
			}

			if m, ok := current.(map[string]interface{}); ok {
				current = m[key]
			} else {
				return nil, fmt.Errorf("cannot access property %s on non-object", key)
			}
		} else if strings.HasPrefix(path, "[") {
			// Bracket notation: ["fieldName"] or [index]
			endIdx := strings.Index(path, "]")
			if endIdx == -1 {
				return nil, fmt.Errorf("unclosed bracket in expression")
			}
			
			inner := path[1:endIdx]
			path = path[endIdx+1:]

			// String key
			if (strings.HasPrefix(inner, "'") && strings.HasSuffix(inner, "'")) ||
				(strings.HasPrefix(inner, "\"") && strings.HasSuffix(inner, "\"")) {
				key := inner[1 : len(inner)-1]
				if m, ok := current.(map[string]interface{}); ok {
					current = m[key]
				} else {
					return nil, fmt.Errorf("cannot access property %s on non-object", key)
				}
			} else {
				// Numeric index
				idx, err := strconv.Atoi(inner)
				if err != nil {
					// Treat as string key without quotes
					if m, ok := current.(map[string]interface{}); ok {
						current = m[inner]
					} else {
						return nil, fmt.Errorf("invalid index: %s", inner)
					}
				} else {
					if arr, ok := current.([]interface{}); ok {
						if idx >= 0 && idx < len(arr) {
							current = arr[idx]
						} else {
							return nil, fmt.Errorf("array index out of bounds: %d", idx)
						}
					} else {
						return nil, fmt.Errorf("cannot use index on non-array")
					}
				}
			}
		} else {
			return nil, fmt.Errorf("unexpected character in path: %s", path)
		}
	}

	return current, nil
}

// evaluateNodeReference evaluates a $node["nodeName"].json reference.
func (e *Evaluator) evaluateNodeReference(expr string) (interface{}, error) {
	// Extract node name from $node["nodeName"]
	key := e.extractBracketKey(expr, "$node")
	if key == "" {
		return nil, fmt.Errorf("invalid node reference: %s", expr)
	}

	nodeData, ok := e.nodeData[key]
	if !ok {
		return nil, fmt.Errorf("node data not found: %s", key)
	}

	// Check for .json suffix
	remaining := expr[strings.Index(expr, "]")+1:]
	if strings.HasPrefix(remaining, ".json") {
		remaining = strings.TrimPrefix(remaining, ".json")
		if remaining == "" {
			return nodeData, nil
		}
		// Further path traversal
		if m, ok := nodeData.(map[string]interface{}); ok {
			return e.evaluateJSONPath("$json"+remaining, m)
		}
	}

	return nodeData, nil
}

// extractBracketKey extracts the key from prefix["key"] or prefix['key'].
func (e *Evaluator) extractBracketKey(expr, prefix string) string {
	s := strings.TrimPrefix(expr, prefix+"[")
	endIdx := strings.Index(s, "]")
	if endIdx == -1 {
		return ""
	}
	inner := s[:endIdx]
	if (strings.HasPrefix(inner, "'") && strings.HasSuffix(inner, "'")) ||
		(strings.HasPrefix(inner, "\"") && strings.HasSuffix(inner, "\"")) {
		return inner[1 : len(inner)-1]
	}
	return inner
}

// toString converts a value to string representation.
func (e *Evaluator) toString(val interface{}) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return v
	case float64:
		if v == float64(int(v)) {
			return strconv.Itoa(int(v))
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case bool:
		return strconv.FormatBool(v)
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

// EvaluateCondition evaluates a simple condition expression.
func (e *Evaluator) EvaluateCondition(left interface{}, operator string, right interface{}) (bool, error) {
	// Evaluate expressions in left and right
	leftVal, err := e.Evaluate(left)
	if err != nil {
		return false, err
	}
	rightVal, err := e.Evaluate(right)
	if err != nil {
		return false, err
	}

	switch operator {
	case "equals", "==", "equal":
		return e.compareEquals(leftVal, rightVal), nil
	case "notEquals", "!=", "notEqual":
		return !e.compareEquals(leftVal, rightVal), nil
	case "contains":
		return e.compareContains(leftVal, rightVal), nil
	case "notContains":
		return !e.compareContains(leftVal, rightVal), nil
	case "startsWith":
		return e.compareStartsWith(leftVal, rightVal), nil
	case "endsWith":
		return e.compareEndsWith(leftVal, rightVal), nil
	case "larger", ">":
		return e.compareLarger(leftVal, rightVal), nil
	case "smaller", "<":
		return e.compareSmaller(leftVal, rightVal), nil
	case "largerEqual", ">=":
		return e.compareLarger(leftVal, rightVal) || e.compareEquals(leftVal, rightVal), nil
	case "smallerEqual", "<=":
		return e.compareSmaller(leftVal, rightVal) || e.compareEquals(leftVal, rightVal), nil
	case "isEmpty":
		return e.compareIsEmpty(leftVal), nil
	case "isNotEmpty":
		return !e.compareIsEmpty(leftVal), nil
	case "exists":
		return leftVal != nil, nil
	case "notExists":
		return leftVal == nil, nil
	case "isTrue":
		return e.isTruthy(leftVal), nil
	case "isFalse":
		return !e.isTruthy(leftVal), nil
	default:
		return false, fmt.Errorf("unknown operator: %s", operator)
	}
}

func (e *Evaluator) compareEquals(left, right interface{}) bool {
	return fmt.Sprintf("%v", left) == fmt.Sprintf("%v", right)
}

func (e *Evaluator) compareContains(left, right interface{}) bool {
	return strings.Contains(fmt.Sprintf("%v", left), fmt.Sprintf("%v", right))
}

func (e *Evaluator) compareStartsWith(left, right interface{}) bool {
	return strings.HasPrefix(fmt.Sprintf("%v", left), fmt.Sprintf("%v", right))
}

func (e *Evaluator) compareEndsWith(left, right interface{}) bool {
	return strings.HasSuffix(fmt.Sprintf("%v", left), fmt.Sprintf("%v", right))
}

func (e *Evaluator) compareLarger(left, right interface{}) bool {
	l := e.toFloat(left)
	r := e.toFloat(right)
	return l > r
}

func (e *Evaluator) compareSmaller(left, right interface{}) bool {
	l := e.toFloat(left)
	r := e.toFloat(right)
	return l < r
}

func (e *Evaluator) compareIsEmpty(val interface{}) bool {
	if val == nil {
		return true
	}
	switch v := val.(type) {
	case string:
		return v == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	}
	return false
}

func (e *Evaluator) isTruthy(val interface{}) bool {
	if val == nil {
		return false
	}
	switch v := val.(type) {
	case bool:
		return v
	case string:
		return v != "" && v != "false" && v != "0"
	case float64:
		return v != 0
	case int:
		return v != 0
	}
	return true
}

func (e *Evaluator) toFloat(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	}
	return 0
}
