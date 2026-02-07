## n8n Compatibility Backlog (Draft)

Use this as a copy-paste issue backlog and implementation checklist.

### Checklist

- [ ] COMPAT-01: Add compatibility harness foundation
- [ ] COMPAT-02: Add deterministic contract suite (20 seed cases)
- [ ] COMPAT-03: Add differential parity runner against n8n
- [ ] COMPAT-04: Fix IF/Switch branch output routing parity
- [ ] COMPAT-05: Fix Set per-item expression evaluation parity
- [ ] COMPAT-06: Harden HTTPRequest resource lifecycle and parity
- [ ] COMPAT-07: Add compatibility CI gate and failure artifacts
- [ ] COMPAT-08: Add upstream fixture sync workflow

---

### COMPAT-01: Add compatibility harness foundation

**Title**
`compat: add harness foundation for n8n parity tests`

**Description**
- Create `tests/compat` with case loader, output normalization, and assertion helpers.
- Add local `workflow-core` runner and pluggable reference runner interface.
- Document usage in `tests/compat/README.md`.

**Acceptance Criteria**
- Harness compiles and runs in `go test`.
- At least one sample case executes end-to-end.

---

### COMPAT-02: Add deterministic contract suite (20 seed cases)

**Title**
`compat: add deterministic contract cases for core node subset`

**Description**
- Add 20 deterministic JSON fixtures for implemented node families.
- Cover control flow, expressions, Set, HTTPRequest, Merge, SplitInBatches.
- Ensure no external API dependency in PR-gated cases.

**Acceptance Criteria**
- `go test ./tests/compat -run TestWorkflowCoreContractCases` passes.
- Cases are stable and deterministic.

---

### COMPAT-03: Add differential parity runner against n8n

**Title**
`compat: add optional n8n differential runner`

**Description**
- Support `N8N_REFERENCE_CMD` with `{workflow}` and `{input}` placeholders.
- Run each case in both runtimes and compare normalized outputs.
- Keep suite opt-in by default.

**Acceptance Criteria**
- Differential test skips cleanly when not configured.
- Differential failures include readable expected/actual diff.

---

### COMPAT-04: Fix IF/Switch branch output routing parity

**Title**
`fix(engine): honor output index/type routing for branch nodes`

**Description**
- Track and route node outputs by output index and connection type.
- Prevent branch leakage where all downstream nodes receive shared output.
- Add regression tests for true/false and multi-rule switch outputs.

**Acceptance Criteria**
- IF/Switch branch parity cases match n8n reference behavior.
- No regressions in existing engine tests.

---

### COMPAT-05: Fix Set per-item expression evaluation parity

**Title**
`fix(set): evaluate expressions against each current input item`

**Description**
- Set evaluator context per item before assignment evaluation.
- Add multi-item expression fixture with differing item values.

**Acceptance Criteria**
- Multi-item Set expression case passes.
- No regressions for non-expression assignments.

---

### COMPAT-06: Harden HTTPRequest resource lifecycle and parity

**Title**
`fix(httpRequest): improve per-item cleanup and parity coverage`

**Description**
- Remove deferred cleanup patterns inside loops where needed.
- Expand fixtures for body modes, headers, query, and response parsing.

**Acceptance Criteria**
- HTTPRequest contract cases pass.
- Resource usage remains stable under multi-item inputs.

---

### COMPAT-07: Add compatibility CI gate and failure artifacts

**Title**
`ci: add compatibility contract gate`

**Description**
- Run deterministic contract suite on pull requests.
- Upload structured diff artifacts on failures.
- Keep differential suite in optional/nightly job.

**Acceptance Criteria**
- PRs fail on contract regressions.
- Failure logs clearly show case id and semantic diff.

---

### COMPAT-08: Add upstream fixture sync workflow

**Title**
`tooling: sync selected n8n upstream fixtures`

**Description**
- Add script to fetch/copy selected upstream fixtures by pinned n8n commit.
- Maintain mapping and transformation rules for local harness format.

**Acceptance Criteria**
- Sync script updates local fixture set reproducibly.
- Pinned upstream commit/hash is recorded.
