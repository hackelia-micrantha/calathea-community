# Calathea RFC Governance

This directory defines the lifecycle and authoritative index for Calathea RFCs. RFC files remain at `docs/rfc_*.md` for stable links.

## Status model

- **Proposed** — under review; not yet authoritative.
- **Accepted** — authoritative semantic decision.
- **Superseded** — replaced by a named later RFC; retained for history.
- **Rejected** — considered but deliberately not adopted.

The Calathea maintainer is the v0 decision authority.

## Current RFC index

| RFC | Status | Scope |
| --- | --- | --- |
| [0000 — Conceptual Domain Model and Canonical Terminology](../rfc_0000_conceptual_domain_model.md) | Accepted | Foundational entities, authority, terminology, identity/versioning |
| [0001 — Evaluation Semantics](../rfc_0001_evaluation_orientation_semantics.md) | Accepted | Evaluation axes, scoring, confidence, freshness inputs, calibration |
| [0002 — Orientation Engine Semantics](../rfc_0002_orientation_engine_policy_semantics.md) | Accepted | Deterministic orientation, queue selection, tie-breaking, diagnostics |
| [0003 — Review, Feedback, and Calibration Semantics](../rfc_0003_review_feedback_and_learning_semantics.md) | Accepted | Reviews, findings, recommendations, dispositions, calibration signals |
| [0004 — AI Interaction and Governance Boundary](../rfc_0004_ai_governance_paved_road_and_tooling_policy.md) | Accepted | Optional AI, context/output validation, authority/governance boundary |
| [0005 — State, History, and Source-of-Truth Semantics](../rfc_0005_state_history_and_source_of_truth.md) | Accepted | Authority, immutable history, concurrency, replay, retention/recovery |
| [0006 — Project Lifecycle and Legal Transitions](../rfc_0006_project_lifecycle_and_transitions.md) | Accepted | Lifecycle states, transitions, outcomes, correction/recovery |
| [0007 — Policy Model, Composition, and Decision Semantics](../rfc_0007_policy_model_and_decision_semantics.md) | Accepted | Versioned policies, deterministic evaluators, composition, exceptions |
| [0008 — Evidence, Explanation, Provenance, and Trace Semantics](../rfc_0008_evidence_explanation_and_trace_semantics.md) | Accepted | Cross-cutting evidence/provenance/trace/redaction/replay contracts |

The inline status in each RFC is synchronized with this index as part of the public-core migration.

## RFC rules

1. RFCs own durable product/domain semantics, not incidental implementation structure.
2. Accepted RFC history is not silently rewritten. Material changes use an amendment with explicit rationale or a superseding RFC.
3. Corrections that do not alter semantics may be made in place when the change and provenance remain clear.
4. An RFC must distinguish recommendation, human decision, canonical mutation, authorization, and external effect where relevant.
5. Security, privacy, failure/recovery, compatibility, and validation implications must be explicit.
6. Unresolved questions must identify a later RFC/ADR/issue or the concrete condition that would justify reopening the decision.
7. Implementation-specific choices belong in ADRs when they do not change RFC semantics.

## Traceability

Reusable-core work should preserve:

```text
PRD → use case → RFC / ADR → public issue → PR → validation evidence
```

The RFCs were originally accepted while implementation planning lived in a private-first repository. Historical bare issue numbers in earlier revisions are provenance only; they are not normative public dependencies. During migration, semantic delegations were converted to the RFC/ADR that actually owns the decision.

New reusable implementation work belongs in `calathea-community` issues and PRs. Private dogfood/data/configuration work remains in the private composition repository.

## Amendment and supersession

A material RFC change must state:

- which accepted requirement changed;
- why the previous decision is no longer sufficient;
- migration and compatibility consequences;
- security/privacy implications;
- validation needed before adoption;
- whether the prior RFC is amended or superseded.

Superseded RFCs remain in the repository and point to their replacement.

## Deferred boundaries

The accepted v0 semantics deliberately defer:

- AI as a requirement for deterministic orientation;
- autonomous external effects;
- generic plugin or arbitrary policy DSL support;
- multi-user collaboration and hosted control-plane concerns;
- automatic policy/weight/heuristic learning;
- bidirectional external synchronization.

Future work in those areas requires explicit product pressure and a new or amended RFC/ADR rather than accidental expansion through implementation.

## Template

Use [the RFC template](template.md) for new semantic decisions.
