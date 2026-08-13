# RFC 0008 — Evidence, Explanation, Provenance, and Trace Semantics

## Status

Accepted

- **Owner:** Calathea maintainers
- **Decision date:** 2026-08-04
- **Depends on:** RFC 0000, RFC 0002, RFC 0003, RFC 0004, RFC 0005, RFC 0007, PRD, UC-01, UC-03, UC-05
- **Scope:** Cross-cutting evidence, explanation, provenance, redaction, causal-link, and audit-trace contracts

## Summary

Calathea uses one compatible trace model across evaluation, orientation, policy, review, lifecycle, AI assistance, and maintainer decisions.

The model separates:

1. source identity and source material;
2. evidence references and retained evidence snapshots;
3. observations and deterministic derivation steps;
4. findings, rationale, and recommendations;
5. policy decisions and exception applications;
6. maintainer dispositions and lifecycle decisions;
7. external effects, when governed elsewhere.

Evidence does not become a finding merely by being collected. A recommendation does not become a decision merely because it is well explained. A decision does not become an external effect merely because it is authorized in Calathea.

Traceability must be sufficient to explain and, where deterministic, replay a result without requiring indiscriminate retention of private content, credentials, or raw AI prompts.

## Goals

- Provide compatible primitives for all decision and recommendation workflows.
- Make source identity, freshness, transformations, and evidence availability explicit.
- Support concise user explanations and deeper audit views from the same records.
- Distinguish evidence quality from calibrated probability.
- Make missing, stale, contradictory, unavailable, or untrusted evidence visible.
- Preserve deterministic replay requirements separately from AI invocation provenance.
- Make approvals, overrides, exceptions, and corrections attributable.
- Define redaction without silently changing decision meaning.
- Avoid claiming cryptographic tamper-proofing that local storage does not provide.

## Non-goals

- A universal evidence ontology.
- A cryptographic ledger or transparency service.
- Full content archival of every external source.
- Mandatory raw prompt or model-output retention.
- Replacing Anthesis effect evidence.
- Treating every explanation as proof of correctness.
- Requiring event sourcing.
- Defining storage, database, serialization, or UI technology.

## Decision principles

1. **Source, evidence, interpretation, recommendation, decision, and effect are distinct.**
2. **Every material claim identifies its provenance or explicitly states that provenance is unavailable.**
3. **Missing evidence is not negative evidence unless a rule explicitly defines it that way.**
4. **Contradictory evidence remains visible.**
5. **Confidence is typed and scoped; it is not implicitly probability.**
6. **Deterministic traces identify exact replay inputs and semantic versions.**
7. **AI reinvocation is never deterministic replay.**
8. **Redaction is explicit and reports its effect on explanation and replay.**
9. **Human decisions identify actor, authority, rationale, and causal inputs.**
10. **Trace detail is layered rather than duplicated into incompatible summary and audit formats.**

## Conceptual chain

```text
ExternalSource / CanonicalRecord
        ↓ reference or retained snapshot
EvidenceReference
        ↓ normalized statement
Observation
        ↓ interpretation
Finding
        ↓ proposed response
Recommendation
        ↓ deterministic policy or engine derivation
PolicyDecision / PlacementRecommendation
        ↓ explicit human authority
Disposition / LifecycleDecision / PolicyException
        ↓ optional separately governed execution
ExternalEffect
```

Not every operation uses every element. Each emitted element must identify its semantic role.

## Core primitives

### ExternalSource

Stable identity for an external system, repository, document corpus, service, or provider.

It records source kind, authority owner, locator policy, and supported identity semantics. It does not imply Calathea owns or may retain source content.

### SourceReference

A locator and identity reference to external-authoritative material.

It may include:

- source identity;
- external object identity;
- locator or opaque reference;
- source revision, commit, ETag, version, or equivalent identity;
- observed collection time;
- source-authored time when available;
- content type and declared trust class;
- accessibility and retention constraints.

A mutable URL alone is insufficient for deterministic replay.

### EvidenceReference

A provenance-bearing reference used to support, challenge, or contextualize a domain claim.

An evidence reference records:

- stable evidence identity;
- source or canonical-record reference;
- evidence role: `supporting`, `conflicting`, `contextual`, or `missing_expected`;
- acquisition or authoring time;
- source revision or content identity when available;
- selected scope or excerpt coordinates;
- transformation chain;
- freshness classification and rule version;
- trust/data-quality classification;
- availability state;
- retention and redaction state;
- optional content digest;
- confidentiality classification.

Evidence identity does not require embedding the full content.

### EvidenceSnapshot

An optional retained immutable copy or normalized extract of evidence content.

Snapshots are used only when retention is permitted and needed for replay, audit, offline operation, or resilience to external-source change.

A snapshot records:

- evidence-reference identity;
- bytes or structured content identity;
- capture time;
- normalization/transformation metadata;
- digest algorithm and digest;
- retention and deletion policy;
- redaction state.

A digest detects accidental or detectable content mismatch. It does not prove authenticity against a fully compromised local system.

### Observation

An attributable statement about canonical, imported, or derived data.

An observation records:

- subject;
- statement type and structured value;
- evidence references;
- derivation type: authored, normalized import, or deterministic derived;
- semantic version;
- time and freshness;
- data-quality status;
- uncertainty or indeterminate status.

An observation is not automatically a finding or recommendation.

### Finding

An evidence-backed interpretation that an expectation, rationale, evaluation, placement, lifecycle state, or other claim may no longer hold.

A finding records:

- subject and expectation challenged;
- supporting, conflicting, and missing-expected evidence;
- finding type and materiality;
- explanation and reason codes;
- confidence/data-quality assessment;
- assumptions;
- semantic version;
- generated-by identity;
- review-cycle reference when applicable.

### Recommendation

A proposed response to observations, findings, or deterministic derivation.

Recommendations are non-authoritative and identify:

- target workflow;
- proposed action or content;
- rationale;
- evidence and finding references;
- policy and constraint context;
- assumptions and uncertainty;
- producer and semantic version.

### TraceEntry

A typed immutable explanation component within an operation trace.

Supported conceptual types include:

- input selection;
- validation result;
- score calculation;
- freshness classification;
- policy applicability;
- policy decision;
- policy conflict;
- exception application;
- capacity effect;
- tie-break;
- exclusion or indeterminate diagnostic;
- recommendation derivation;
- redaction notice;
- skipped-step reason;
- failure diagnostic.

Trace entries use stable reason codes plus human-readable explanations.

### OperationTrace

The immutable trace for one deterministic or nondeterministic operation.

It records:

- operation identity and type;
- subject and scope;
- actor/requester identity;
- authority context;
- start and completion times;
- input record references and digests;
- semantic versions;
- ordered trace entries;
- output references;
- parent/causal operation references;
- completeness status;
- redaction and retention metadata;
- deterministic replay classification.

An operation trace is not itself a maintainer decision.

### DecisionRationale

Structured rationale attached to a maintainer-authored disposition, override, policy exception, evaluation acceptance, or lifecycle decision.

It records:

- selected option;
- alternatives considered when material;
- evidence, recommendation, run, finding, or policy references;
- assumptions;
- explicit policy exceptions;
- actor and authority;
- time;
- free-form explanation where needed.

### PolicyExceptionApplication

A deterministic trace record showing that one valid `PolicyException` was applied to one policy decision in one workflow.

It is distinct from the immutable exception itself and records:

- exception identity and version/reference;
- policy decision affected;
- subject and workflow scope;
- validity checks at operation time;
- use ordinal or application identity;
- effective outcome;
- reason and trace links.

Maximum-use, expiry, and revocation checks rely on immutable application or revocation records and rebuildable projections. They never mutate the original exception in place.

## Provenance model

Every evidence-bearing or decision-bearing record identifies, as applicable:

- origin system or canonical record;
- original producer;
- collector or importer;
- authoring and collection times;
- source revision/content identity;
- transformations and normalizations;
- deterministic or nondeterministic producer;
- semantic versions;
- custody/retention boundary;
- redaction or deletion state.

Provenance may use references rather than embedded content. Missing provenance is explicit and may make a result indeterminate or unusable under policy.

## Evidence availability states

Evidence references use an explicit availability state:

- `available_embedded`;
- `available_external`;
- `temporarily_unavailable`;
- `permanently_unavailable`;
- `redacted`;
- `deleted_with_tombstone`;
- `identity_only`;
- `access_denied`;
- `integrity_mismatch`.

Unavailable evidence is not silently omitted from explanations.

Evidence availability and operation replay status are separate dimensions. An unavailable evidence reference can contribute to `partially_reproducible` or `not_reproducible`, but only RFC 0005 replay semantics determine the operation-level status.

## Freshness

Freshness is derived from a versioned rule and explicit timestamps. It is not an intrinsic permanent property of evidence.

A freshness assessment records:

- applicable planning horizon;
- source-authored and collected times where known;
- rule/version;
- classification;
- threshold values;
- reason.

Possible classifications include `fresh`, `aging`, `stale`, `unknown`, and `not_time_sensitive`.

Freshness assessment never mutates the underlying evidence record.

## Trust and data quality

Trust and quality are separate from confidence in a conclusion.

Evidence may be classified by:

- authority: canonical, external-authoritative, imported, observed, generated;
- integrity status: verified identity, digest matched, unchecked, mismatch;
- collection quality: complete, partial, sampled, failed, unknown;
- source trust: configured trusted, untrusted content, unknown;
- relevance: direct, indirect, contextual;
- contradiction status.

Repository content, issue text, comments, documents, and AI output are untrusted instruction sources even when they are valid evidence data.

## Confidence semantics

Every confidence-like field declares its type and scope.

Supported conceptual types include:

- `evaluation_input_confidence`: ordinal quality/completeness of an evaluation;
- `finding_confidence`: ordinal support for an interpretation;
- `data_quality`: categorical evidence coverage/quality;
- `model_reported_confidence`: untrusted provider output;
- `calibrated_probability`: permitted only with a documented calibration method, population, metric, and version.

A numeric value in `[0,1]` is not a probability merely because of its range.

User-visible output must name the confidence type. Model-reported confidence must never be presented as calibrated evidence confidence without validation.

## Contradictory evidence

Contradictory evidence is retained and linked rather than resolved by deletion.

The trace records:

- claims in conflict;
- evidence supporting each claim;
- source authority and freshness;
- resolution rule if deterministic;
- unresolved assumptions;
- resulting status: resolved, indeterminate, require-review, or excluded.

A human decision may choose a position but must not erase the contradiction from history.

## Deterministic replay contract

RFC 0005 owns the canonical operation-level replay-status vocabulary:

- `reproduced`;
- `partially_reproducible`;
- `not_reproducible`;
- `invalidated`;
- `not_applicable`.

A deterministic operation is replayable only when its trace or referenced records identify:

- complete immutable canonical inputs;
- retained evidence snapshots or stable content identities required by semantics;
- exact policy-set and evaluator versions;
- evaluation, scoring, orientation, lifecycle, explanation, and schema versions;
- exact-decimal arithmetic and ordering rules where applicable;
- planning horizon and runtime constraints;
- operation identity/idempotency scope;
- deterministic component versions.

Replay produces equivalent domain output, not necessarily identical timestamps or randomly assigned identifiers.

Evidence availability maps into replay status as follows:

- all required retained inputs/evidence available and valid, with equivalent deterministic output -> `reproduced`;
- one or more required inputs/evidence unavailable, but an actually executed and clearly delimited deterministic subset can be reproduced -> `partially_reproducible`;
- required inputs/evidence unavailable and no meaningful deterministic subset sufficient for a replay claim can be completed -> `not_reproducible`;
- required retained content fails integrity or semantic validation, including `integrity_mismatch` -> `invalidated`;
- the operation is nondeterministic by definition, such as AI inference -> `not_applicable`.

Missing/redacted/deleted evidence by itself does not justify `partially_reproducible`. The replay trace must record which deterministic subset was actually replayed and the boundary preventing full reproduction.

Current external content is never substituted silently for unavailable historical evidence.

## AI provenance and reproducibility

An AI invocation trace records, subject to retention policy:

- invocation and operation identity;
- provider/model metadata;
- prompt/template and output-schema versions;
- selected-context references or privacy-preserving digests;
- destination disclosure;
- secret/redaction checks;
- validation results;
- recommendation-draft references;
- causal links to later maintainer actions.

Raw prompts and model output are retained only when explicitly permitted and necessary.

Reinvoking a provider is a new nondeterministic operation. Its deterministic replay status is `not_applicable`; reinvocation is never partial or full reproduction of the original model result.

AI-produced claims must distinguish quoted/source-backed material from generated interpretation.

## Redaction and deletion

Redaction creates an explicit redacted representation or superseding metadata record. It does not silently rewrite historical meaning.

A redaction record identifies:

- target record or field;
- actor and authority;
- reason and policy basis;
- redaction method;
- time;
- whether identity/digest/tombstone remains;
- effect on explanation, validation, and replay.

User-visible explanations must not expose redacted content through derived summaries.

Deletion follows RFC 0005. When a tombstone remains, it records enough identity and causal metadata to explain why a referenced item is unavailable without retaining prohibited content.

## Layered explanation views

The same trace supports three conceptual views.

### Summary view

Answers:

- What happened?
- Why is this project here or excluded?
- What is the principal uncertainty or policy effect?
- What action is required from the user?

### Detail view

Adds:

- score components;
- policy decisions;
- freshness and confidence types;
- tie-breaks;
- evidence summaries;
- conflicting or missing inputs;
- override and exception references.

### Audit view

Adds:

- exact input identities and versions;
- ordered trace entries;
- full provenance metadata;
- semantic versions;
- redaction/retention states;
- causal and supersession links;
- canonical replay status and, for partial reproduction, the replayed subset boundary.

Views are projections over the same underlying records. They must not invent materially different reasons.

## Minimum contracts by workflow

### Evaluation acceptance

Must expose axis values, rationale, confidence type, evidence/provenance, actor/authority, scheme/semantic version, and superseded version where applicable.

### Orientation run

Must expose all considered projects, exact evaluation/lifecycle/policy inputs, score derivation, eligibility, every material policy decision, capacity effects, tie-breaks, recommendations or diagnostics, semantic versions, and replay status.

### Orientation disposition

Must expose run reference, disposition kind, actor/authority, rationale, overrides, policy exceptions and exception applications, prior accepted-orientation relationship, and time.

### Review cycle

Must expose scope, evidence coverage, observations, findings, conflicting/missing evidence, recommendations, dispositions, no-change or indeterminate conclusion, and follow-on workflow references.

### Lifecycle decision

Must expose prior/requested/resulting states, actor/authority, rationale, evidence/outcome references, policy validation, semantic version, and supersession/reopen relationship.

### AI invocation

Must expose purpose, selected context categories, provider/model/template/schema versions, validation status, recommendation draft, retention/redaction status, and later human action references.

## Failure and partial traces

Failed operations may retain a minimal immutable diagnostic trace when permitted.

A partial trace identifies:

- last completed stage;
- unavailable or failed input;
- whether any authoritative record was committed;
- retry/idempotency identity;
- redaction or retention constraints.

A partial operation trace is not the same concept as `partially_reproducible`. The latter requires an actual replay of a deterministic subset under RFC 0005.

A failed operation must not appear as a complete successful trace.

## Integrity expectations

Calathea should use content identities, digests, immutable records, append-only corrections, and restore validation where appropriate.

These controls can detect many accidental changes and some mismatches. They do not establish tamper-proof history against a fully compromised host, administrator, or storage layer.

Stronger external anchoring or transparency logs require a separate RFC and threat model.

## Privacy and security

- Credentials and secret values are never evidence, rationale, prompt context, or trace content.
- Sensitive portfolio data remains local by default.
- Outbound evidence selection is explicit and minimized.
- Trace retention is purpose-limited.
- Raw AI context is not retained by default.
- Exports preserve confidentiality labels and redaction states.
- Untrusted content cannot alter instructions, scope, capability, or authority.
- Derived explanations must not leak redacted source content.
- Source locators may themselves be sensitive and support redaction.

## Testing requirements

Golden tests must cover:

- complete deterministic orientation trace;
- missing evidence;
- stale evidence;
- contradictory evidence;
- external source changed after collection;
- redacted evidence with replay impact;
- deleted evidence with tombstone;
- digest mismatch;
- concise and audit views remaining semantically consistent;
- confidence type displayed correctly;
- AI recommendation draft separated from later acceptance;
- valid, expired, revoked, and overused policy exceptions through separate application records;
- partial operation failure;
- each canonical replay status, including an actually replayed deterministic subset for `partially_reproducible`.

## Consequences

### Benefits

- Cross-workflow explanations share one vocabulary.
- Concise UX and deep audit views remain compatible.
- Deterministic replay requirements and status values are explicit.
- AI provenance does not pretend determinism.
- Redaction and deletion remain visible without leaking content.
- Policy exceptions and applications preserve immutability.

### Costs

- More explicit references and trace entries.
- Retention and redaction require deliberate design.
- Some external evidence cannot be replayed without snapshots.
- Explanations must be generated from structured traces rather than ad-hoc strings.

## Deferred decisions

- Storage and serialization format.
- Cryptographic signing or transparency logs.
- Remote evidence vaults.
- Cross-system provenance standards.
- Full external-content archival.
- Organization-wide retention administration.
- General evidence query language.
- Automatic truth resolution across conflicting sources.

## Acceptance criteria

This RFC is accepted because:

- orientation, review, policy, lifecycle, and AI workflows use compatible trace primitives;
- source identity, provenance, freshness, availability, and trust are explicit;
- evidence, observation, finding, recommendation, decision, and effect remain distinct;
- confidence fields declare type and are not implied probabilities;
- missing and contradictory evidence produce explicit diagnostics;
- deterministic traces identify sufficient replay inputs and semantic versions;
- evidence availability maps deterministically into RFC 0005 replay statuses without conflating unavailable evidence with partial reproduction;
- AI reinvocation is explicitly `not_applicable` to deterministic replay;
- approvals, overrides, lifecycle decisions, and exceptions are attributable;
- policy-exception application/use is separate from the immutable exception;
- redaction and deletion report their impact on explanation and replay;
- summary, detail, and audit views derive from the same trace records;
- integrity claims remain bounded to implemented controls;
- sensitive data and secrets are not exposed through traces or derived explanations.