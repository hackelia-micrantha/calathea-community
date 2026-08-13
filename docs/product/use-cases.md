# Calathea End-to-End Use Cases

## Status

Accepted for v0 planning

- **Owner:** Calathea maintainers
- **Decision date:** 2026-08-03
- **Product source of truth:** [Product requirements document](prd.md)

## Purpose

This document defines the user-visible workflows that Calathea must support or deliberately defer. It is technology-neutral and constrains the domain model, RFCs, architecture, tests, and MVP roadmap.

The selected v0 vertical slice is **UC-01: Orient a portfolio of existing projects**. UC-02 contributes only the minimum registration behavior required by UC-01. UC-03 through UC-05 must remain compatible with the core contracts but are not required for the first release unless the roadmap explicitly promotes them.

## Authority versus system of record

These terms are distinct:

- **Decision authority** identifies the actor allowed to approve or change Calathea-owned state.
- **System of record** identifies where an authoritative record is stored.
- **Source authority** identifies the external system that owns imported facts such as repository or issue state.

For v0, the maintainer is the only canonical decision authority. Calathea is the system of record for its own portfolio records, evaluations, orientation runs, and decisions. External systems remain authoritative for their own repository, issue, pull-request, and CI state.

## Global actors

| Actor | Responsibility | Authority |
| --- | --- | --- |
| Maintainer | Owns the local portfolio and its decisions | May create or amend Calathea-authored state and accept, reject, or override recommendations |
| Calathea deterministic core | Validates inputs, evaluates versioned semantics, and derives recommendations | May create derived and historical records; may not approve its own recommendations or mutate external systems |
| External source adapter | Reads scoped repository or project-system signals | May create imported observations only; may not silently convert them into canonical Calathea state |
| AI assistant | Produces optional drafts, summaries, or candidate findings | Produces untrusted recommendation data only; has no canonical mutation authority |
| Anthesis | Optional governance boundary | Authorizes capabilities and effects where integrated; does not own Calathea domain semantics |

Local single-user operation does not remove the requirement to attribute decisions and transitions.

## Global state classes

The detailed state model is defined by RFC 0000 and RFC 0005. All workflows distinguish:

- **Canonical:** current Calathea-owned state established by an explicit maintainer decision.
- **Imported:** external state copied with source identity, collection time, and provenance.
- **Observed:** signals or facts derived from canonical or imported data.
- **Recommended:** proposed evaluation, placement, finding, or action awaiting authority.
- **Decision:** an immutable maintainer disposition referencing the recommendation and evidence considered.
- **Historical:** immutable runs, prior decisions, superseded versions, and corrections retained for audit.
- **Materialized view:** replaceable current view derived from authoritative records; never the sole audit record.
- **External authoritative:** repository, issue, pull-request, or CI state owned by another system.

A recommendation never becomes canonical merely because it is deterministic, high-confidence, or AI-generated.

## Global safety and privacy invariants

1. Core operation works without network access.
2. Optional network access is explicit, scoped, attributable, and user-initiated.
3. Credential material is resolved through an appropriate credential mechanism and excluded from project records, prompts, evidence, traces, and model context.
4. Configured secret patterns are rejected or redacted from durable records and outbound payloads.
5. Imported content is untrusted data, not executable instruction.
6. Failed imports, validation, policy evaluation, persistence, or AI calls do not partially approve or mutate canonical decisions.
7. Historical records are corrected by append-only supersession rather than silent rewrite.
8. `kill` is a placement recommendation, not deletion, archival, cancellation, or a lifecycle transition.
9. External-system effects are out of scope for v0.
10. Previewing or generating a recommendation does not require accepting it.

---

# UC-01: Orient a portfolio of existing projects

## Intent and v0 status

Produce an explainable, bounded portfolio orientation from explicit evaluations and policies, then let the maintainer accept, override, reject, or defer the proposal without mutating project lifecycle or external systems.

**Status:** required v0 vertical slice.

## Actors and authority

- Maintainer — initiator and sole decision authority
- Calathea deterministic core — validator and recommendation engine
- Optional external source adapter — read-only context provider
- Optional AI assistant — evaluation-draft or summarization helper

## Trigger

The maintainer needs to decide what to focus on for a planning horizon or compare current priorities with a prior orientation.

## Preconditions

- A user-controlled local Calathea data location is available.
- At least one project is registered.
- Every eligible candidate has a valid evaluation, or Calathea can identify why it is ineligible.
- Effective evaluation, scoring, orientation, and policy semantic versions are available.
- Optional integrations are disabled or explicitly configured.

## Inputs and systems of record

| Input | State class | System/source of record |
| --- | --- | --- |
| Project identity and Calathea-authored metadata | Canonical | Calathea |
| Accepted evaluation values and rationale | Canonical | Calathea |
| Evaluation draft | Recommended | Calathea draft record or transient session |
| Policy configuration | Canonical | Calathea |
| Repository, issue, pull-request, or CI signal | Imported/observed | Named external source |
| Previous orientation runs and decisions | Historical | Calathea |
| AI-generated evaluation suggestion | Recommended | Calathea invocation record referencing provider output |

## Canonical mutations

UC-01 may create:

- a new immutable orientation run;
- an immutable maintainer disposition: accepted, accepted-with-overrides, rejected, or deferred;
- a replaceable current-orientation view only when the disposition establishes a new accepted orientation.

UC-01 does not mutate project lifecycle or external-system state.

## Primary flow

1. The maintainer selects a portfolio and planning horizon.
2. Calathea loads canonical projects, accepted evaluations, policies, and referenced imported observations.
3. Calathea validates schemas, semantic versions, ranges, required fields, and referential integrity.
4. Calathea identifies eligible and ineligible candidates and retains a reason for every exclusion.
5. Calathea calculates versioned deterministic scores.
6. Calathea evaluates hard constraints, soft preferences, precedence, and indeterminate policy outcomes.
7. Calathea derives proposed `now`, `next`, `later`, and `kill` recommendations.
8. Calathea durably records an immutable orientation run containing inputs or stable references, semantic versions, recommendations, explanations, policy traces, and diagnostics.
9. The maintainer reviews each material placement and exclusion.
10. The maintainer chooses one disposition:
    - accept the proposal;
    - accept with valid scoped overrides;
    - reject the proposal;
    - defer a decision while retaining the run for later review.
11. Calathea records the immutable disposition and rationale.
12. If accepted, Calathea updates the current accepted-orientation view and presents a comparison with the previous accepted decision.
13. If rejected or deferred, the prior accepted orientation remains current.

## Alternate flows

### A1: Missing evaluation

- The project is ineligible for active placement.
- Calathea emits a missing-evaluation diagnostic.
- Calathea does not synthesize canonical values silently.

### A2: Stale evaluation

- Calathea applies the accepted freshness semantics: warning, penalty, ineligibility, or review requirement.
- The exact effect and threshold are visible in the trace.
- Stale data is never reused without disclosure.

### A3: Maintainer uses an AI evaluation draft

- The maintainer explicitly selects context for outbound use.
- Calathea displays destination, data categories, and retention assumptions.
- Credential material is supplied out-of-band and excluded from the model context.
- AI output is validated as an untrusted recommendation draft.
- The maintainer edits or accepts the draft before it becomes a canonical evaluation.
- Failure or cancellation leaves canonical state unchanged.

### A4: Policy conflict or indeterminate result

- Calathea identifies the conflicting or indeterminate policies.
- Hard policies are not silently relaxed.
- The run may be partial or blocked.
- The maintainer may correct inputs, amend policy in a separate canonical edit, or use an explicitly supported exception mechanism.

### A5: Maintainer overrides a placement

- Calathea validates the replacement against hard policy and queue invariants.
- A valid override records the original recommendation, replacement, actor, timestamp, rationale, and applicable exception record.
- If an override would violate a non-overridable hard policy, Calathea rejects it and explains why.
- The underlying orientation run remains unchanged.

### A6: No eligible candidates

- Calathea records an empty recommendation with exclusion diagnostics.
- The maintainer may accept, reject, or defer the empty proposal.
- No project is promoted artificially.

### A7: No changes from the prior accepted orientation

- Calathea may record a new run and comparison showing no material change.
- The maintainer may affirm the current orientation.
- A no-change decision is valid.

### A8: Maintainer rejects the proposal

- Calathea records the rejected disposition and optional rationale.
- The run remains historical evidence.
- The prior accepted orientation remains current.
- No substitute orientation is fabricated.

## Failure and recovery

| Failure | Required behavior |
| --- | --- |
| Persistence fails before the run commit | Do not claim a durable run exists; present a clearly non-durable result only if safe |
| Persistence fails before the disposition commit | The run may remain durable, but no new accepted decision exists |
| Process terminates before disposition | Resume review of the durable run or generate a new run explicitly |
| Duplicate run request | Use an idempotency key or create a clearly related distinct run; never create ambiguous duplicate decisions |
| External adapter fails | Continue without optional observations when policy permits, or mark coverage incomplete/blocked |
| Semantic version unavailable | Refuse deterministic replay and identify the missing version |
| Evidence reference unavailable | Mark evidence unavailable or stale; do not silently omit it |
| Materialized-view update fails after decision commit | Rebuild the view from the immutable decision; do not roll back or rewrite the decision silently |

## State transitions

```text
canonical evaluations + policies
        ↓ deterministic derivation
immutable orientation run (recommended)
        ↓ explicit maintainer disposition
accepted | accepted-with-overrides | rejected | deferred
        ↓ only for accepted dispositions
current accepted-orientation view
```

No project lifecycle transition occurs in UC-01.

## Evidence and audit output

The run and disposition together identify:

- portfolio, horizon, and project identities;
- evaluation, scoring, orientation, and policy versions;
- input digests or stable references;
- imported-source identity and collection time;
- eligibility and exclusion reasons;
- score components and effective score;
- policy outcomes, precedence, and exceptions;
- tie-breaking and queue limits;
- confidence and freshness semantics;
- recommended and accepted placements;
- overrides, rejection, or deferral rationale;
- actor and timestamp;
- comparison against the prior accepted decision.

## Observability requirements

- Structured validation, eligibility, and policy diagnostics
- Run, operation, and causal trace identifiers
- Explicit durable-persistence success or failure
- Machine-readable deterministic output suitable for golden tests
- Optional integration invocation metadata without credential leakage
- A verifiable no-network mode for the core workflow
- A rebuildable current-orientation view

## Completion criteria

UC-01 completes when:

- every candidate is placed or has an explicit exclusion reason;
- every material recommendation is explainable;
- an immutable orientation run is durable;
- the maintainer has recorded a disposition or intentionally left the run pending;
- every accepted decision is reproducible from retained inputs and semantic versions;
- rejected/deferred runs do not replace the current accepted orientation;
- no external or project-lifecycle state was mutated.

## Security and privacy considerations

- Core flow works offline.
- Optional outbound context requires explicit user initiation.
- Portfolio data remains outside the source repository by default.
- Read-only credentials use least privilege and remain outside project/audit content.
- Untrusted imported text cannot alter policy or command execution.

## Mapped requirements and contracts

- PRD FR-1 through FR-10, FR-13, FR-14
- Conditional PRD FR-11 and FR-12 for optional adapters and AI
- RFC 0000 domain model
- RFC 0001 evaluation semantics
- RFC 0002 orientation semantics
- RFC 0005 state/history semantics
- RFC 0007 policy model
- RFC 0008 evidence and explanation

---

# UC-02: Intake and shape a new project

## Intent and v0 status

Capture enough intent, constraints, outcomes, and uncertainty to register a project for later evaluation without turning Calathea into a general requirements-management system.

**Status:** minimum registration subset required by UC-01; fuller shaping is post-core.

## Actors and authority

- Maintainer — initiator and registration authority
- Calathea — validation and drafting support
- Optional external adapter — read-only metadata provider
- Optional AI assistant — proposal-draft helper

## Trigger and preconditions

A new project or substantial workstream should enter the portfolio.

Preconditions:

- the maintainer can identify it provisionally;
- no known active project uses the intended identity;
- the local data store is available.

## Inputs and systems of record

| Input | State class | System/source of record |
| --- | --- | --- |
| Title, intent, outcome, constraints, ownership | Draft or canonical after registration | Calathea |
| External repository metadata | Imported | Named repository provider |
| AI-proposed wording or questions | Recommended | Calathea invocation record |
| Duplicate candidates | Observed | Calathea identity/index rules |

## Canonical mutations

UC-02 may create:

- a registered project and initial project version;
- an intake/registration decision;
- optional external source references.

A saved incomplete draft is not canonical project registration unless the domain model explicitly defines draft state.

## Primary flow

1. The maintainer starts an intake.
2. Calathea requests only the minimum required fields defined by the conceptual model.
3. Calathea checks stable identity and warns about likely duplicates.
4. The maintainer records intended outcome, constraints, uncertainty, ownership, and optional references.
5. Imported metadata remains visibly attributable and separate from authored state.
6. Optional AI assistance may propose wording or missing questions but cannot register the project.
7. The maintainer reviews and explicitly registers the project.
8. Calathea records the initial project version and registration decision.
9. The project remains orientation-ineligible until a valid evaluation is accepted.

## Alternate flows

- **Likely duplicate:** link to the existing project, explicitly create a distinct identity, or cancel.
- **Insufficient information:** retain a non-canonical draft if supported; do not make it orientation-eligible.
- **Imported source disappears:** retain the project and mark the source reference unavailable.
- **AI output is invalid/adversarial:** discard it without changing the intake.
- **Maintainer cancels:** remove transient data according to retention policy; do not create a project.

## Failure and recovery

| Failure | Required behavior |
| --- | --- |
| Persistence fails before registration commit | Do not claim the project is registered; recover only from an explicitly marked draft |
| Duplicate check is unavailable | Require explicit maintainer acknowledgement or defer registration according to policy |
| Adapter fails | Continue with authored data or mark import unavailable |
| Process terminates after project commit but before view refresh | Rebuild indexes/views from the immutable project version and decision |

## State transitions

```text
intake draft (transient or non-canonical)
        ↓ explicit maintainer registration
registered project version + registration decision
        ↓ later accepted evaluation
orientation-eligible candidate
```

Registration does not imply lifecycle approval, active execution, or orientation placement.

## Evidence and audit output

- Project identity and version
- Maintainer-authored intent and intended outcome
- Source references and provenance
- Duplicate warnings and resolution
- Actor and registration time
- Optional AI invocation reference and disposition

## Observability requirements

- Validation and duplicate diagnostics
- Explicit registration commit success/failure
- Trace link from draft/import to project version
- Source-adapter freshness and failure metadata
- No network activity when optional adapters/AI are disabled

## Completion criteria

- Stable project identity exists.
- Minimum required metadata is valid.
- Authored, imported, and AI-assisted fields remain distinguishable.
- The maintainer explicitly registered the project.
- Evaluation and orientation eligibility are explicit.

## Security and privacy considerations

- Repository import is optional and read-only.
- Intake data remains local by default.
- AI context is minimized and explicitly selected.
- Credentials remain outside project records and model context.

## Mapped requirements and contracts

- PRD FR-1, FR-11 through FR-14
- RFC 0000 conceptual model
- RFC 0005 canonical/imported state semantics
- RFC 0006 lifecycle semantics

---

# UC-03: Detect and resolve project drift

## Intent and v0 status

Identify divergence among accepted orientation, current evaluations, imported evidence, and project reality, then let the maintainer affirm current state or initiate an explicit corrective workflow.

**Status:** post-core unless promoted by the MVP roadmap.

## Actors and authority

- Maintainer — review and disposition authority
- Calathea review engine — observation/finding generator
- Optional external adapter — read-only signal provider
- Optional AI assistant — summarization or candidate-finding helper

## Trigger and preconditions

Triggers include:

- manual or scheduled review;
- stale evaluation threshold crossed;
- material imported signal change;
- `now` project lacks expected evidence;
- maintainer suspects the accepted rationale no longer holds.

Preconditions:

- at least one canonical evaluation or accepted orientation decision exists;
- review, drift, and evidence-freshness semantic versions are available.

## Inputs and systems of record

| Input | State class | System/source of record |
| --- | --- | --- |
| Accepted orientation decision | Historical/canonical authority record | Calathea |
| Current evaluations and project metadata | Canonical | Calathea |
| Repository/issue/CI signals | Imported/observed | Named external source |
| Review rules and evidence thresholds | Canonical/versioned | Calathea |
| AI-suggested finding | Recommended | Calathea invocation record |

## Canonical mutations

UC-03 may create:

- an immutable review cycle;
- observations, findings, and recommendations;
- maintainer dispositions: affirm, dismiss, defer, or initiate a separate action.

It does not directly change evaluation, orientation, lifecycle, or external-system state.

## Primary flow

1. The maintainer starts a review and selects its scope.
2. Calathea gathers canonical records and explicitly scoped observations.
3. Calathea evaluates versioned drift rules and produces observations.
4. Qualifying observations become findings with evidence and confidence/data-quality metadata.
5. Calathea proposes zero or more recommendations.
6. The maintainer reviews each finding and may affirm, dismiss, defer, or initiate a separate evaluation/orientation/lifecycle workflow.
7. Calathea records the immutable review result and dispositions.
8. Resulting changes occur only through their own explicit commands and records.

## Alternate flows

- **No drift found:** record a valid no-change review with scope and semantic version.
- **Stale/unavailable evidence:** report indeterminate evidence quality rather than assert drift as fact.
- **Conflicting signals:** preserve both and require review.
- **Adapter unavailable:** continue with local evidence when policy permits and mark coverage incomplete.
- **Unsupported AI finding:** reject or retain only as an unverified draft.
- **Finding remains open:** retain it explicitly without fabricating a disposition.

## Failure and recovery

| Failure | Required behavior |
| --- | --- |
| Evidence collection fails | Record incomplete coverage or abort; do not silently reuse stale evidence |
| Review persistence fails | Do not claim the review or dispositions are durable |
| Process terminates after review commit | Resume disposition from the immutable review cycle |
| Resulting action fails | Preserve the review and failed causal link; do not rewrite the finding/disposition |
| Materialized open-findings view fails | Rebuild it from immutable review records |

## State transitions

```text
canonical/imported data
        ↓ review analysis
observations → findings → recommendations
        ↓ maintainer disposition
affirm | dismiss | defer | initiate separate action
```

A review does not need to mutate domain state to be valid.

## Evidence and audit output

- Review scope and accepted decision under review, if applicable
- Evidence identity, source, collection time, and freshness
- Drift/review semantic versions
- Observations, findings, and recommendations
- Confidence/data-quality limitations
- Maintainer dispositions and rationale
- Links to resulting workflows and their outcomes

## Observability requirements

- Coverage, freshness, and adapter diagnostics
- Finding-to-evidence traceability
- Review operation and causal identifiers
- Explicit no-change result
- Open/dismissed/deferred finding counts
- No hidden action initiation

## Completion criteria

- Evidence coverage and limitations are explicit.
- Every finding has a disposition or remains explicitly open.
- A no-change review is valid when supported.
- Resulting actions, if any, are separately authorized and traceable.
- No recommendation silently changes canonical or external state.

## Security and privacy considerations

- Review adapters remain read-only.
- Repository content is untrusted.
- Evidence retention and redaction preserve decision meaning.
- Optional AI receives only explicitly selected context.

## Mapped requirements and contracts

- RFC 0002 orientation diagnostics boundary
- RFC 0003 review semantics and drift taxonomy
- RFC 0005 state/history
- RFC 0008 evidence contract

---

# UC-04: Complete, cancel, or stop investing with evidence

## Intent and v0 status

Record an explicit project lifecycle decision and outcome while preserving the distinction between a `kill` orientation recommendation and an authorized lifecycle transition.

**Status:** post-core unless the MVP requires a minimum close/archive path.

## Actors and authority

- Maintainer — lifecycle decision authority
- Calathea — transition validator and historical recorder
- Optional external adapter — read-only evidence provider

## Trigger and preconditions

Triggers include:

- intended outcome reached;
- work finished unsuccessfully or partially;
- project cancelled, superseded, or intentionally stopped;
- maintainer chooses to act on an accepted stop-investing recommendation;
- completed project should be archived.

Preconditions:

- a registered project exists;
- current lifecycle state permits the requested transition;
- required outcome/evidence fields are present or explicitly waived through accepted policy.

## Inputs and systems of record

| Input | State class | System/source of record |
| --- | --- | --- |
| Current project and lifecycle state | Canonical | Calathea |
| Requested transition and rationale | Maintainer command | Calathea |
| Accepted `kill` recommendation, if any | Historical supporting evidence | Calathea |
| Outcome, actual effort, and authored notes | Canonical on decision | Calathea |
| Repository/issue/CI evidence | Imported/reference | Named external source |
| Lifecycle rules | Canonical/versioned | Calathea |

## Canonical mutations

UC-04 may create:

- an immutable lifecycle decision;
- an outcome record;
- a new current lifecycle state/view;
- supersession or reopen links.

It does not mutate external repositories, issues, or CI state.

## Primary flow

1. The maintainer selects a project and requested lifecycle action.
2. Calathea displays current lifecycle state, accepted orientation, relevant `kill` recommendation, and evidence requirements.
3. The maintainer records outcome, rationale, actual effort where known, and evidence references.
4. Calathea validates the transition under the versioned lifecycle semantics.
5. The maintainer confirms the transition explicitly.
6. Calathea records the immutable lifecycle decision and resulting current lifecycle state.
7. Calathea recomputes orientation eligibility from the resulting lifecycle state and accepted orientation semantics.
8. Any repository archival, issue closure, deletion, or other external effect remains outside v0.

## Alternate flows

- **Completed without success:** allowed when work is finished and the outcome is recorded as unsuccessful or partial.
- **`kill` recommendation rejected:** no lifecycle transition occurs; record the orientation disposition only.
- **Illegal transition:** reject it and show legal alternatives.
- **Missing evidence:** block, require an explicit exception, or record evidence unavailable according to lifecycle policy.
- **Reopen:** create a new authorized transition; do not rewrite prior history.
- **Stop investing without closure:** use a lifecycle transition defined by RFC 0006 rather than overloading orientation placement.

## Failure and recovery

| Failure | Required behavior |
| --- | --- |
| Validation fails | No lifecycle decision or current-state change occurs |
| Persistence fails before decision commit | Do not report success |
| Current-state view update fails after decision commit | Rebuild the view from the immutable decision |
| Evidence source becomes unavailable | Preserve provenance and mark evidence unavailable; do not erase the decision |
| Recompute of orientation eligibility fails | Preserve lifecycle decision and mark derived eligibility stale/incomplete |

## State transitions

```text
current lifecycle state
        + maintainer command
        + required evidence/exception
        ↓ validate
immutable lifecycle decision
        ↓
new current lifecycle state
        ↓ derive
orientation eligibility
```

A `kill` recommendation may support the decision but is never the transition command.

## Evidence and audit output

- Prior and resulting lifecycle states
- Transition semantic version
- Outcome status and summary
- Rationale and actual effort where available
- Evidence and provenance
- Related orientation recommendation and disposition
- Actor and timestamp
- Reopen/supersession links

## Observability requirements

- Transition validation diagnostics
- Decision and causal trace identifiers
- Explicit persistence and view-rebuild status
- Derived eligibility recalculation status
- Evidence availability/freshness indicators
- No external-effect invocation

## Completion criteria

- Requested transition is legal and explicit.
- Outcome and required evidence/exception are recorded.
- Historical state remains immutable.
- Orientation eligibility is deterministically derived or explicitly marked stale.
- No external system was mutated automatically.

## Security and privacy considerations

- Outcome evidence may be sensitive and remains local by default.
- External evidence retains provenance and retention constraints.
- Redaction/deletion must not silently invalidate required audit history.

## Mapped requirements and contracts

- PRD authority, history, recovery, and privacy requirements
- RFC 0003 outcome recording
- RFC 0005 state semantics
- RFC 0006 lifecycle semantics
- RFC 0008 evidence contract

---

# UC-05: AI-assisted recommendation under governance

## Intent and v0 status

Use optional AI assistance to reduce synthesis effort while preserving local-first operation, explicit context selection, validation, provenance, and human authority.

**Status:** optional and not MVP-critical.

## Actors and authority

- Maintainer — initiator and canonical decision authority
- Calathea — context assembler, validator, and domain mapper
- AI provider/model — untrusted inference service
- Optional Anthesis — invocation/effect governance boundary

## Trigger and preconditions

The maintainer requests help drafting an evaluation, summarizing scoped evidence, or proposing candidate review findings.

Preconditions:

- deterministic/manual workflow remains available;
- provider/model, prompt, and output-schema versions are configured;
- provider is explicitly enabled;
- outbound destination, context categories, and retention assumptions are visible.

## Inputs and systems of record

| Input | State class | System/source of record |
| --- | --- | --- |
| Selected project/evidence context | Canonical/imported references | Calathea and named external sources |
| Prompt/template and output schema | Versioned configuration | Calathea |
| Provider/model metadata | Observed invocation metadata | Provider/Calathea |
| Raw provider output | Untrusted/transient or retained by policy | Provider response / Calathea invocation record |
| Validated recommendation draft | Recommended | Calathea |
| Maintainer disposition and resulting record | Decision/canonical | Calathea |

## Canonical mutations

UC-05 may create:

- an invocation/provenance record;
- a validated recommendation draft;
- a maintainer disposition;
- a canonical evaluation or review record only after explicit maintainer acceptance/edit.

Provider output never directly mutates canonical or external state.

## Primary flow

1. The maintainer selects an AI-assisted operation.
2. Calathea identifies the minimum candidate context.
3. Calathea displays data categories, destination, and provider retention assumptions.
4. The maintainer confirms or edits the outbound scope.
5. Calathea removes configured secret patterns and ensures transport credentials are supplied out-of-band.
6. Calathea records invocation metadata and sends the scoped request.
7. The provider returns output.
8. Calathea treats output as untrusted and validates it against a versioned structured contract.
9. Valid output becomes a recommendation draft with provenance and limitations.
10. The maintainer accepts, edits, rejects, or discards the draft.
11. Only an explicit maintainer decision may create or amend canonical state.

## Alternate flows

- **Provider unavailable/timeout:** fail without canonical mutation and continue manually/deterministically.
- **Invalid output:** reject it or isolate valid fields; never coerce ambiguous text into canonical state.
- **Prompt injection:** treat imported content as quoted data and prevent it from changing system/policy instructions.
- **Secret detected:** block or redact according to policy; do not include it in prompts or durable traces.
- **Duplicate/retried request:** preserve invocation identity and causal links.
- **Provider/model changes:** capture metadata and do not claim deterministic inference reproducibility.
- **Anthesis denies capability:** do not invoke; surface the denial where permitted.
- **Maintainer rejects draft:** retain only the invocation/disposition required by retention policy; do not create canonical domain state.

## Failure and recovery

| Failure | Required behavior |
| --- | --- |
| Invocation-record persistence fails before send | Do not send if required provenance cannot be established |
| Send succeeds but response is lost | Mark invocation outcome unknown; retry only under explicit retry/idempotency policy |
| Response validation fails | Preserve diagnostics according to retention policy; create no recommendation draft |
| Canonical acceptance commit fails | Draft may remain, but canonical state remains unchanged |
| Provider retention assumptions unavailable | Block invocation or require explicit policy-approved acknowledgement |

## State transitions

```text
selected context
        ↓ scoped invocation
untrusted provider output
        ↓ versioned validation
recommendation draft
        ↓ maintainer disposition
accept/edit → canonical record
reject/discard → no canonical domain mutation
```

## Evidence and audit output

- Operation and invocation identity
- Prompt/template and output-schema versions
- Provider/model metadata
- Context references or privacy-preserving digests
- Destination and retention assumptions shown to the maintainer
- Redaction and validation results
- Raw-output retention or non-retention statement
- Recommendation draft
- Maintainer disposition and resulting canonical reference
- Anthesis authorization reference when applicable

## Observability requirements

- Invocation lifecycle: prepared, authorized, sent, received, validated, disposed
- Provider latency and failure category without sensitive payload leakage
- Retry/idempotency and causal identifiers
- Redaction/secret-detection result
- Validation diagnostics
- Canonical mutation status kept distinct from invocation success

## Completion criteria

- Outbound scope and destination were explicit.
- Secret handling completed before invocation.
- Output was structurally validated or rejected.
- Result remains distinguishable as AI-assisted.
- Canonical state changed only through maintainer authority.
- Failure leaves deterministic/manual workflows usable.

## Security and privacy considerations

- AI is optional and disabled until configured.
- Data minimization and destination disclosure are mandatory.
- Imported content is untrusted.
- Provider retention/disclosure assumptions are visible.
- Raw prompts/outputs are retained only under explicit policy and should be minimized when they contain sensitive data.

## Mapped requirements and contracts

- PRD FR-12 and privacy requirements
- RFC 0004 AI governance
- RFC 0008 evidence/provenance
- Invokrum and structured invocation architecture contracts for optional AI orchestration

---

# Traceability matrix

| Use case | v0 status | Primary requirements | Governing contracts |
| --- | --- | --- | --- |
| UC-01 Orient portfolio | Required vertical slice | FR-1–FR-10, FR-13–FR-14; conditional FR-11/FR-12 | RFC 0000, 0001, 0002, 0005, 0007, 0008 |
| UC-02 Intake project | Minimum registration subset required | FR-1, FR-11–FR-14 | RFC 0000, 0005, 0006 |
| UC-03 Review drift | Post-core unless promoted | Review, audit, reliability NFRs | RFC 0003, 0005, 0008 |
| UC-04 Lifecycle outcome | Post-core unless minimum close path required | History, authority, recovery | RFC 0005, 0006, 0008 |
| UC-05 AI assistance | Optional, not MVP-critical | FR-12, security/privacy NFRs | RFC 0004, 0008; architecture AI contracts |

## Domain requirements exposed by these use cases

RFC 0000 supports at least:

- Portfolio
- Project and project version
- Intake draft and registration decision, if drafts are persisted
- Evaluation and evaluation version
- Policy and policy version
- Orientation run
- Placement recommendation
- Eligibility/exclusion diagnostic
- Orientation disposition: accepted, accepted-with-overrides, rejected, deferred
- Override and policy exception
- Current accepted-orientation materialized view
- Imported source reference and observation
- Review cycle, finding, recommendation, and disposition
- Lifecycle decision and outcome
- Evidence/provenance reference
- Actor/authority reference
- AI invocation and recommendation draft
- Semantic-version, idempotency, supersession, and causal references

## RFC consistency requirements

The accepted RFC set ensures:

- orientation is derived recommendation state;
- accepted orientation is an explicit maintainer decision referencing a run;
- rejected/deferred runs do not replace current accepted orientation;
- lifecycle state is independent of placement;
- no-op orientation and drift reviews are valid;
- observations, findings, recommendations, dispositions, approvals, and transitions are distinct;
- automatic heuristic mutation is deferred from v0;
- drift taxonomy is owned by RFC 0003;
- confidence is not presented as calibrated probability unless supported;
- policy overrides cannot bypass non-overridable hard constraints silently;
- `kill` cannot cause destructive or lifecycle action automatically.

## MVP planning constraints

The [MVP roadmap](mvp-roadmap.md) implements only enough of UC-02 to register projects and then complete UC-01 end to end.

UC-03 and UC-04 may follow after the deterministic orientation core and state model are stable. UC-05 remains optional and must not delay a useful offline MVP.

## Definition of complete documentation

This specification is complete for planning when:

- each workflow identifies actors, triggers, preconditions, inputs, systems of record, authority, and canonical mutations;
- primary, alternate, failure, and recovery flows are explicit;
- every canonical mutation requires an actor and decision record;
- evidence, privacy, and observability requirements are stated for every use case;
- RFC and architecture dependencies are explicit;
- future implementation gaps are tracked publicly or explicitly deferred;
- UC-01 is explicitly selected as the v0 vertical slice in this document and the PRD.
