# Calathea Community Kernel RFC Governance

## Authority

Public RFCs in this repository define durable **kernel behavior, compatibility, safety, or extension contracts**.

They do not own the complete private Calathea product strategy, roadmap, full use-case model, planning/review/lifecycle policy, prioritization philosophy, calibration, AI paved road, dogfood, or private operating assumptions.

The private `hackelia-micrantha/calathea` repository is canonical for those product decisions.

## Publication test

Before adding or expanding a public RFC, ask:

> Must an independent community consumer or implementation know this decision for compatibility, safety, contribution, useful standalone operation, or an intentionally supported extension surface?

- **Yes:** document the minimum generic public contract here.
- **No:** keep the decision private.
- **Mixed:** split the concern; public owns interface/behavior/versioning, private owns product rationale/policy/defaults/calibration.

A concept being broadly reusable is not sufficient reason to make the entire product decision public.

## Status model

- **Proposed** — under review; not yet a public compatibility commitment.
- **Accepted** — authoritative for the public-kernel scope stated by the RFC.
- **Superseded** — replaced by a named later public contract; retained for history.
- **Rejected** — considered but not adopted.

Material compatibility changes use an amendment or superseding RFC rather than silently rewriting accepted behavior.

## Existing RFC 0000–0008 migration classification

These RFCs were promoted when this repository was treated as canonical for broad reusable product/domain semantics. Issue #48 is contracting that boundary.

The existing files remain historical/current implementation references until each contraction is reviewed. Target ownership is:

| RFC | Target public scope |
| --- | --- |
| [0000 — Conceptual Domain Model and Canonical Terminology](../rfc_0000_conceptual_domain_model.md) | **Split** — retain serialized/core identity/versioning invariants required by public behavior; private owns complete product model/terminology |
| [0001 — Evaluation Semantics](../rfc_0001_evaluation_orientation_semantics.md) | **Split** — retain deterministic evaluation contract; private owns calibration and product interpretation |
| [0002 — Orientation Engine Semantics](../rfc_0002_orientation_engine_policy_semantics.md) | **Mostly public** where queueing/tie-breaking/diagnostics are exposed kernel behavior; private owns portfolio defaults/policy strategy |
| [0003 — Review, Feedback, and Calibration](../rfc_0003_review_feedback_and_learning_semantics.md) | **Mostly private** — retain only concrete public state/compatibility mechanisms if exposed |
| [0004 — AI Interaction and Governance Boundary](../rfc_0004_ai_governance_paved_road_and_tooling_policy.md) | **Mostly private** — public only generic adapter validation/safety/interoperability required by a supported surface |
| [0005 — State, History, and Source of Truth](../rfc_0005_state_history_and_source_of_truth.md) | **Split** — retain persistence/replay/migration invariants required by public formats; private owns broader product authority/retention policy |
| [0006 — Project Lifecycle](../rfc_0006_project_lifecycle_and_transitions.md) | **Mostly private** — public only serialized state/transition behavior required by supported kernel workflows |
| [0007 — Policy Model and Decision Semantics](../rfc_0007_policy_model_and_decision_semantics.md) | **Split** — retain public evaluator/composition contract; private owns policy philosophy/defaults/calibration |
| [0008 — Evidence, Explanation, Provenance, Trace](../rfc_0008_evidence_explanation_and_trace_semantics.md) | **Split** — retain public evidence/provenance shapes and safety/redaction invariants; private owns broader product review/explanation strategy |

Until a file is narrowed, its historical breadth must not be used as precedent for adding new private-product semantics publicly.

## RFC rules

1. RFCs own durable public-kernel semantics, not incidental implementation structure or complete product strategy.
2. State the externally observable/publicly relied-on behavior the RFC protects.
3. Distinguish public mechanism from private product policy/defaults/calibration.
4. Compatibility, migration, validation, failure/recovery, security, and privacy implications must be explicit.
5. Accepted compatibility history is not silently rewritten.
6. Internal implementation choices belong in ADRs when they do not alter a public contract.
7. If no external/public contract is affected, prefer an ADR, issue, test, or private product RFC rather than broadening the public RFC surface.

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

When the work originates from private Calathea, the cross-repository trace is:

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

## Deferred/public non-goals

Do not use public RFCs to speculatively define:

- complete private planning/review/lifecycle strategy;
- private AI model routing or prompt/profile policy;
- automatic policy/weight/heuristic learning;
- a generic plugin marketplace or policy DSL;
- private stakeholder/maintenance strategy;
- hosted control-plane/product strategy;
- dogfood-derived defaults without a deliberate publication decision.

## Template

Use [the RFC template](template.md) for new public-kernel contract decisions, applying the publication test above first.
