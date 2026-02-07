## n8n Compatibility Draft

This draft defines how to measure and improve `workflow-core` compatibility with n8n behavior.

### Goals

- Make compatibility measurable with repeatable tests.
- Detect behavior drift early when `workflow-core` changes.
- Prioritize parity for the node subset already implemented in `workflow-core`.

### Scope

In scope (current engine):
- Workflow JSON parsing and structural validation.
- Execution semantics for control-flow and data-flow.
- Expression evaluation behavior (`$json`, `$env`, `$node`, operators).
- Node parity for:
  - `manualTrigger`, `webhook`, `cron`, `scheduleTrigger`
  - `set`, `merge`, `splitInBatches`
  - `if`, `switch`, `noOp`
  - `httpRequest`, `respondToWebhook`
  - `code`, `debug`

Out of scope (for now):
- Full n8n product features (UI/editor behavior, credential manager, source control, enterprise features).
- Compatibility with all 400+ app nodes.
- Performance benchmark equivalence.

### Sources of Truth

Primary upstream references:
- n8n execution tests: `packages/core/src/execution-engine/__tests__/workflow-execute.test.ts`
- n8n run-node behavior tests: `packages/core/src/execution-engine/__tests__/workflow-execute-run-node.test.ts`
- n8n expression and proxy tests: `packages/workflow/test/expression.test.ts`, `packages/workflow/test/workflow-data-proxy.test.ts`
- node-specific tests/fixtures:
  - `packages/nodes-base/nodes/If/test/v2/IfV2.node.test.ts`
  - `packages/nodes-base/nodes/HttpRequest/test/node/HttpRequestV3.test.ts`
  - `packages/nodes-base/nodes/Code/test/Code.node.test.ts`
  - `packages/nodes-base/nodes/Set/test/Set.v3.workflow.json`
  - `packages/nodes-base/nodes/Switch/V3/test/switch.rules.workflow.json`
- n8n CLI workflow corpus runner:
  - `packages/testing/playwright/tests/cli-workflows/workflow-tests.spec.ts`

Note: n8n does not publish a stable external "compatibility certification suite". Upstream tests are de-facto references and may evolve.

### Compatibility Strategy

Use differential testing:
- Execute the same workflow and input in:
  1) `workflow-core`
  2) n8n reference runtime
- Normalize outputs.
- Compare semantic results.

#### Result Normalization Rules

Normalize before comparison to avoid false negatives:
- Remove non-deterministic fields (`executionIndex`, timestamps, ids, internal metadata).
- Sort object keys recursively.
- Compare arrays preserving order only when order is semantically required.
- Coerce numeric formatting differences where equivalent (`1` vs `1.0`).
- Support per-test ignore paths for known non-contract fields.

### Test Matrix (Draft)

#### Layer 1: Parser and Graph

- Parse valid n8n workflow JSON variants (typeVersion differences).
- Reject invalid references in strict mode.
- Validate DAG ordering and cycle handling.
- Validate branch connection index semantics.

#### Layer 2: Expressions

- `$json` dot and bracket access.
- `$env` values and missing keys.
- `$node["X"].json` references.
- Condition operators (`equals`, `contains`, `larger`, `isEmpty`, etc.).

#### Layer 3: Execution Semantics

- Branch routing correctness for `if` and `switch` outputs.
- Merge behavior for multiple inputs and branch combinations.
- Disabled node behavior.
- Continue-on-fail behavior.

#### Layer 4: Node Parity

- `set`: manual/raw modes, assignment expressions, keepOnlySet.
- `httpRequest`: method, headers, query, body mode, response parsing.
- `code`: supported patterns and failure behavior.
- `respondToWebhook`: payload and status behavior.

### Proposed Harness Design

Create a new test package (suggested):
- `tests/compat/`

Core components:
- `tests/compat/cases/*.json`
  - Contains workflow, input, expected profile, and compare options.
- `tests/compat/harness/reference_n8n.ts` (or Go wrapper + subprocess)
  - Runs reference execution in n8n (containerized or local).
- `tests/compat/harness/workflow_core.go`
  - Runs `workflow-core` execution.
- `tests/compat/harness/normalize.go`
  - Canonicalization and ignore-path handling.
- `tests/compat/harness/assert.go`
  - Semantic diff and readable failure output.

### Execution Modes

Support two modes:
- `contract` mode: deterministic offline fixtures only (CI-required).
- `extended` mode: broader corpus, optional network/integration tests.

### Initial Compatibility Gate (MVP)

Pass criteria for first gate:
- 20 deterministic parity cases passing.
- Mandatory coverage:
  - `if` true/false branch routing
  - `switch` multi-rule output routing
  - `set` expression assignment on multi-item inputs
  - `httpRequest` JSON + text response handling
  - `$node` and `$json` expression references

Recommended CI behavior:
- Run `contract` suite on PR.
- Block merge on parity regressions.
- Publish diff artifact on failure.

### Seed Case List (First 20)

1. `if`: true branch only
2. `if`: false branch only
3. `if`: mixed items split across branches
4. `switch`: first-match rule
5. `switch`: fallback output
6. `switch`: expression mode output selection
7. `set`: v3 assignments simple values
8. `set`: expression in assignment (`$json`)
9. `set`: `keepOnlySet=true`
10. `merge`: append mode
11. `merge`: combine by position
12. `splitInBatches`: fixed size with loop metadata
13. `httpRequest`: GET + query params
14. `httpRequest`: POST JSON body
15. `httpRequest`: text response parsing
16. `code`: return items passthrough
17. `code`: run once per item assignment pattern
18. `expression`: `$env` and missing key behavior
19. `expression`: `$node` previous output reference
20. `disabled node` + downstream behavior

### Risks and Mitigations

- Upstream drift in n8n tests
  - Pin n8n commit/hash in harness metadata.
- False mismatches from non-contract fields
  - Centralize canonicalization and ignore paths.
- Flaky integration cases
  - Keep PR gate deterministic; run flaky/integration in nightly.

### Roadmap

Phase 1 (MVP):
- Implement harness skeleton and 20 deterministic cases.
- Add CI gate.

Phase 2:
- Expand to 50+ cases from n8n fixtures/workflows.
- Add property-based tests for expression/operator edge cases.

Phase 3:
- Track compatibility score per node family over time.
- Add automated sync script for selected upstream fixtures.

### Open Decisions

- Reference runner strategy:
  - A) Embed n8n execution path directly (higher fidelity, higher complexity)
  - B) Execute n8n CLI in container and parse output (recommended for MVP)
- Fixture ownership:
  - A) Mirror selected upstream fixtures
  - B) Maintain curated local contract fixtures (recommended for stability)

### Implementation Backlog

- See `docs/n8n-compatibility-issue-backlog.md` for copy-paste issue drafts and checklist.
