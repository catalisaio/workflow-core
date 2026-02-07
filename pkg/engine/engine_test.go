package engine

import (
	"context"
	"testing"
	"time"

	"github.com/catalisaio/workflow-core/pkg/model"
	"github.com/catalisaio/workflow-core/pkg/nodes"
)

func TestBuildExecutionOrder(t *testing.T) {
	workflow := &model.Workflow{
		Name: "Test Workflow",
		Nodes: []model.Node{
			{Name: "Start", Type: "n8n-nodes-base.manualTrigger"},
			{Name: "Set1", Type: "n8n-nodes-base.set"},
			{Name: "Set2", Type: "n8n-nodes-base.set"},
			{Name: "End", Type: "n8n-nodes-base.noOp"},
		},
		Connections: model.Connections{
			"Start": {
				"main": [][]model.ConnectionTarget{{{Node: "Set1", Type: "main", Index: 0}}},
			},
			"Set1": {
				"main": [][]model.ConnectionTarget{{{Node: "Set2", Type: "main", Index: 0}}},
			},
			"Set2": {
				"main": [][]model.ConnectionTarget{{{Node: "End", Type: "main", Index: 0}}},
			},
		},
	}

	engine := NewEngine(DefaultEngineOptions())
	order, err := engine.buildExecutionOrder(workflow)
	if err != nil {
		t.Fatalf("buildExecutionOrder failed: %v", err)
	}

	if len(order) != 4 {
		t.Errorf("Expected 4 nodes in order, got %d", len(order))
	}

	// Verify order - Start must come before Set1, Set1 before Set2, Set2 before End
	nodeIndex := make(map[string]int)
	for i, name := range order {
		nodeIndex[name] = i
	}

	if nodeIndex["Start"] >= nodeIndex["Set1"] {
		t.Error("Start should come before Set1")
	}
	if nodeIndex["Set1"] >= nodeIndex["Set2"] {
		t.Error("Set1 should come before Set2")
	}
	if nodeIndex["Set2"] >= nodeIndex["End"] {
		t.Error("Set2 should come before End")
	}
}

func TestBuildExecutionOrderCycle(t *testing.T) {
	workflow := &model.Workflow{
		Name: "Cyclic Workflow",
		Nodes: []model.Node{
			{Name: "A", Type: "test"},
			{Name: "B", Type: "test"},
			{Name: "C", Type: "test"},
		},
		Connections: model.Connections{
			"A": {"main": [][]model.ConnectionTarget{{{Node: "B"}}}},
			"B": {"main": [][]model.ConnectionTarget{{{Node: "C"}}}},
			"C": {"main": [][]model.ConnectionTarget{{{Node: "A"}}}}, // Cycle!
		},
	}

	engine := NewEngine(DefaultEngineOptions())
	_, err := engine.buildExecutionOrder(workflow)
	if err == nil {
		t.Error("Expected error for cyclic workflow")
	}
}

func TestExecuteSimpleWorkflow(t *testing.T) {
	workflow := &model.Workflow{
		Name: "Simple Workflow",
		Nodes: []model.Node{
			{
				Name:       "Start",
				Type:       "n8n-nodes-base.manualTrigger",
				Parameters: map[string]interface{}{},
			},
			{
				Name: "Set Data",
				Type: "n8n-nodes-base.set",
				Parameters: map[string]interface{}{
					"mode": "manual",
					"assignments": map[string]interface{}{
						"assignments": []interface{}{
							map[string]interface{}{
								"name":  "message",
								"value": "Hello, World!",
								"type":  "string",
							},
						},
					},
				},
			},
			{
				Name:       "End",
				Type:       "n8n-nodes-base.noOp",
				Parameters: map[string]interface{}{},
			},
		},
		Connections: model.Connections{
			"Start": {
				"main": [][]model.ConnectionTarget{{{Node: "Set Data", Type: "main", Index: 0}}},
			},
			"Set Data": {
				"main": [][]model.ConnectionTarget{{{Node: "End", Type: "main", Index: 0}}},
			},
		},
	}

	engine := NewEngine(DefaultEngineOptions())
	engine.SetRegistry(nodes.NewRegistry())
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := engine.Execute(ctx, workflow, nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Status != StatusSuccess {
		t.Errorf("Expected status %s, got %s", StatusSuccess, result.Status)
	}

	if len(result.NodeResults) != 3 {
		t.Errorf("Expected 3 node results, got %d", len(result.NodeResults))
	}
}

func TestExecuteWithInputData(t *testing.T) {
	workflow := &model.Workflow{
		Name: "Input Workflow",
		Nodes: []model.Node{
			{
				Name:       "Start",
				Type:       "n8n-nodes-base.manualTrigger",
				Parameters: map[string]interface{}{},
			},
			{
				Name:       "Pass",
				Type:       "n8n-nodes-base.noOp",
				Parameters: map[string]interface{}{},
			},
		},
		Connections: model.Connections{
			"Start": {
				"main": [][]model.ConnectionTarget{{{Node: "Pass", Type: "main", Index: 0}}},
			},
		},
	}

	engine := NewEngine(DefaultEngineOptions())
	engine.SetRegistry(nodes.NewRegistry())
	ctx := context.Background()

	inputData := []NodeData{{
		JSON: map[string]interface{}{
			"test": "value",
		},
	}}

	result, err := engine.Execute(ctx, workflow, inputData)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Status != StatusSuccess {
		t.Errorf("Expected status %s, got %s", StatusSuccess, result.Status)
	}
}

func TestExecuteSkipDisabledNodes(t *testing.T) {
	workflow := &model.Workflow{
		Name: "Disabled Node Workflow",
		Nodes: []model.Node{
			{Name: "Start", Type: "n8n-nodes-base.manualTrigger"},
			{Name: "Disabled", Type: "n8n-nodes-base.set", Disabled: true},
			{Name: "End", Type: "n8n-nodes-base.noOp"},
		},
		Connections: model.Connections{
			"Start":    {"main": [][]model.ConnectionTarget{{{Node: "Disabled"}}}},
			"Disabled": {"main": [][]model.ConnectionTarget{{{Node: "End"}}}},
		},
	}

	engine := NewEngine(DefaultEngineOptions())
	engine.SetRegistry(nodes.NewRegistry())
	result, err := engine.Execute(context.Background(), workflow, nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Disabled node should not have a result
	if _, ok := result.NodeResults["Disabled"]; ok {
		t.Error("Disabled node should not have been executed")
	}
}

func TestExecutionTimeout(t *testing.T) {
	workflow := &model.Workflow{
		Name: "Timeout Workflow",
		Nodes: []model.Node{
			{Name: "Start", Type: "n8n-nodes-base.manualTrigger"},
		},
		Connections: model.Connections{},
	}

	engine := NewEngine(DefaultEngineOptions())
	engine.SetRegistry(nodes.NewRegistry())
	
	// Create already-canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := engine.Execute(ctx, workflow, nil)
	if err == nil {
		t.Error("Expected error for canceled context")
	}
	if result.Status != StatusCanceled {
		t.Errorf("Expected status %s, got %s", StatusCanceled, result.Status)
	}
}

func TestExecutionContext(t *testing.T) {
	workflow := &model.Workflow{Name: "Test"}
	ctx := context.Background()
	execCtx := NewExecutionContext(ctx, workflow)

	// Test environment
	execCtx.SetEnvironment(map[string]string{"KEY": "VALUE"})
	env := execCtx.GetEnvironment()
	if env["KEY"] != "VALUE" {
		t.Error("Environment not set correctly")
	}

	// Test variables
	execCtx.SetVariable("test", 123)
	val, ok := execCtx.GetVariable("test")
	if !ok || val != 123 {
		t.Error("Variable not set correctly")
	}

	// Test status
	execCtx.SetStatus(StatusRunning)
	if execCtx.Status() != StatusRunning {
		t.Error("Status not set correctly")
	}

	// Test node output
	execCtx.SetNodeOutput("node1", []NodeData{{JSON: map[string]interface{}{"a": 1}}})
	output, ok := execCtx.GetNodeOutput("node1")
	if !ok || len(output) != 1 {
		t.Error("Node output not set correctly")
	}

	// Test webhook response
	execCtx.SetWebhookResponse(map[string]string{"status": "ok"}, 200)
	resp, code, set := execCtx.GetWebhookResponse()
	if !set || code != 200 || resp == nil {
		t.Error("Webhook response not set correctly")
	}
}
