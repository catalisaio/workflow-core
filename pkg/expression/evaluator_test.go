package expression

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestEvaluateSimpleJSONPath(t *testing.T) {
	e := NewEvaluator()
	e.SetData(map[string]interface{}{
		"name":  "John",
		"age":   float64(30), // JSON numbers are always float64
		"email": "john@example.com",
	})

	tests := []struct {
		expr     string
		expected interface{}
		isMap    bool
	}{
		{"{{ $json.name }}", "John", false},
		{"{{ $json.age }}", float64(30), false},
		{"{{ $json.email }}", "john@example.com", false},
		{"{{ $json }}", nil, true}, // Map comparison handled separately
	}

	for _, tt := range tests {
		result, err := e.Evaluate(tt.expr)
		if err != nil {
			t.Errorf("Evaluate(%s) error: %v", tt.expr, err)
			continue
		}
		if tt.isMap {
			// Just check it's a map
			if _, ok := result.(map[string]interface{}); !ok {
				t.Errorf("Evaluate(%s) expected map, got %T", tt.expr, result)
			}
			continue
		}
		if result != tt.expected {
			t.Errorf("Evaluate(%s) = %v, want %v", tt.expr, result, tt.expected)
		}
	}
}

func TestEvaluateBracketNotation(t *testing.T) {
	e := NewEvaluator()
	e.SetData(map[string]interface{}{
		"field-with-dash": "value1",
		"nested": map[string]interface{}{
			"inner": "value2",
		},
	})

	tests := []struct {
		expr     string
		expected interface{}
	}{
		{`{{ $json["field-with-dash"] }}`, "value1"},
		{`{{ $json['field-with-dash'] }}`, "value1"},
		{`{{ $json.nested.inner }}`, "value2"},
	}

	for _, tt := range tests {
		result, err := e.Evaluate(tt.expr)
		if err != nil {
			t.Errorf("Evaluate(%s) error: %v", tt.expr, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("Evaluate(%s) = %v, want %v", tt.expr, result, tt.expected)
		}
	}
}

func TestEvaluateNow(t *testing.T) {
	e := NewEvaluator()
	
	result, err := e.Evaluate("{{ $now }}")
	if err != nil {
		t.Fatalf("Evaluate($now) error: %v", err)
	}
	
	resultStr, ok := result.(string)
	if !ok {
		t.Fatalf("Expected string, got %T", result)
	}
	
	// Parse the result to verify it's a valid RFC3339 timestamp
	_, err = time.Parse(time.RFC3339, resultStr)
	if err != nil {
		t.Errorf("Invalid timestamp format: %v", err)
	}
}

func TestEvaluateTimestamp(t *testing.T) {
	e := NewEvaluator()
	
	before := time.Now().Unix()
	result, err := e.Evaluate("{{ $timestamp }}")
	after := time.Now().Unix()
	
	if err != nil {
		t.Fatalf("Evaluate($timestamp) error: %v", err)
	}
	
	ts, ok := result.(int64)
	if !ok {
		t.Fatalf("Expected int64, got %T", result)
	}
	
	if ts < before || ts > after {
		t.Errorf("Timestamp %d not in expected range [%d, %d]", ts, before, after)
	}
}

func TestEvaluateEnv(t *testing.T) {
	e := NewEvaluator()
	e.SetEnv(map[string]string{
		"API_KEY": "secret123",
		"DEBUG":   "true",
	})

	tests := []struct {
		expr     string
		expected interface{}
	}{
		{"{{ $env.API_KEY }}", "secret123"},
		{"{{ $env.DEBUG }}", "true"},
		{`{{ $env["API_KEY"] }}`, "secret123"},
		{"{{ $env.NONEXISTENT }}", ""},
	}

	for _, tt := range tests {
		result, err := e.Evaluate(tt.expr)
		if err != nil {
			t.Errorf("Evaluate(%s) error: %v", tt.expr, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("Evaluate(%s) = %v, want %v", tt.expr, result, tt.expected)
		}
	}
}

func TestEvaluateStringInterpolation(t *testing.T) {
	e := NewEvaluator()
	e.SetData(map[string]interface{}{
		"name": "World",
	})

	result, err := e.Evaluate("Hello, {{ $json.name }}!")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	expected := "Hello, World!"
	if result != expected {
		t.Errorf("Got %v, want %v", result, expected)
	}
}

func TestEvaluateMap(t *testing.T) {
	e := NewEvaluator()
	e.SetData(map[string]interface{}{
		"value": "test",
	})

	input := map[string]interface{}{
		"key1": "{{ $json.value }}",
		"key2": "static",
		"key3": 123,
	}

	result, err := e.Evaluate(input)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map, got %T", result)
	}

	if resultMap["key1"] != "test" {
		t.Errorf("key1 = %v, want 'test'", resultMap["key1"])
	}
	if resultMap["key2"] != "static" {
		t.Errorf("key2 = %v, want 'static'", resultMap["key2"])
	}
}

func TestEvaluateCondition(t *testing.T) {
	e := NewEvaluator()
	e.SetData(map[string]interface{}{
		"count": float64(10),
		"name":  "John",
		"empty": "",
	})

	tests := []struct {
		left     interface{}
		operator string
		right    interface{}
		expected bool
	}{
		{"{{ $json.count }}", "larger", float64(5), true},
		{"{{ $json.count }}", "smaller", float64(5), false},
		{"{{ $json.count }}", "equals", float64(10), true},
		{"{{ $json.name }}", "contains", "oh", true},
		{"{{ $json.name }}", "startsWith", "Jo", true},
		{"{{ $json.name }}", "endsWith", "hn", true},
		{"{{ $json.empty }}", "isEmpty", nil, true},
		{"{{ $json.name }}", "isNotEmpty", nil, true},
	}

	for _, tt := range tests {
		result, err := e.EvaluateCondition(tt.left, tt.operator, tt.right)
		if err != nil {
			t.Errorf("EvaluateCondition(%v, %s, %v) error: %v", tt.left, tt.operator, tt.right, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("EvaluateCondition(%v, %s, %v) = %v, want %v", tt.left, tt.operator, tt.right, result, tt.expected)
		}
	}
}

func TestEvaluateNodeReference(t *testing.T) {
	e := NewEvaluator()
	e.SetNodeData("PreviousNode", map[string]interface{}{
		"result": "success",
		"data": map[string]interface{}{
			"id": float64(123),
		},
	})

	tests := []struct {
		expr     string
		expected interface{}
	}{
		{`{{ $node["PreviousNode"].json.result }}`, "success"},
		{`{{ $node["PreviousNode"].json.data.id }}`, float64(123)},
	}

	for _, tt := range tests {
		result, err := e.Evaluate(tt.expr)
		if err != nil {
			t.Errorf("Evaluate(%s) error: %v", tt.expr, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("Evaluate(%s) = %v, want %v", tt.expr, result, tt.expected)
		}
	}
}

func TestEvaluateArrayAccess(t *testing.T) {
	e := NewEvaluator()
	e.SetData(map[string]interface{}{
		"items": []interface{}{"first", "second", "third"},
		"nested": map[string]interface{}{
			"list": []interface{}{
				map[string]interface{}{"name": "item1"},
				map[string]interface{}{"name": "item2"},
			},
		},
	})

	tests := []struct {
		expr     string
		expected interface{}
	}{
		{"{{ $json.items[0] }}", "first"},
		{"{{ $json.items[1] }}", "second"},
		{"{{ $json.nested.list[0].name }}", "item1"},
	}

	for _, tt := range tests {
		result, err := e.Evaluate(tt.expr)
		if err != nil {
			t.Errorf("Evaluate(%s) error: %v", tt.expr, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("Evaluate(%s) = %v, want %v", tt.expr, result, tt.expected)
		}
	}
}

func TestEvaluateLiterals(t *testing.T) {
	e := NewEvaluator()

	tests := []struct {
		expr     string
		expected interface{}
	}{
		{"{{ 42 }}", float64(42)},
		{"{{ 3.14 }}", float64(3.14)},
		{"{{ true }}", true},
		{"{{ false }}", false},
		{`{{ 'hello' }}`, "hello"},
		{`{{ "world" }}`, "world"},
	}

	for _, tt := range tests {
		result, err := e.Evaluate(tt.expr)
		if err != nil {
			t.Errorf("Evaluate(%s) error: %v", tt.expr, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("Evaluate(%s) = %v, want %v", tt.expr, result, tt.expected)
		}
	}
}

func TestEvaluateInput(t *testing.T) {
	e := NewEvaluator()
	e.SetData(map[string]interface{}{
		"key": "value",
	})

	result, err := e.Evaluate("{{ $input.item.key }}")
	if err != nil {
		t.Fatalf("Evaluate error: %v", err)
	}
	
	resultStr, ok := result.(string)
	if !ok {
		t.Fatalf("Expected string, got %T", result)
	}
	if !strings.Contains(resultStr, "value") {
		t.Errorf("Expected result to contain 'value', got %v", resultStr)
	}
}

func TestEvaluateDeepEqual(t *testing.T) {
	e := NewEvaluator()
	e.SetData(map[string]interface{}{
		"obj": map[string]interface{}{
			"nested": "value",
		},
	})

	result, err := e.Evaluate("{{ $json.obj }}")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	expected := map[string]interface{}{"nested": "value"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Got %v, want %v", result, expected)
	}
}
