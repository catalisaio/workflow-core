# Example Workflows

This document provides example workflows demonstrating various features of workflow-core.

## Simple Data Processing

A basic workflow that sets data and passes it through.

```json
{
  "name": "Simple Data Processing",
  "nodes": [
    {
      "name": "Start",
      "type": "n8n-nodes-base.manualTrigger",
      "position": [100, 200],
      "parameters": {}
    },
    {
      "name": "Set Data",
      "type": "n8n-nodes-base.set",
      "position": [300, 200],
      "parameters": {
        "mode": "manual",
        "assignments": {
          "assignments": [
            {"name": "name", "value": "John Doe", "type": "string"},
            {"name": "email", "value": "john@example.com", "type": "string"},
            {"name": "age", "value": 30, "type": "number"}
          ]
        }
      }
    },
    {
      "name": "Output",
      "type": "n8n-nodes-base.noOp",
      "position": [500, 200],
      "parameters": {}
    }
  ],
  "connections": {
    "Start": {
      "main": [[{"node": "Set Data", "type": "main", "index": 0}]]
    },
    "Set Data": {
      "main": [[{"node": "Output", "type": "main", "index": 0}]]
    }
  }
}
```

## Conditional Logic

A workflow demonstrating IF node for conditional branching.

```json
{
  "name": "Conditional Processing",
  "nodes": [
    {
      "name": "Start",
      "type": "n8n-nodes-base.manualTrigger",
      "position": [100, 200],
      "parameters": {}
    },
    {
      "name": "Set Score",
      "type": "n8n-nodes-base.set",
      "position": [300, 200],
      "parameters": {
        "mode": "manual",
        "assignments": {
          "assignments": [
            {"name": "studentName", "value": "Alice", "type": "string"},
            {"name": "score", "value": 85, "type": "number"}
          ]
        }
      }
    },
    {
      "name": "Check Pass",
      "type": "n8n-nodes-base.if",
      "position": [500, 200],
      "parameters": {
        "conditions": {
          "conditions": [
            {
              "leftValue": "={{ $json.score }}",
              "operator": "largerEqual",
              "rightValue": 60
            }
          ]
        }
      }
    },
    {
      "name": "Pass Result",
      "type": "n8n-nodes-base.set",
      "position": [700, 150],
      "parameters": {
        "mode": "manual",
        "assignments": {
          "assignments": [
            {"name": "result", "value": "PASS", "type": "string"},
            {"name": "message", "value": "Congratulations!", "type": "string"}
          ]
        }
      }
    },
    {
      "name": "Fail Result",
      "type": "n8n-nodes-base.set",
      "position": [700, 300],
      "parameters": {
        "mode": "manual",
        "assignments": {
          "assignments": [
            {"name": "result", "value": "FAIL", "type": "string"},
            {"name": "message", "value": "Please try again.", "type": "string"}
          ]
        }
      }
    }
  ],
  "connections": {
    "Start": {
      "main": [[{"node": "Set Score", "type": "main", "index": 0}]]
    },
    "Set Score": {
      "main": [[{"node": "Check Pass", "type": "main", "index": 0}]]
    },
    "Check Pass": {
      "main": [
        [{"node": "Pass Result", "type": "main", "index": 0}],
        [{"node": "Fail Result", "type": "main", "index": 0}]
      ]
    }
  }
}
```

## HTTP API Request

A workflow that fetches data from an external API.

```json
{
  "name": "API Request",
  "nodes": [
    {
      "name": "Start",
      "type": "n8n-nodes-base.manualTrigger",
      "position": [100, 200],
      "parameters": {}
    },
    {
      "name": "Fetch Post",
      "type": "n8n-nodes-base.httpRequest",
      "position": [300, 200],
      "parameters": {
        "method": "GET",
        "url": "https://jsonplaceholder.typicode.com/posts/1"
      }
    },
    {
      "name": "Extract Data",
      "type": "n8n-nodes-base.set",
      "position": [500, 200],
      "parameters": {
        "mode": "manual",
        "keepOnlySet": true,
        "assignments": {
          "assignments": [
            {"name": "title", "value": "={{ $json.data.title }}", "type": "string"},
            {"name": "body", "value": "={{ $json.data.body }}", "type": "string"},
            {"name": "fetchedAt", "value": "={{ $now }}", "type": "string"}
          ]
        }
      }
    }
  ],
  "connections": {
    "Start": {
      "main": [[{"node": "Fetch Post", "type": "main", "index": 0}]]
    },
    "Fetch Post": {
      "main": [[{"node": "Extract Data", "type": "main", "index": 0}]]
    }
  }
}
```

## Using Environment Variables

A workflow that uses environment variables for configuration.

```json
{
  "name": "Environment Config",
  "nodes": [
    {
      "name": "Start",
      "type": "n8n-nodes-base.manualTrigger",
      "position": [100, 200],
      "parameters": {}
    },
    {
      "name": "Build Config",
      "type": "n8n-nodes-base.set",
      "position": [300, 200],
      "parameters": {
        "mode": "manual",
        "assignments": {
          "assignments": [
            {"name": "apiKey", "value": "={{ $env.API_KEY }}", "type": "string"},
            {"name": "environment", "value": "={{ $env.ENV }}", "type": "string"},
            {"name": "debugMode", "value": "={{ $env.DEBUG }}", "type": "string"}
          ]
        }
      }
    }
  ],
  "connections": {
    "Start": {
      "main": [[{"node": "Build Config", "type": "main", "index": 0}]]
    }
  }
}
```

Run with environment variables:

```bash
API_KEY=secret123 ENV=production DEBUG=false ./workflow-core run env-workflow.json
```

## Data Transformation Pipeline

A workflow demonstrating data transformation through multiple steps.

```json
{
  "name": "Data Pipeline",
  "nodes": [
    {
      "name": "Start",
      "type": "n8n-nodes-base.manualTrigger",
      "position": [100, 200],
      "parameters": {}
    },
    {
      "name": "Raw Data",
      "type": "n8n-nodes-base.set",
      "position": [300, 200],
      "parameters": {
        "mode": "manual",
        "assignments": {
          "assignments": [
            {"name": "firstName", "value": "John", "type": "string"},
            {"name": "lastName", "value": "Doe", "type": "string"},
            {"name": "birthYear", "value": 1990, "type": "number"}
          ]
        }
      }
    },
    {
      "name": "Calculate Age",
      "type": "n8n-nodes-base.set",
      "position": [500, 200],
      "parameters": {
        "mode": "manual",
        "assignments": {
          "assignments": [
            {"name": "fullName", "value": "{{ $json.firstName }} {{ $json.lastName }}", "type": "string"},
            {"name": "currentYear", "value": 2024, "type": "number"}
          ]
        }
      }
    },
    {
      "name": "Final Format",
      "type": "n8n-nodes-base.set",
      "position": [700, 200],
      "parameters": {
        "mode": "manual",
        "keepOnlySet": true,
        "assignments": {
          "assignments": [
            {"name": "name", "value": "={{ $json.fullName }}", "type": "string"},
            {"name": "processed", "value": true, "type": "boolean"},
            {"name": "timestamp", "value": "={{ $now }}", "type": "string"}
          ]
        }
      }
    }
  ],
  "connections": {
    "Start": {
      "main": [[{"node": "Raw Data", "type": "main", "index": 0}]]
    },
    "Raw Data": {
      "main": [[{"node": "Calculate Age", "type": "main", "index": 0}]]
    },
    "Calculate Age": {
      "main": [[{"node": "Final Format", "type": "main", "index": 0}]]
    }
  }
}
```

## Running with Input Data

Create an input file `input.json`:

```json
[
  {"id": 1, "name": "Alice", "active": true},
  {"id": 2, "name": "Bob", "active": false},
  {"id": 3, "name": "Charlie", "active": true}
]
```

Workflow that processes input data:

```json
{
  "name": "Process Input",
  "nodes": [
    {
      "name": "Start",
      "type": "n8n-nodes-base.manualTrigger",
      "position": [100, 200],
      "parameters": {}
    },
    {
      "name": "Check Active",
      "type": "n8n-nodes-base.if",
      "position": [300, 200],
      "parameters": {
        "conditions": {
          "conditions": [
            {
              "leftValue": "={{ $json.active }}",
              "operator": "isTrue",
              "rightValue": true
            }
          ]
        }
      }
    },
    {
      "name": "Active Users",
      "type": "n8n-nodes-base.noOp",
      "position": [500, 150],
      "parameters": {}
    }
  ],
  "connections": {
    "Start": {
      "main": [[{"node": "Check Active", "type": "main", "index": 0}]]
    },
    "Check Active": {
      "main": [
        [{"node": "Active Users", "type": "main", "index": 0}]
      ]
    }
  }
}
```

Run:

```bash
./workflow-core run -input input.json -verbose workflow.json
```

## Tips for Building Workflows

1. **Start Simple**: Begin with a manual trigger and a set node
2. **Test Incrementally**: Add nodes one at a time and test
3. **Use Debug Nodes**: Add debug nodes to inspect data flow
4. **Expression Syntax**: Remember the `={{ }}` prefix for expressions
5. **Keep Data Clean**: Use `keepOnlySet: true` to reduce data payload
