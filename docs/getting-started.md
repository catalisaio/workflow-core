# Getting Started

This guide will help you get started with workflow-core.

## Installation

### Prerequisites

- Go 1.22 or later
- Git

### Building from Source

```bash
# Clone the repository
git clone https://github.com/catalisaio/workflow-core.git
cd workflow-core

# Build the binary
go build -o workflow-core ./cmd/workflow-core/

# Verify installation
./workflow-core version
```

### Using Go Install

```bash
go install github.com/catalisaio/workflow-core/cmd/workflow-core@latest
```

## Your First Workflow

### 1. Create a Workflow File

Create a file called `hello.json`:

```json
{
  "name": "Hello World",
  "nodes": [
    {
      "name": "Start",
      "type": "n8n-nodes-base.manualTrigger",
      "position": [100, 200],
      "parameters": {}
    },
    {
      "name": "Set Greeting",
      "type": "n8n-nodes-base.set",
      "position": [300, 200],
      "parameters": {
        "mode": "manual",
        "assignments": {
          "assignments": [
            {
              "name": "greeting",
              "value": "Hello, World!",
              "type": "string"
            },
            {
              "name": "timestamp",
              "value": "={{ $now }}",
              "type": "string"
            }
          ]
        }
      }
    }
  ],
  "connections": {
    "Start": {
      "main": [[{"node": "Set Greeting", "type": "main", "index": 0}]]
    }
  }
}
```

### 2. Run the Workflow

```bash
./workflow-core run hello.json
```

Expected output:
```
=== Execution Complete ===
Workflow: Hello World
Status: success
Duration: 1.234ms
Nodes executed: 2

--- Final Output ---
Item 0:
{
  "greeting": "Hello, World!",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## CLI Commands

### Run a Workflow

```bash
# Basic execution
./workflow-core run workflow.json

# With verbose output
./workflow-core run -verbose workflow.json

# With input data
./workflow-core run -input data.json workflow.json

# With custom timeout
./workflow-core run -timeout 5m workflow.json
```

### Validate a Workflow

```bash
./workflow-core validate workflow.json
```

### List Supported Nodes

```bash
./workflow-core list-nodes
```

## Using as a Library

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/catalisaio/workflow-core/pkg/engine"
    "github.com/catalisaio/workflow-core/pkg/parser"
)

func main() {
    // Parse workflow from file
    p := parser.NewParser()
    workflow, err := p.ParseFile("workflow.json")
    if err != nil {
        log.Fatalf("Failed to parse: %v", err)
    }

    // Create engine with default options
    eng := engine.NewEngine(engine.DefaultEngineOptions())

    // Execute workflow
    ctx := context.Background()
    result, err := eng.Execute(ctx, workflow, nil)
    if err != nil {
        log.Fatalf("Execution failed: %v", err)
    }

    // Process results
    fmt.Printf("Status: %s\n", result.Status)
    for _, output := range result.FinalOutput {
        fmt.Printf("Output: %v\n", output.JSON)
    }
}
```

### With Input Data

```go
inputData := []engine.NodeData{
    {
        JSON: map[string]interface{}{
            "name": "Alice",
            "age":  30,
        },
    },
}

result, err := eng.Execute(ctx, workflow, inputData)
```

### With Environment Variables

```go
opts := engine.DefaultEngineOptions()
opts.Environment = map[string]string{
    "API_KEY": "your-api-key",
    "DEBUG":   "true",
}

eng := engine.NewEngine(opts)
```

### With Custom Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

result, err := eng.Execute(ctx, workflow, nil)
```

## Working with Expressions

Workflow-core supports n8n-style expressions in node parameters:

### Accessing Input Data

```json
{
  "value": "={{ $json.fieldName }}"
}
```

### Using Environment Variables

```json
{
  "apiKey": "={{ $env.API_KEY }}"
}
```

### Current Timestamp

```json
{
  "createdAt": "={{ $now }}"
}
```

### Referencing Other Nodes

```json
{
  "previousResult": "={{ $node[\"NodeName\"].json.field }}"
}
```

## Workflow Structure

### Nodes

Each node requires:
- `name`: Unique identifier
- `type`: Node type (e.g., `n8n-nodes-base.set`)
- `parameters`: Node-specific configuration

Optional:
- `position`: Visual position [x, y]
- `disabled`: Skip execution if true

### Connections

Connections define the data flow:

```json
{
  "connections": {
    "SourceNode": {
      "main": [
        [{"node": "TargetNode", "type": "main", "index": 0}]
      ]
    }
  }
}
```

For conditional nodes (IF, Switch), multiple outputs are supported:

```json
{
  "connections": {
    "IF Node": {
      "main": [
        [{"node": "TrueOutput", "type": "main", "index": 0}],
        [{"node": "FalseOutput", "type": "main", "index": 0}]
      ]
    }
  }
}
```

## Next Steps

- Explore [Supported Nodes](supported-nodes.md)
- Check out [Example Workflows](examples.md)
- Read the [Architecture](architecture.md) documentation
