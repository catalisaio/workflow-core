package nodes

import (
	"testing"
)

func TestRegistryRegisterAndGet(t *testing.T) {
	r := NewRegistry()

	// Check that defaults are registered
	if !r.Has("n8n-nodes-base.manualTrigger") {
		t.Error("ManualTrigger should be registered by default")
	}

	if !r.Has("n8n-nodes-base.set") {
		t.Error("Set node should be registered by default")
	}

	if !r.Has("n8n-nodes-base.if") {
		t.Error("IF node should be registered by default")
	}
}

func TestRegistryList(t *testing.T) {
	r := NewRegistry()
	
	types := r.List()
	if len(types) == 0 {
		t.Error("Registry should have registered nodes")
	}

	// Check for some expected types
	hasManual := false
	hasSet := false
	hasHTTP := false
	
	for _, nodeType := range types {
		switch nodeType {
		case "n8n-nodes-base.manualTrigger":
			hasManual = true
		case "n8n-nodes-base.set":
			hasSet = true
		case "n8n-nodes-base.httpRequest":
			hasHTTP = true
		}
	}

	if !hasManual {
		t.Error("ManualTrigger not in list")
	}
	if !hasSet {
		t.Error("Set not in list")
	}
	if !hasHTTP {
		t.Error("HTTPRequest not in list")
	}
}

func TestRegistryUnregister(t *testing.T) {
	r := NewRegistry()
	
	nodeType := "n8n-nodes-base.noOp"
	if !r.Has(nodeType) {
		t.Fatal("NoOp should be registered")
	}

	r.Unregister(nodeType)
	
	if r.Has(nodeType) {
		t.Error("NoOp should be unregistered")
	}
}

func TestRegistryCustomNode(t *testing.T) {
	r := NewRegistry()
	
	// Create a custom node
	custom := &NoOpNode{}
	
	// The NoOp is already registered, but we can get it
	executor, ok := r.Get("n8n-nodes-base.noOp")
	if !ok {
		t.Fatal("NoOp should be retrievable")
	}

	if executor.Type() != custom.Type() {
		t.Error("Node types should match")
	}
}

func TestRegistryGetNonExistent(t *testing.T) {
	r := NewRegistry()
	
	_, ok := r.Get("non.existent.node")
	if ok {
		t.Error("Non-existent node should not be found")
	}
}

func TestRegistryGetSupportedNodes(t *testing.T) {
	r := NewRegistry()
	
	nodes := r.GetSupportedNodes()
	if len(nodes) == 0 {
		t.Error("Should have supported nodes")
	}
}
