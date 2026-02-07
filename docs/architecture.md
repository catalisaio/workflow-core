# Architecture

This document describes the architecture of workflow-core.

## Overview

Workflow-core is designed as a modular, extensible workflow execution engine. The architecture follows clean separation of concerns with distinct layers for parsing, execution, and node implementation.

## Components

### 1. Parser (`pkg/parser`)

The parser is responsible for:
- Reading workflow JSON files
- Deserializing into Go structs
- Validating workflow structure
- Normalizing node configurations

```go
parser := parser.NewParser()
workflow, err := parser.ParseFile("workflow.json")
```

### 2. Model (`pkg/model`)

The model package defines the data structures that represent workflows:

- **Workflow**: The top-level container with nodes, connections, and settings
- **Node**: Individual workflow steps with type, parameters, and position
- **Connections**: Graph edges defining data flow between nodes

### 3. Expression Evaluator (`pkg/expression`)

Handles n8n-style expression evaluation:

- `{{ $json.field }}` - Access current input data
- `{{ $env.VAR }}` - Environment variables
- `{{ $node["Name"].json }}` - Reference other node outputs
- `{{ $now }}` / `{{ $timestamp }}` - Time functions

The evaluator also handles condition evaluation for IF and Switch nodes.

### 4. Engine (`pkg/engine`)

The execution engine orchestrates workflow execution:

1. **Build Execution Order**: Uses topological sort (Kahn's algorithm) to determine node execution order
2. **Execute Nodes**: Runs each node in order, passing data through connections
3. **Handle Branching**: Routes data through conditional paths (IF, Switch)
4. **Track Results**: Maintains execution status and results for each node

```go
engine := engine.NewEngine(engine.DefaultEngineOptions())
result, err := engine.Execute(ctx, workflow, inputData)
```

### 5. Execution Context (`pkg/engine/context.go`)

The execution context maintains state during workflow execution:

- Input/output data for each node
- Environment variables
- Workflow variables
- Execution status
- Webhook response data

### 6. Node Registry (`pkg/nodes`)

The registry manages node type implementations:

- Registers default node types at startup
- Allows custom node registration
- Provides lookup by node type string

### 7. Node Implementations (`pkg/nodes/*.go`)

Each node type implements the `NodeExecutor` interface:

```go
type NodeExecutor interface {
    Execute(ctx *ExecutionContext, params map[string]interface{}, input []NodeData) ([]NodeData, error)
    Type() string
}
```

## Execution Flow

```
┌─────────────┐
│ Parse JSON  │
└─────┬───────┘
      │
      ▼
┌─────────────────────┐
│ Validate & Normalize│
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│ Build Execution DAG │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│ Topological Sort    │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────────────────┐
│ For each node in order:         │
│  1. Get input from predecessors │
│  2. Evaluate expressions        │
│  3. Execute node                │
│  4. Store output                │
│  5. Update context              │
└─────────┬───────────────────────┘
          │
          ▼
┌─────────────────────┐
│ Return Results      │
└─────────────────────┘
```

## DAG Construction

The workflow is represented as a Directed Acyclic Graph (DAG):

- **Nodes**: Workflow nodes become graph vertices
- **Connections**: Workflow connections become directed edges
- **Entry Points**: Trigger nodes have no incoming edges

### Topological Sort

The engine uses Kahn's algorithm for topological sorting:

1. Calculate in-degree for all nodes
2. Add nodes with in-degree 0 to queue
3. Process queue: execute node, reduce in-degree of successors
4. Repeat until queue is empty

This ensures nodes execute only after their dependencies complete.

## Data Flow

Data flows through the workflow as `NodeData` objects:

```go
type NodeData struct {
    JSON   map[string]interface{}  // Main data payload
    Binary map[string]interface{}  // Binary data (files, etc.)
    Error  *NodeError              // Error information
}
```

Each node:
1. Receives `[]NodeData` from connected predecessors
2. Processes the data according to its type and parameters
3. Returns `[]NodeData` for connected successors

## Error Handling

Errors are handled at multiple levels:

1. **Parse Errors**: Invalid JSON or missing required fields
2. **Validation Errors**: Invalid connections, missing nodes
3. **Execution Errors**: Node failures during execution

Options for error handling:
- `ContinueOnError`: Continue workflow on node failure
- `ContinueOnFail` (per-node): Node-specific error handling

## Extensibility

### Adding Custom Nodes

1. Implement the `NodeExecutor` interface
2. Register with the node registry

```go
type MyCustomNode struct{}

func (n *MyCustomNode) Type() string {
    return "custom.myNode"
}

func (n *MyCustomNode) Execute(ctx *ExecutionContext, params map[string]interface{}, input []NodeData) ([]NodeData, error) {
    // Implementation
}

// Register
registry.Register(&MyCustomNode{})
```

### Custom Expression Functions

Extend the expression evaluator by adding new variable handlers or functions.

## Thread Safety

The execution context uses `sync.RWMutex` for thread-safe access to shared state. While the current implementation executes nodes sequentially, the architecture supports future parallel execution.

## Performance Considerations

- **Memory**: NodeData is passed by reference where possible
- **Allocation**: Reuse maps and slices when practical
- **HTTP**: Connection pooling for HTTP requests
- **Context**: Proper cancellation support for long-running operations
