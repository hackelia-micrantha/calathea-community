# Structured Invocation Contract Boundary

Status: architecture refinement / companion to `invokrum-instruction-boundary.md`  
Date: 2026-08-12

## Purpose

Calathea's existing Invokrum boundary correctly makes composed instruction bytes attributable. This refinement makes explicit that **rendered instructions are a derived runtime artifact, not the canonical representation of workflow state**.

Calathea owns structured project/workflow semantics. Invokrum owns deterministic composition and, when its portable invocation-state contract is stable, the portable invocation boundary. A runtime owns its private working state.

## Target model

```text
Calathea canonical state
  - operation / goal
  - project state revision
  - selected inputs/evidence
  - selected examples/context
  - role
  - requested capabilities
  - constraints
  - expected outputs
          |
          | compile/materialize
          v
Invokrum invocation boundary
  - exact instruction artifact
  - task-context identity
  - constraints / requested capabilities
  - output/evidence contract
          |
          v
runtime / provider
  - private working state
          |
          v
validated output / evidence
          |
          v
Calathea reviewed state transition
```

The existing exact-byte invariant remains important: if Invokrum returns an instruction artifact with a digest, only those exact bytes are covered by that artifact identity. This note adds a semantic layer above it; exact prompt bytes are not the whole workflow contract.

## Canonical versus derived state

### Canonical Calathea state

Canonical state remains governed by Calathea's existing source-of-truth and lifecycle RFCs. It includes the authoritative project/portfolio state and reviewed transitions used to decide what work should be attempted.

### Structured invocation input

At invocation time, Calathea selects a bounded snapshot of canonical and imported state:

```yaml
workflow_invocation:
  operation: portfolio.orientation.review
  project_state_revision: 42
  project_state_digest: sha256:...
  selected_evidence_refs: []
  selected_context_refs: []
  selected_example_refs: []
  instruction_profile: orientation-review
  requested_role: reviewer
  requested_capabilities: []
  constraints: {}
  output_contract: recommendation-draft/v1
```

These field names are illustrative. A portable schema should be owned by the reusable Invokrum contract; Calathea should map to it rather than create a parallel protocol.

### Derived runtime instructions

Invokrum-composed instruction bytes, provider messages, prompt templates, and transport serialization are derived material.

Rules:

- a rendered prompt must be attributable to the structured invocation and exact instruction identity;
- changing authority-relevant Calathea state requires a new invocation identity;
- mutable chat/history is not an implicit extension of canonical workflow state;
- post-resolution mutation of instruction content creates a new artifact identity under the existing exact-byte rule;
- provider request formatting does not turn the entire request into an Invokrum-attested artifact.

### Runtime-private working state

The runtime may maintain arbitrary intermediate computation. Calathea neither defines nor depends on its representation.

Runtime-private state:

- cannot authorize capabilities or effects;
- cannot silently mutate Calathea source-of-truth state;
- is not automatically imported into feedback/learning;
- need not be serializable for Calathea replay/audit claims.

## State revision binding

Every AI invocation that can influence a `RecommendationDraft` should identify the exact Calathea state revision used to assemble it.

If material source-of-truth state changes between assembly and execution:

```text
state r42 -> assemble invocation i7
state r43 -> material project change
invocation i7 -> stale
```

Calathea should reject, reassemble, or explicitly re-evaluate the invocation according to the operation's policy. It must not silently execute `i7` against `r43` while recording only the newer state.

This parallels existing stale-state/optimistic-concurrency discipline and prevents AI context from becoming a hidden second source of truth.

## Context selection

Selected repository excerpts, issue text, historical records, examples, and retrieved memory are explicit invocation inputs or references.

They remain data unless intentionally promoted through the trusted instruction/configuration path.

Context selection must not:

- select a more privileged Invokrum profile;
- widen requested capabilities;
- override canonical Calathea project state;
- self-assert an authenticated runtime principal;
- make imported text authoritative instruction material.

## Capability and authority boundary

Calathea may structurally request a role, capability, or constraint. That request is not executable authority.

```text
Calathea request
      |
      v
Invokrum contract/materialization
      |
      v
authenticated runtime principal (where used)
      |
      v
policy/approval decision (where governed)
      |
      v
bounded effect
```

This preserves clean ownership:

- Calathea says what workflow operation is being attempted;
- Invokrum makes the execution contract/materialization deterministic;
- an identity layer establishes who/what is invoking where required;
- a governance boundary decides whether consequential effects are allowed;
- a runtime performs the computation/execution.

Specific identity, governance, and runtime products remain optional adapters and are not dependencies of the Calathea core.

## Feedback and learning

Model/runtime output never mutates canonical state directly.

The existing Calathea pattern remains:

```text
untrusted output
  -> validation
  -> RecommendationDraft / evidence
  -> review / accepted transition
  -> canonical state revision
```

Any future memory/learning feedback must reuse the same reviewed state-transition semantics. Intermediate runtime computation is not a memory source merely by existing.

## Runtime portability

The conceptual invocation should support multiple execution styles without changing Calathea's domain model:

1. ordinary request/response LLM provider;
2. supervisor/specialist runtime;
3. iterative/recurrent runtime whose private computation is not serialized as prompt text.

The provider adapter may map the structured invocation differently for each target, but must preserve exact semantic inputs, output/evidence identity, and state-revision binding.

## Failure modes to test

- stale Calathea state revision after invocation assembly;
- runtime context includes data not present in selected context refs;
- imported data attempts to select a privileged instruction profile;
- provider adapter mutates authoritative instruction bytes after Invokrum resolution;
- capability request is treated as authority without authentication/governance;
- runtime output attempts to mutate canonical state without a reviewed transition;
- a resumed/retried invocation is incorrectly rebound to newer project state;
- provider-private runtime state is mistakenly treated as canonical evidence.

## Relationship to existing docs

This note refines rather than replaces:

- `docs/architecture/invokrum-instruction-boundary.md`;
- `docs/architecture/runtime-boundaries.md`;
- RFC-0005 state/history/source-of-truth semantics;
- RFC-0006 project lifecycle/transitions;
- RFC-0008 evidence/explanation/trace semantics;
- review/feedback/learning semantics.

The note should be reconciled into the Invokrum boundary when the portable invocation-state contract stabilizes.

## Research motivation

BDH-CQ (arXiv:2608.09888) is a useful external signal because it demonstrates that context acquisition and iterative computation can be architecturally separate. Calathea's durable response is therefore to keep workflow semantics structured and model-neutral, not to depend on prompt text as the universal program representation.
