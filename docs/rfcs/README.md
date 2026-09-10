# Calathea Community Kernel RFC Governance

## Authority

Public RFCs in this repository define durable **kernel behavior, compatibility, safety, or extension contracts**.

They do not own the complete Calathea product strategy, roadmap, planning/review/lifecycle policy, prioritization philosophy, calibration, AI paved road, dogfood, or private operating assumptions. Those are private product decisions in `hackelia-micrantha/calathea`.

## Publication test

Before adding or expanding a public RFC, ask:

> Must an independent community consumer or implementation know this decision for compatibility, safety, contribution, useful standalone operation, or an intentionally supported extension surface?

- **Yes:** document the minimum generic public contract here.
- **No:** keep the decision private.
- **Mixed:** split the concern; public owns interface/behavior/versioning, private owns product rationale/policy/defaults/calibration.

"Reusable" alone is not sufficient reason to publish the full product decision.

## Status model

- **Proposed** — under review; not yet a public compatibility commitment.
- **Accepted** — authoritative for the stated public-kernel scope.
- **Deferred** — no current standalone/public surface requires the broader behavior; retained as a boundary marker.
- **Superseded** — replaced by a named later public contract; retained for history.
- **Rejected** — considered but not adopted.

Material compatibility changes use an amendment or superseding RFC rather than silently changing supported behavior.

## Current RFC index

RFC 0000–0008 were originally promoted under a broader public-product ownership model. Issue #48 narrows those files in place so historical links remain valid while current head exposes only the community-kernel contract.

| RFC | Status | Current public scope |
| --- | --- | --- |
| [0000 — Community Kernel Record and Identity Model](../rfc_0000_conceptual_domain_model.md) | Accepted | Record identity/versioning/state distinctions needed by supported public process/file/schema behavior |
| [0001 — Deterministic Evaluation Contract](../rfc_0001_evaluation_orientation_semantics.md) | Accepted | Deterministic evaluation inputs/outputs/versioning; excludes private calibration and strategic interpretation |
| [0002 — Deterministic Orientation Contract](../rfc_0002_orientation_engine_policy_semantics.md) | Accepted | Deterministic placement/ordering/capacity/diagnostic behavior exposed by the kernel |
| [0003 — Public Review Extension Boundary](../rfc_0003_review_feedback_and_learning_semantics.md) | Deferred | Boundary for future concrete public review state; complete review/learning strategy is private |
| [0004 — Optional AI Adapter Safety Contract](../rfc_0004_ai_governance_paved_road_and_tooling_policy.md) | Accepted | Validation, non-authority, data/safety and adapter compatibility constraints only |
| [0005 — Public State, History, Replay, and Persistence Invariants](../rfc_0005_state_history_and_source_of_truth.md) | Accepted | Persisted/history/replay guarantees required by supported public formats |
| [0006 — Lifecycle Compatibility Boundary](../rfc_0006_project_lifecycle_and_transitions.md) | Deferred | Separation of orientation from lifecycle plus any explicitly exposed lifecycle serialization/transition behavior |
| [0007 — Policy Evaluation and Composition Compatibility](../rfc_0007_policy_model_and_decision_semantics.md) | Accepted | Evaluator identity/versioning, deterministic composition, result classes, validation and trace compatibility |
| [0008 — Evidence, Provenance, and Trace Compatibility](../rfc_0008_evidence_explanation_and_trace_semantics.md) | Accepted | Evidence/source identity, provenance, trace/reason-code, redaction and replay compatibility |

The complete product counterparts are private and may be broader than these public contracts.

## RFC rules

1. RFCs own durable public-kernel semantics, not incidental implementation structure or complete product strategy.
2. State the externally observable/publicly relied-on behavior the RFC protects.
3. Distinguish public mechanism from private product policy/defaults/calibration.
4. Compatibility, migration, validation, failure/recovery, security, and privacy implications must be explicit.
5. Accepted compatibility history is not silently rewritten.
6. Internal implementation choices belong in ADRs when they do not alter a public contract.
7. If no external/public contract is affected, prefer an ADR, issue, test, or private product RFC rather than broadening this RFC surface.

## Traceability

Public implementation work should preserve:

```text
public kernel use case / concrete compatibility need
        ↓
public RFC or ADR when durable
        ↓
issue
        ↓
PR
        ↓
synthetic validation / compatibility evidence
```

When work originates from private Calathea:

```text
private product requirement
        ↓
minimum extracted public contract
        ↓
public issue / RFC-or-ADR / PR / validation
        ↓
private reviewed public-core pin
        ↓
private dogfood / product review
```

## Amendment and supersession

A material public RFC change must state:

- which supported public behavior changes;
- why the previous contract is insufficient;
- migration/compatibility consequences;
- security/privacy implications;
- validation required before adoption;
- whether the prior RFC is amended or superseded.

Superseded RFCs remain available for history.

## Public non-goals

Do not use public RFCs to speculatively define:

- complete private planning/review/lifecycle strategy;
- private AI model routing or prompt/profile policy;
- automatic policy/weight/heuristic learning;
- a generic plugin marketplace or unrestricted policy DSL;
- private stakeholder/maintenance strategy;
- hosted control-plane/product strategy;
- dogfood-derived defaults without a deliberate publication decision.

## Template

Use [the RFC template](template.md) for new public-kernel contract decisions, applying the publication test first.
