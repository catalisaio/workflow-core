# CLAUDE.md

This file documents repository-specific guidance for contributors and coding agents.

## Compatibility Testing Policy

`workflow-core` targets behavioral compatibility with n8n workflows. To keep this reliable over time, we use two distinct test layers:

- Contract tests (`TestWorkflowCoreContractCases`)
  - Validate expected behavior of `workflow-core` itself.
  - Fast and deterministic.
  - Required on pull requests.

- Differential tests (`TestDifferentialParityAgainstN8N`)
  - Run the same case in `workflow-core` and n8n reference runtime.
  - Catch semantic drift that can pass local unit/contract tests.
  - Run nightly and on-demand.

## Why Differential Failures Matter

Differential failures are not optional noise. They are an early warning that we are diverging from n8n semantics.

- Unit/contract tests can stay green while compatibility regresses.
- Workflow imports depend on n8n behavior expectations.
- Drift caught in CI/nightly is cheaper than debugging production workflow mismatches.
- Differential history is a leading indicator of compatibility quality over time.

## Standard Commands

- Contract suite:

```bash
go test ./tests/compat -run TestWorkflowCoreContractCases -v
```

- Differential suite (Docker n8n reference):

```bash
N8N_REFERENCE_CMD='./scripts/run-n8n-reference-docker.sh {workflow} {input}' \
go test ./tests/compat -run TestDifferentialParityAgainstN8N -v
```

## CI Expectations

- PR and `master` pushes run the contract suite.
- Nightly workflow runs the differential suite.
- When differential parity fails, treat it as a compatibility incident to investigate, not a flaky test to ignore.
