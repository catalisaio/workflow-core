# workflow-core

[![Go Version](https://img.shields.io/badge/go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

High-performance workflow execution engine compatible with n8n workflow format. Execute complex automation workflows defined in JSON with support for conditional logic, HTTP requests, data transformations, IAM integration, and more.

## Features

- **Workflow Format Compatibility**: Parse and execute workflows in n8n-compatible JSON format
- **DAG-Based Execution**: Automatic dependency resolution and topological sorting
- **Rich Node Library**: Built-in support for triggers, conditions, HTTP requests, data manipulation
- **Expression Evaluation**: Support for `{{ }}` expression syntax with `$json`, `$env`, and `$node` references
- **IAM Integration**: Authentication and authorization via building-blocks IAM service
- **Multi-Tenant Support**: Organization context for workflow isolation
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
- `manualTrigger` - Manual execution trigger
- `webhook` - HTTP webhook trigger
- `cron` - Cron/schedule trigger
- `scheduleTrigger` - Schedule trigger

### Data Manipulation
- `set` - Set/modify data fields
- `merge` - Merge data from multiple sources
- `splitInBatches` - Split data into batches

### Control Flow
- `if` - Conditional branching
- `switch` - Multi-way branching
- `noOp` - Pass-through node

### HTTP
- `httpRequest` - Make HTTP requests
- `respondToWebhook` - Send webhook response

### Code Execution
- `code` - Execute JavaScript-like code

### Debugging
- `debug` - Log debug information

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
    "github.com/catalisaio/workflow-core/pkg/nodes"
    "github.com/catalisaio/workflow-core/pkg/parser"
)

func main() {
    // Parse workflow
    p := parser.NewParser()
    workflow, err := p.ParseFile("workflow.json")
    if err != nil {
        log.Fatal(err)
    }

    // Create engine with registry
    eng := engine.NewEngine(engine.DefaultEngineOptions())
    eng.SetRegistry(nodes.NewRegistry())

    // Execute
    result, err := eng.Execute(context.Background(), workflow, nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Status: %s\n", result.Status)
    fmt.Printf("Duration: %v\n", result.Duration)
}
```

## IAM Integration

workflow-core integrates with the building-blocks IAM service for authentication and authorization:

```go
import "github.com/catalisaio/workflow-core/pkg/auth"

// Create auth middleware
middleware, _ := auth.NewAuthMiddleware(auth.AuthMiddlewareConfig{
    JWTSecret: os.Getenv("JWT_SECRET"),
})

// Protect API routes
http.Handle("/api/", middleware.Handler(apiHandler))

// Require specific permission
executeHandler := middleware.RequirePermission(auth.PermissionWorkflowsExecute)(handler)
```

See [IAM Integration](docs/iam-integration.md) for detailed documentation.

### Environment Variables

| Variable | Description |
|----------|-------------|
| `IAM_BASE_URL` | IAM service URL |
| `JWT_SECRET` | JWT validation secret |
| `AUTH_ENABLED` | Enable authentication (default: true) |

## Expression Syntax

workflow-core supports expressions in node parameters:

```
{{ $json.fieldName }}       - Access input data
{{ $json["field-name"] }}   - Bracket notation
{{ $env.API_KEY }}          - Environment variables
{{ $now }}                  - Current timestamp
{{ $node["NodeName"].json }} - Reference other node output
```

## Compatibility Testing

`workflow-core` uses two complementary compatibility layers:

- Contract tests: validate expected `workflow-core` behavior with deterministic fixtures.
- Differential tests: compare `workflow-core` output against n8n reference execution for the same cases.

Why this is important:

- Contract/unit tests can remain green while n8n compatibility drifts.
- Differential failures detect semantic mismatches early (branching, expressions, node parameter behavior).
- Catching drift in CI/nightly avoids production workflow surprises.

Run contract suite:

```bash
go test ./tests/compat -run TestWorkflowCoreContractCases -v
```

Run differential suite (Docker n8n):

```bash
N8N_REFERENCE_CMD='./scripts/run-n8n-reference-docker.sh {workflow} {input}' \
go test ./tests/compat -run TestDifferentialParityAgainstN8N -v
```

CI setup in this repository:

- `.github/workflows/compat-contract.yml` runs on pull requests and `master` pushes.
- `.github/workflows/compat-differential-nightly.yml` runs nightly and on manual dispatch.

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
│              Auth Middleware (IAM)                  │
├─────────────────────────────────────────────────────┤
│                  Node Registry                      │
│  ┌────────┬────────┬────────┬────────┬─────────┐   │
│  │Triggers│  Data  │Control │  HTTP  │  Code   │   │
│  └────────┴────────┴────────┴────────┴─────────┘   │
└─────────────────────────────────────────────────────┘
```

## Roadmap

- [x] Core workflow execution engine
- [x] IAM integration
- [ ] REST API server mode
- [ ] Additional node types (Email, Database, Queue)
- [ ] Parallel execution support
- [ ] Retry policies
- [ ] Webhook server mode
- [ ] Sub-workflow execution
- [ ] Credential management
- [ ] OpenTelemetry tracing

## Documentation

- [Getting Started](docs/getting-started.md)
- [Architecture](docs/architecture.md)
- [Supported Nodes](docs/supported-nodes.md)
- [Example Workflows](docs/examples.md)
- [IAM Integration](docs/iam-integration.md)

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details.

## License

MIT License - see [LICENSE](LICENSE) for details.
