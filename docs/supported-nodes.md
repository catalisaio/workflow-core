# Supported Nodes

This document lists all node types supported by workflow-core.

## Trigger Nodes

### Manual Trigger

**Type:** `n8n-nodes-base.manualTrigger`

Entry point for manual workflow execution.

**Parameters:** None

**Output:** Passes through input data or empty data if none provided.

---

### Webhook

**Type:** `n8n-nodes-base.webhook`

HTTP webhook trigger for receiving external requests.

**Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| httpMethod | string | GET | HTTP method to accept |
| path | string | "" | Webhook path |
| responseMode | string | onReceived | When to respond |

**Output:** Webhook request data with `$webhook` metadata.

---

### Cron

**Type:** `n8n-nodes-base.cron`

Schedule-based trigger using cron expressions.

**Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| cronExpression | string | Cron expression |

**Output:** Empty data with `$trigger.type: "cron"` metadata.

---

### Schedule Trigger

**Type:** `n8n-nodes-base.scheduleTrigger`

Schedule-based trigger with human-readable intervals.

**Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| rule | object | Schedule configuration |

**Output:** Empty data with `$trigger.type: "schedule"` metadata.

---

## Data Manipulation Nodes

### Set

**Type:** `n8n-nodes-base.set`

Set or modify data fields.

**Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| mode | string | manual | Mode: "manual" or "raw" |
| keepOnlySet | boolean | false | Discard input fields |
| assignments | object | - | Field assignments (v3) |
| values | object | - | Field values (v1/v2) |

**Example (v3):**
```json
{
  "mode": "manual",
  "assignments": {
    "assignments": [
      {"name": "field1", "value": "hello", "type": "string"},
      {"name": "count", "value": 42, "type": "number"}
    ]
  }
}
```

**Example (v1/v2):**
```json
{
  "values": {
    "string": [{"name": "field1", "value": "hello"}],
    "number": [{"name": "count", "value": 42}]
  }
}
```

---

### Merge

**Type:** `n8n-nodes-base.merge`

Combine data from multiple inputs.

**Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| mode | string | append | Merge mode |
| combineMode | string | mergeByPosition | How to combine |

**Modes:**
- `append`: Append all items
- `combine`: Merge items by position or fields
- `chooseBranch`: Select specific input
- `multiplex`: Create all combinations

---

### Split In Batches

**Type:** `n8n-nodes-base.splitInBatches`

Split input items into smaller batches.

**Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| batchSize | number | 10 | Items per batch |
| options.reset | boolean | false | Reset batch counter |

**Output:** Batch of items with `$batch` metadata.

---

## Control Flow Nodes

### IF

**Type:** `n8n-nodes-base.if`

Conditional branching based on conditions.

**Parameters (v2):**
```json
{
  "conditions": {
    "options": {"combineOperation": "and"},
    "conditions": [
      {"leftValue": "={{ $json.count }}", "operator": "larger", "rightValue": 10}
    ]
  }
}
```

**Operators:**
- `equals`, `notEquals`
- `contains`, `notContains`
- `startsWith`, `endsWith`
- `larger`, `smaller`, `largerEqual`, `smallerEqual`
- `isEmpty`, `isNotEmpty`
- `isTrue`, `isFalse`

**Outputs:**
- Output 0: Items matching condition (true)
- Output 1: Items not matching (false)

---

### Switch

**Type:** `n8n-nodes-base.switch`

Multi-way branching based on rules or expressions.

**Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| mode | string | rules | "rules" or "expression" |
| rules | array | - | Rule definitions |
| fallbackOutput | string | none | Fallback behavior |

**Example:**
```json
{
  "mode": "rules",
  "rules": [
    {
      "conditions": [
        {"leftValue": "={{ $json.type }}", "operator": "equals", "rightValue": "A"}
      ]
    }
  ]
}
```

---

### NoOp

**Type:** `n8n-nodes-base.noOp`

Pass-through node that does nothing.

**Parameters:** None

**Output:** Input data unchanged.

---

## HTTP Nodes

### HTTP Request

**Type:** `n8n-nodes-base.httpRequest`

Make HTTP requests to external APIs.

**Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| method | string | GET | HTTP method |
| url | string | - | Request URL |
| bodyContentType | string | json | Body content type |
| body | object/string | - | Request body |
| headerParameters | array | - | HTTP headers |
| queryParameters | array | - | URL query parameters |

**Example:**
```json
{
  "method": "POST",
  "url": "https://api.example.com/data",
  "bodyContentType": "json",
  "body": {"key": "value"},
  "headerParameters": [
    {"name": "Authorization", "value": "Bearer {{ $env.API_KEY }}"}
  ]
}
```

**Output:**
```json
{
  "statusCode": 200,
  "statusText": "200 OK",
  "headers": {"Content-Type": "application/json"},
  "data": {...}
}
```

---

### Respond to Webhook

**Type:** `n8n-nodes-base.respondToWebhook`

Send HTTP response for webhook triggers.

**Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| respondWith | string | firstIncomingItem | Response data source |
| responseCode | number | 200 | HTTP status code |
| responseHeaders | object | - | Response headers |
| responseBody | string | - | Response body (for text/json) |

**respondWith options:**
- `firstIncomingItem`: First input item
- `allIncomingItems`: All input items as array
- `text`: Custom text
- `json`: Custom JSON
- `noData`: Empty response

---

## Code Execution

### Code

**Type:** `n8n-nodes-base.code`

Execute JavaScript-like code.

**Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| mode | string | runOnceForAllItems | Execution mode |
| jsCode | string | - | Code to execute |

**Modes:**
- `runOnceForAllItems`: Execute once with all items
- `runOnceForEachItem`: Execute for each item

**Supported Patterns:**
```javascript
// Return items unchanged
return items;

// Return all input
return $input.all();

// Map transformation
items.map(item => ({ json: {...item.json, newField: "value"} }))
```

**Note:** This is a simplified implementation. Complex JavaScript logic may need to be implemented as custom nodes.

---

## Debugging

### Debug

**Type:** `n8n-nodes-base.debug`

Log data for debugging purposes.

**Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| logOutput | boolean | true | Enable logging |
| logLevel | string | info | Log level |
| message | string | - | Custom message |

**Log levels:** `info`, `warn`, `error`, `debug`

---

## Adding Custom Nodes

To add a custom node:

```go
type MyNode struct{}

func (n *MyNode) Type() string {
    return "custom.myNode"
}

func (n *MyNode) Execute(
    ctx *engine.ExecutionContext,
    params map[string]interface{},
    input []engine.NodeData,
) ([]engine.NodeData, error) {
    // Your implementation
    return input, nil
}

// Register
registry.Register(&MyNode{})
```
