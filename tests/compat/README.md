# n8n Compatibility Harness (Draft)

This directory contains a draft compatibility harness for `workflow-core`.

## What is here

- `cases/`: deterministic contract cases (first 5 seed cases)
- `harness/`: loader, runners, normalization, and assertions
- `workflow_core_contract_test.go`: local contract checks for `workflow-core`
- `n8n_differential_test.go`: optional differential checks against n8n

## Run

Run local contract tests:

```bash
go test ./tests/compat -run TestWorkflowCoreContractCases -v
```

Run optional differential tests (skips unless configured):

```bash
N8N_REFERENCE_CMD='my-runner --workflow {workflow} --input {input}' go test ./tests/compat -run TestDifferentialParityAgainstN8N -v
```

`N8N_REFERENCE_CMD` must print a JSON array of final output items to stdout.

## Placeholders in `N8N_REFERENCE_CMD`

- `{workflow}`: temporary workflow JSON path
- `{input}`: temporary input JSON path

Example (illustrative):

```bash
N8N_REFERENCE_CMD='n8n execute --file {workflow} --input {input} --raw-output'
```

Docker runner (included in this repository):

```bash
N8N_REFERENCE_CMD='./scripts/run-n8n-reference-docker.sh {workflow} {input}' \
go test ./tests/compat -run TestDifferentialParityAgainstN8N -v
```

Notes on the Docker runner:
- It imports the workflow into an isolated temporary n8n user folder.
- It executes by workflow id via `n8n execute --rawOutput`.
- It extracts the final node `json` items from execution data and prints only a JSON array.
- For workflows starting with `manualTrigger`, it injects the first input item as seed data through a temporary Code node (`__compat_input__`).

Optional image override:

```bash
N8N_DOCKER_IMAGE='n8nio/n8n:latest'
```

## Notes

- Differential tests are intentionally opt-in to keep CI deterministic.
- Contract tests are deterministic and safe for PR gating.
