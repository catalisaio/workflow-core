# workflow-core

[![Go Version](https://img.shields.io/badge/go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

High-performance workflow execution engine with n8n workflow format compatibility. Execute complex automation workflows defined in JSON with support for conditional logic, HTTP requests, data transformations, and more.

## Features

- **n8n Format Compatible**: Parse and execute workflows in the n8n JSON format
- **DAG-Based Execution**: Automatic dependency resolution and topological sorting
- **Rich Node Library**: Built-in support for triggers, conditions, HTTP requests, data manipulation
- **Expression Evaluation**: Support for `{{ }}` expression syntax with `$json`, `$env`, and `$node` references
- **Extensible Architecture**: Easy to add custom node types
- **CLI & Library**: Use as a command-line tool or embed in your Go applications

## Installation

### From Source

```bash
git clone https://github.com/catalisaio/workflow-core.git
cd workflow-core
go build -o workflow-core ./cmd/workflow-core/
```

### Using Go Install

```bash
go install github.com/catalisaio/workflow-core/cmd/workflow-core@latest
```

## Quick Start

### Running a Workflow

```bash
# Execute a workflow
./workflow-core run examples/simple-workflow.json

# Execute with verbose output
./workflow-core run -verbose examples/simple-workflow.json

# Execute with input data
./workflow-core run -input data.json workflow.json
```

### Validating a Workflow

```bash
./workflow-core validate workflow.json
```

### Listing Supported Nodes

```bash
./workflow-core list-nodes
```

## Supported Node Types

### Triggers
- `n8n-nodes-base.manualTrigger` - Manual execution trigger
- `n8n-nodes-base.webhook` - HTTP webhook trigger
- `n8n-nodes-base.cron` - Cron/schedule trigger
- `n8n-nodes-base.scheduleTrigger` - Schedule trigger

### Data Manipulation
- `n8n-nodes-base.set` - Set/modify data fields
- `n8n-nodes-base.merge` - Merge data from multiple sources
- `n8n-nodes-base.splitInBatches` - Split data into batches

### Control Flow
- `n8n-nodes-base.if` - Conditional branching
- `n8n-nodes-base.switch` - Multi-way branching
- `n8n-nodes-base.noOp` - Pass-through node

### HTTP
- `n8n-nodes-base.httpRequest` - Make HTTP requests
- `n8n-nodes-base.respondToWebhook` - Send webhook response

### Code Execution
- `n8n-nodes-base.code` - Execute JavaScript-like code

### Debugging
- `n8n-nodes-base.debug` - Log debug information

## Example Workflow

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
      "name": "Set Message",
      "type": "n8n-nodes-base.set",
      "position": [300, 200],
      "parameters": {
        "mode": "manual",
        "assignments": {
          "assignments": [
            {"name": "message", "value": "Hello, World!", "type": "string"}
          ]
        }
      }
    }
  ],
  "connections": {
    "Start": {
      "main": [[{"node": "Set Message", "type": "main", "index": 0}]]
    }
  }
}
```

## Using as a Library

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
    // Parse workflow
    p := parser.NewParser()
    workflow, err := p.ParseFile("workflow.json")
    if err != nil {
        log.Fatal(err)
    }

    // Create engine
    eng := engine.NewEngine(engine.DefaultEngineOptions())

    // Execute
    result, err := eng.Execute(context.Background(), workflow, nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Status: %s\n", result.Status)
    fmt.Printf("Duration: %v\n", result.Duration)
}
```

## Expression Syntax

Workflow-core supports n8n-style expressions:

```
{{ $json.fieldName }}       - Access input data
{{ $json["field-name"] }}   - Bracket notation
{{ $env.API_KEY }}          - Environment variables
{{ $now }}                  - Current timestamp
{{ $node["NodeName"].json }} - Reference other node output
```

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                    CLI / API                        │
├─────────────────────────────────────────────────────┤
│                     Engine                          │
│  ┌─────────────┬──────────────┬─────────────────┐  │
│  │   Parser    │  Expression  │    Executor     │  │
│  │             │  Evaluator   │                 │  │
│  └─────────────┴──────────────┴─────────────────┘  │
├─────────────────────────────────────────────────────┤
│                  Node Registry                      │
│  ┌────────┬────────┬────────┬────────┬─────────┐   │
│  │Triggers│  Data  │Control │  HTTP  │  Code   │   │
│  └────────┴────────┴────────┴────────┴─────────┘   │
└─────────────────────────────────────────────────────┘
```

## Roadmap

- [ ] Additional node types (Email, Database, Queue)
- [ ] Parallel execution support
- [ ] Retry policies
- [ ] Webhook server mode
- [ ] Workflow variables and static data
- [ ] Sub-workflow execution
- [ ] Credential management
- [ ] OpenTelemetry tracing

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details.

## License

MIT License - see [LICENSE](LICENSE) for details.
