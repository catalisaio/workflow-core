package parser

import (
	"strings"
	"testing"
)

func TestParseValidWorkflow(t *testing.T) {
	workflowJSON := `{
		"name": "Test Workflow",
		"nodes": [
			{
				"name": "Start",
				"type": "n8n-nodes-base.manualTrigger",
				"typeVersion": 1,
				"position": [100, 200],
				"parameters": {}
			},
			{
				"name": "Set Data",
				"type": "n8n-nodes-base.set",
				"typeVersion": 1,
				"position": [300, 200],
				"parameters": {
					"values": {
						"string": [{"name": "key", "value": "value"}]
					}
				}
			}
		],
		"connections": {
			"Start": {
				"main": [[{"node": "Set Data", "type": "main", "index": 0}]]
			}
		}
	}`

	parser := NewParser()
	workflow, err := parser.Parse([]byte(workflowJSON))
	if err != nil {
		t.Fatalf("Failed to parse valid workflow: %v", err)
	}

	if workflow.Name != "Test Workflow" {
		t.Errorf("Expected workflow name 'Test Workflow', got '%s'", workflow.Name)
	}

	if len(workflow.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(workflow.Nodes))
	}

	if workflow.Nodes[0].Name != "Start" {
		t.Errorf("Expected first node name 'Start', got '%s'", workflow.Nodes[0].Name)
	}
}

func TestParseEmptyNodes(t *testing.T) {
	workflowJSON := `{
		"name": "Empty Workflow",
		"nodes": [],
		"connections": {}
	}`

	parser := NewParser()
	_, err := parser.Parse([]byte(workflowJSON))
	if err == nil {
		t.Error("Expected error for workflow with no nodes")
	}
	if !strings.Contains(err.Error(), "no nodes") {
		t.Errorf("Expected 'no nodes' error, got: %v", err)
	}
}

func TestParseDuplicateNodeNames(t *testing.T) {
	workflowJSON := `{
		"name": "Duplicate Names",
		"nodes": [
			{"name": "Node1", "type": "test", "position": [0, 0]},
			{"name": "Node1", "type": "test", "position": [100, 0]}
		],
		"connections": {}
	}`

	parser := NewParser()
	_, err := parser.Parse([]byte(workflowJSON))
	if err == nil {
		t.Error("Expected error for duplicate node names")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("Expected 'duplicate' error, got: %v", err)
	}
}

func TestParseInvalidJSON(t *testing.T) {
	parser := NewParser()
	_, err := parser.Parse([]byte("not valid json"))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestParseWithStrictMode(t *testing.T) {
	workflowJSON := `{
		"name": "Bad Connections",
		"nodes": [
			{"name": "Start", "type": "test", "position": [0, 0]}
		],
		"connections": {
			"NonExistent": {
				"main": [[{"node": "Start", "type": "main", "index": 0}]]
			}
		}
	}`

	parser := NewParser()
	parser.SetStrictMode(true)
	
	_, err := parser.Parse([]byte(workflowJSON))
	if err == nil {
		t.Error("Expected error for invalid connection reference in strict mode")
	}
}

func TestParseReader(t *testing.T) {
	workflowJSON := `{
		"name": "Reader Test",
		"nodes": [
			{"name": "Start", "type": "n8n-nodes-base.manualTrigger", "position": [0, 0]}
		],
		"connections": {}
	}`

	parser := NewParser()
	reader := strings.NewReader(workflowJSON)
	workflow, err := parser.ParseReader(reader)
	if err != nil {
		t.Fatalf("Failed to parse from reader: %v", err)
	}

	if workflow.Name != "Reader Test" {
		t.Errorf("Expected workflow name 'Reader Test', got '%s'", workflow.Name)
	}
}

func TestExtractNodeTypes(t *testing.T) {
	workflowJSON := `{
		"name": "Type Test",
		"nodes": [
			{"name": "Start", "type": "n8n-nodes-base.manualTrigger", "position": [0, 0]},
			{"name": "Set1", "type": "n8n-nodes-base.set", "position": [100, 0]},
			{"name": "Set2", "type": "n8n-nodes-base.set", "position": [200, 0]},
			{"name": "HTTP", "type": "n8n-nodes-base.httpRequest", "position": [300, 0]}
		],
		"connections": {}
	}`

	parser := NewParser()
	workflow, err := parser.Parse([]byte(workflowJSON))
	if err != nil {
		t.Fatalf("Failed to parse workflow: %v", err)
	}

	types := ExtractNodeTypes(workflow)
	if len(types) != 3 {
		t.Errorf("Expected 3 unique node types, got %d", len(types))
	}
}

func TestParseMultipleOutputConnections(t *testing.T) {
	// n8n IF node has multiple outputs (true/false)
	workflowJSON := `{
		"name": "IF Test",
		"nodes": [
			{"name": "Start", "type": "n8n-nodes-base.manualTrigger", "position": [0, 0]},
			{"name": "IF", "type": "n8n-nodes-base.if", "position": [100, 0]},
			{"name": "TrueNode", "type": "n8n-nodes-base.noOp", "position": [200, 0]},
			{"name": "FalseNode", "type": "n8n-nodes-base.noOp", "position": [200, 100]}
		],
		"connections": {
			"Start": {
				"main": [[{"node": "IF", "type": "main", "index": 0}]]
			},
			"IF": {
				"main": [
					[{"node": "TrueNode", "type": "main", "index": 0}],
					[{"node": "FalseNode", "type": "main", "index": 0}]
				]
			}
		}
	}`

	parser := NewParser()
	workflow, err := parser.Parse([]byte(workflowJSON))
	if err != nil {
		t.Fatalf("Failed to parse workflow with multiple outputs: %v", err)
	}

	// Check that IF node has two output groups
	ifOutputs := workflow.Connections["IF"]["main"]
	if len(ifOutputs) != 2 {
		t.Errorf("Expected 2 output groups for IF node, got %d", len(ifOutputs))
	}
}
