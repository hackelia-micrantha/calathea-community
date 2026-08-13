# Calathea MVP Roadmap

## Status

Accepted for v0 implementation planning.

- **Owner:** Calathea maintainers
- **Decision date:** 2026-08-11
- **Primary use case:** UC-01 — Orient a portfolio of existing projects
- **Product boundary:** `docs/product/prd.md`
- **Architecture baseline:** `docs/architecture/`

## Migration note

This roadmap predates the public-core/private-composition split. The phase ordering and release gates remain authoritative. Implementation tasks are now tracked in `calathea-community` when they change reusable code/contracts rather than by preserving the old private issue-number map.

## Release hypothesis

Calathea v0 proves one claim:

> A maintainer can locally evaluate a portfolio, receive a deterministic and explainable orientation, explicitly accept or override it, and later reproduce and compare the decision without AI, network access, or autonomous mutation.

The release is complete when this workflow is usable end to end for at least ten projects and satisfies the PRD integrity requirements.

## Planning principles

1. Build UC-01 vertically rather than implementing each RFC as an isolated subsystem.
2. Keep the first runnable system single-process and local-only.
3. Introduce persistence early enough to exercise immutable history and recovery semantics.
4. Test deterministic semantics with golden fixtures before adding optional integrations.
5. Treat privacy, backup/restore, replay, and failure recovery as release behavior rather than cleanup work.
6. Do not implement AI, GitHub integration, review/drift workflows, or effect governance until the deterministic local workflow is complete.
7. Avoid speculative framework, plugin, and distributed-system abstractions.
8. Keep implementation issues independently reviewable and dependency-ordered.

## Phase 0 — Product and architecture baseline

**Status:** complete.

Delivered:

- PRD and v0 product boundary;
- end-to-end use cases with UC-01 selected;
- conceptual domain model and terminology;
- evaluation, orientation, review, lifecycle, policy, state/history, and evidence semantics;
- architecture baseline and ADRs;
- no Anthesis dependency;
- explicit v0 deferrals.

### Exit criteria

- No unresolved product-level ambiguity blocks UC-01 implementation.
- Recommendation, maintainer decision, lifecycle transition, authorization, and external effect are distinct.
- Core operation requires no network, AI, GitHub, or hosted service.

## Phase 1 — Executable deterministic core

### Goal

Create a runnable, storage-independent domain/application core capable of evaluating and orienting an in-memory portfolio with structured traces.

### Required work

#### 1. Runtime and repository foundation

Use the dependency direction from ADR 0001 and the Go runtime selected by ADR 0004, with public module/process ownership governed by ADR 0005.

Minimum result:

```text
CLI adapter
Application services
Domain model
Deterministic services
Port contracts
Tests
```

The technology decision must optimize for deterministic domain modeling, strong typing, local distribution, maintainability, and fast test feedback without altering accepted domain semantics.

Establish one local quality command and minimal CI for formatting/static checks plus unit/golden tests. If CI infrastructure is temporarily unavailable, the same checks remain mandatory and runnable locally; CI availability must not block domain work.

#### 2. Core domain contracts

Implement the UC-01-required subset:

- Portfolio;
- Project / ProjectVersion;
- Evaluation / EvaluationVersion;
- PolicySet / PolicySetVersion;
- immutable policy-selection decision sufficient to reconstruct the effective set;
- PolicyException;
- PolicyExceptionApplication;
- PolicyDecision;
- OrientationRun;
- PlacementRecommendation;
- OrientationDisposition;
- PlacementOverride;
- EvidenceReference / TraceEntry / OperationTrace;
- minimum lifecycle state and LifecycleDecision contracts needed by UC-01 bootstrap/eligibility.

Use explicit constructors/validation rather than persistence-shaped structs leaking into domain code.

#### 3. Evaluation semantics

Implement RFC 0001:

- validation of axis values;
- versioned scoring semantics;
- effort handling;
- ordinal confidence handling without treating confidence as calibrated probability;
- rationale requirements;
- deterministic score output.

#### 4. Policy engine subset

Implement only the built-in evaluators needed by UC-01:

- lifecycle eligibility;
- accepted-evaluation requirement;
- `now` capacity, default `3`;
- `next` capacity;
- freshness/confidence diagnostics;
- explicit exception validation where an override requires one.

No DSL or user-supplied executable policy.

#### 5. Orientation engine

Implement RFC 0002:

- candidate filtering;
- deterministic scoring/ranking;
- policy application;
- bounded queue selection;
- stable tie-breaking;
- explicit exclusions and indeterminate outcomes;
- `now / next / later / kill` recommendations;
- structured operation trace.

No project lifecycle mutation occurs as part of orientation.

### Testing gate

Golden fixtures must cover at minimum:

- stable ranking;
- score ties;
- `max_now = 3`;
- missing accepted evaluation;
- stale evaluation;
- low/unknown confidence diagnostics;
- candidate and completed/cancelled/archived lifecycle exclusion;
- approved/active lifecycle eligibility;
- exceptionable and non-exceptionable hard policies;
- `kill` remaining a recommendation only;
- identical versioned inputs producing equivalent deterministic output;
- trace reasons matching placements and exclusions.

### Security boundary checkpoint

- Domain/deterministic packages have no network/provider/storage dependencies.
- Tests prove external content cannot act as executable instruction inside the core.
- No credential-bearing type is accepted by core evidence/trace contracts.

### Demo

Given an in-memory fixture containing ten approved/active projects, run the deterministic core and print a structured orientation trace showing all placements and reasons.

### Exit criteria

- Domain/deterministic test suite runs without storage, network, AI, or external adapters.
- Golden fixtures pass reproducibly.
- Every placement or material exclusion has a structured trace reason.
- Minimal local quality checks are repeatable and CI runs them when infrastructure is available.

## Phase 2 — Local persistence and CLI vertical slice

### Goal

Turn the deterministic engine into the actual local UC-01 product.

### Required work

#### 1. Physical persistence ADR and implementation

Select the simplest local persistence technology that satisfies ADR 0002 and RFC 0005.

It must support:

- immutable historical records;
- stable identity/version lookup;
- atomic authoritative command boundaries;
- idempotency lookup;
- expected-version checks for current projections;
- rebuildable projections;
- migration/version handling;
- backup/export/restore;
- redaction/tombstone requirements.

Do not adopt event sourcing unless concrete implementation requirements justify it.

#### 2. CLI workflow

Provide commands equivalent to:

```text
calathea project add
calathea project add --bootstrap-approved --reason <text>
calathea project list
calathea evaluation add
calathea policy show
calathea policy select
calathea orient
calathea orientation show
calathea orientation accept
calathea orientation override
calathea orientation reject
calathea orientation defer
calathea orientation compare
```

Exact naming may change, but the complete UC-01 interaction must be available from one local executable.

Normal `project add` creates a `candidate`. The explicit bootstrap form is the RFC 0006 migration convenience for an existing portfolio: it records a maintainer-authorized lifecycle/bootstrap decision, the skipped intake context, actor, rationale, and exact resulting `approved` state. There is no silent default promotion to `approved`.

A later general lifecycle CLI for UC-02/UC-04 is not required to ship UC-01.

#### 3. Persistence adapters and projections

Implement application-owned ports for:

- authoritative record storage;
- projection storage;
- clock;
- identity generation.

Required rebuildable views include:

- current project version;
- current effective policy-set selection;
- current lifecycle state;
- current accepted orientation;
- policy-exception usage/validity where needed.

#### 4. Disposition workflow

Support:

- accept;
- accept with overrides;
- reject;
- defer.

Only accepted dispositions update the current accepted-orientation projection.

Overrides must record rationale and must not bypass a non-exceptionable policy.

#### 5. Comparison and replay

Support:

- compare two orientation runs;
- explain score/input/policy/placement changes;
- deterministic replay where retained inputs permit it;
- explicit `partially_reproducible` / `not_reproducible` status when required evidence is unavailable.

### Failure/recovery gate

Tests must cover:

- failure before authoritative commit;
- response loss after successful commit and idempotent retry;
- projection failure after authoritative commit;
- projection rebuild;
- optimistic-concurrency conflict;
- interrupted backup/export;
- restore into a staged location before activation;
- invalid/corrupt restore;
- redacted historical evidence affecting replay.

### Security/privacy gate

Verify with an automated integration test or equivalent harness that the core UC-01 workflow:

- performs no outbound network access;
- requires no hosted account;
- stores portfolio data outside the source repository by default;
- does not persist credentials/secrets in domain records or traces;
- treats restored files and imported local inputs as untrusted until validated.

### Demo

From an empty local data directory:

1. register ten existing projects using the explicit approved bootstrap path;
2. confirm each bootstrap decision records skipped intake context and maintainer rationale;
3. enter evaluations;
4. run orientation;
5. inspect reasons;
6. accept with at least one legal override;
7. re-run after changing one evaluation;
8. compare runs;
9. rebuild projections;
10. export/backup and restore;
11. repeat with network access disabled.

### Exit criteria

This phase is the **v0 MVP release candidate** when all PRD MVP criteria are met. Phase 3 determines whether that candidate is practical and hardened enough for the v0 dogfood release.

## Phase 3 — Hardening and dogfood release

### Goal

Make the local vertical slice safe and practical enough for repeated real portfolio use.

### Required work

#### Diagnostics and operability

- structured diagnostics;
- `calathea doctor` command;
- explicit data-location/schema/projection health checks;
- actionable recovery guidance without exposing sensitive record payloads.

#### Data/recovery hardening

- migration testing across persisted schema versions;
- property/invariant tests for history immutability and queue capacity;
- expanded replay/rebuild/backup/restore failure cases;
- redaction/tombstone behavior exercised end to end.

#### Security hardening

- threat model for local data, untrusted imports, backup files, restore inputs, and CLI arguments;
- automated no-network and secret-handling regression tests;
- least-disclosure diagnostic/logging review.

#### Release engineering and documentation

- release packaging appropriate to the selected runtime;
- CI expanded to formatting, static analysis, unit, golden, integration, migration, and security checks;
- documentation for installation, initialization, bootstrap, backup, restore, redaction, and recovery;
- performance baseline for portfolios larger than the minimum ten-project target.

#### Dogfood

- deterministic fixture corpus expanded from real dogfood scenarios;
- baseline measurement for recurring orientation time and usability;
- anonymized/synthetic fixtures for any discovered semantic edge cases.

### Demo

Use Calathea for a real recurring project-orientation session and preserve an anonymized or synthetic equivalent fixture reproducing any discovered semantic edge cases.

### Exit criteria

- no known path silently mutates historical records;
- no known path treats AI/imported data as canonical authority;
- offline UC-01 remains fully functional;
- backup/restore procedure has been exercised successfully;
- diagnostics identify deliberately broken projections/data-version mismatches safely;
- documented dogfood run is explainable end to end.

Passing this phase defines the **v0 dogfood release**.

## Phase 4 — Optional read-only external source adapter

**Post-MVP unless dogfooding demonstrates it is essential.**

Introduce at most one adapter, likely GitHub, for scoped read-only signals.

Requirements:

- explicit source scope;
- least-privilege credential use;
- provenance and source revision retained;
- imported content treated as data, never instructions;
- partial import semantics;
- prior imported history preserved;
- no repository writes;
- deterministic core remains usable with the adapter removed.

The adapter must solve a measured workflow problem rather than merely demonstrate integration capability.

## Phase 5 — Review and drift workflow

**Post-MVP.**

Implement UC-03 using RFC 0003 and RFC 0008:

- scoped evidence collection;
- observations;
- findings;
- recommendations;
- valid no-change review outcomes;
- indeterminate reviews;
- explicit maintainer dispositions;
- outcome/calibration signals.

Automatic heuristic learning remains deferred.

## Explicitly deferred

Until separate evidence justifies them:

- AI provider integration;
- Anthesis integration;
- effect-governance adapters;
- repository or issue writes;
- autonomous agents;
- automatic heuristic/confidence/policy learning;
- multi-user evaluation aggregation;
- real-time collaboration;
- hosted/cloud persistence;
- web application;
- generic plugin framework or marketplace;
- bidirectional synchronization;
- reminders/notifications;
- organization-wide portfolio management.

## Implementation work decomposition

Reusable implementation should be tracked as independently reviewable public issues in this dependency order:

1. runtime/language decision, implementation skeleton, local quality command, and minimal CI;
2. foundational domain types, lifecycle/bootstrap contracts, and invariants;
3. evaluation semantics and golden fixtures;
4. policy subset and exception semantics;
5. deterministic orientation engine and structured traces;
6. physical local persistence and migration ADR;
7. immutable record persistence and identity/idempotency semantics;
8. rebuildable projections and optimistic-concurrency behavior;
9. project/bootstrap, evaluation, and policy CLI commands;
10. orientation and disposition CLI workflow;
11. comparison and deterministic replay;
12. projection rebuild and authoritative-command recovery paths;
13. export/backup/restore with staged validation;
14. redaction/tombstone behavior and secret-safe local data handling;
15. no-network, corruption, concurrency, migration, and failure-recovery integration tests;
16. threat model, diagnostics/`doctor`, and operational documentation;
17. release packaging, expanded CI gates, performance baseline, and dogfood documentation/fixtures.

Each implementation issue must cite the specific RFC/ADR/use-case acceptance criteria it satisfies. Work 1–5 defines Phase 1; 6–15 defines the Phase 2 release candidate; 16–17 completes Phase 3 hardening and dogfood release.

## Release definition

Calathea v0 is **not** defined by implementing every RFC capability.

It is defined by successfully delivering UC-01 as a privacy-preserving, local-first, deterministic, explainable, human-approved workflow with immutable history and tested recovery semantics.
