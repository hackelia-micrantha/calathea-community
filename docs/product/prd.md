# Calathea Product Requirements Document

## Status

Accepted for v0 planning

- **Owner:** Calathea maintainers
- **Decision date:** 2026-08-03
- **Scope:** Initial product boundary and release requirements
- **Selected v0 use case:** [UC-01 — Orient a portfolio of existing projects](use-cases.md#uc-01-orient-a-portfolio-of-existing-projects)

## Product summary

Calathea is a privacy-preserving, local-first project portfolio orientation system for maintainers who need to decide what to work on now, what should come next, what can wait, and what should be stopped.

It converts explicit project evaluations and policy constraints into deterministic, explainable `now`, `next`, `later`, and `kill` recommendations. Users remain authoritative: Calathea may observe, evaluate, derive, compare, explain, and recommend, but it does not silently mutate canonical project state or execute repository changes.

`kill` is an orientation recommendation to stop investing in a project. It is not a project lifecycle transition, deletion, archival action, or external-system mutation. Any resulting lifecycle change requires a separate explicit maintainer decision.

Calathea may later support broader project operating workflows, but v0 is deliberately narrow. Its first responsibility is to make portfolio-level prioritization inspectable, reproducible, private, and governable.

## Problem statement

Technical maintainers often manage more projects and plausible workstreams than available capacity permits. Intent, risk, effort, urgency, and status are distributed across repositories, issues, documents, and conversations. Existing backlogs preserve work but rarely provide a trustworthy answer to:

- What deserves immediate focus?
- Why was one project selected over another?
- Which constraints affected the decision?
- What changed since the previous review?
- Which projects should be explicitly stopped rather than deferred indefinitely?

AI-assisted planning can reduce synthesis effort, but it creates additional risk when recommendations are opaque, inputs are untraceable, private data leaves user control without clear intent, or generated output mutates state without review.

Calathea addresses this with explicit evaluation inputs, versioned policy effects, deterministic recommendation traces, maintainer dispositions, and immutable history.

## Primary user

The primary v0 user is a single technical maintainer responsible for a portfolio of software projects or substantial workstreams.

Typical characteristics:

- owns or stewards several repositories or long-lived initiatives;
- has more plausible work than available execution capacity;
- values explicit trade-offs, privacy, auditability, and repeatable planning;
- may use AI for analysis but does not want AI to become silently authoritative;
- already uses GitHub or files as execution and documentation systems of record.

## Secondary and excluded users

Secondary future users may include small engineering teams, technical program leads, open-source maintainers, and governance-focused organizations.

The following are not initial target users:

- large organizations requiring real-time multi-user portfolio collaboration;
- teams seeking a complete replacement for Jira, Linear, or GitHub Issues;
- users seeking autonomous software-delivery agents;
- organizations requiring a hosted enterprise control plane in v0;
- non-technical general-purpose task-management users.

## Primary job to be done

When a maintainer has multiple active or potential projects competing for limited capacity, they need to produce and review an explainable portfolio orientation so they can commit to a small active set without losing rationale, history, privacy, or control.

## Product boundary

### Calathea owns

- project registration and portfolio identity;
- structured project evaluation inputs and rationale;
- versioned evaluation and scoring semantics;
- policy-aware orientation recommendations;
- explanation of inclusion, exclusion, penalties, and tie-breaking;
- immutable orientation runs;
- explicit maintainer dispositions: accepted, accepted-with-overrides, rejected, or deferred;
- a rebuildable current accepted-orientation view;
- comparison between runs and accepted decisions;
- review findings based on scoped project and repository signals;
- domain-specific AI context assembly and validation of AI-produced recommendations.

The exact scoring formula, calibration rubrics, policy representation, lifecycle semantics, and evidence schema are owned by their corresponding RFCs rather than fixed by this PRD.

### External systems own

- source-code repositories and their content;
- GitHub issues, pull requests, checks, and execution work unless explicitly imported;
- credential issuance and secret values;
- deployment and runtime infrastructure;
- authoritative external project state when Calathea is a read-only consumer.

Calathea may retain scoped credential references required by an integration, but plaintext credentials must not be stored in project records, prompts, evidence, traces, or model context.

### Anthesis boundary

Calathea owns project-orientation semantics and domain-specific recommendation workflows.

When integrated, Anthesis owns effect governance, including:

- actor and capability authorization;
- least-privilege tool grants;
- approval requirements for effectful operations;
- effect attribution and evidence;
- bypass resistance and enforcement.

Calathea must not require Anthesis for deterministic read-only v0 workflows. Without Anthesis, Calathea still prohibits autonomous repository mutation and requires explicit local maintainer authority for canonical changes.

## Initial deployment and interaction model

Calathea v0 is local-first and single-user.

The initial interaction surface is a CLI backed by a stable application boundary suitable for later API or web adapters. Runtime and module choices are documented in the accepted ADRs.

Core portfolio data is stored in a user-controlled local location, separate from the Calathea source repository by default. Exporting private data into a repository is explicit user action.

The deterministic core must operate without network access. Hosted persistence, automatic cloud synchronization, mandatory telemetry, and automatic cloud backup are out of scope for v0.

Persisted records must support versioning, provenance, immutable historical runs and decisions, and deterministic replay.

## Data privacy and residency

Portfolio metadata, evaluations, rationales, outcomes, and evidence may contain sensitive personal or professional information. v0 therefore defaults to:

- no automatic project-data upload;
- no required telemetry;
- opt-in, explicitly configured external integrations and AI providers;
- minimized outbound data limited to the requested operation;
- visible destination, data categories, and provider-retention assumptions;
- transport credentials resolved out-of-band and excluded from project/audit/model content;
- configured secret patterns rejected or redacted from durable and outbound data;
- other sensitive portfolio data remaining local unless explicitly selected;
- no automatic commit of private portfolio data to the source repository;
- documented backup, restore, export, retention, redaction, and deletion behavior.

At-rest protection may rely on documented operating-system facilities in v0; Calathea must not claim built-in encryption it does not provide. Backup guidance defaults to encrypted, user-controlled storage. No cloud provider is the authoritative storage or trust center for v0.

## Systems of record

Calathea is authoritative for:

- registered portfolio/project metadata authored in Calathea;
- accepted evaluations and their versions;
- policy configuration used by an orientation run;
- immutable orientation runs;
- maintainer dispositions and overrides;
- the current accepted-orientation view derived from accepted decisions;
- Calathea-authored review findings and user-recorded outcomes.

External systems remain authoritative for imported repository, issue, pull-request, and CI state.

Imported or observed data must retain source identity, collection time, and provenance. Calathea must not silently rewrite external authoritative state.

## Selected v0 workflow

The authoritative workflow is [UC-01](use-cases.md#uc-01-orient-a-portfolio-of-existing-projects):

1. The maintainer registers a small portfolio of projects.
2. The maintainer records or reviews structured evaluations.
3. Calathea validates inputs and calculates versioned deterministic scores.
4. Calathea applies configured orientation policies.
5. Calathea records a proposed `now`, `next`, `later`, and `kill` orientation run.
6. The maintainer reviews explanations, evidence, confidence, exclusions, and policy effects.
7. The maintainer accepts, overrides, rejects, or defers the proposal explicitly.
8. Calathea records an immutable disposition referencing the run.
9. Only an accepted disposition updates the current accepted-orientation view.
10. Later runs can be compared with prior accepted decisions.
11. Any lifecycle or external-system change remains a separate authorized workflow.

## Functional requirements

### FR-1: Project portfolio registration

The maintainer must be able to register projects with stable identity and basic metadata without repository integration.

### FR-2: Structured evaluation

Each active candidate must support explicit values for impact, effort, risk reduction, optionality, urgency, confidence, rationale, and evaluation time.

### FR-3: Validation and calibration

Calathea must reject malformed evaluation data and provide scoring rubrics and calibration examples sufficient for consistent use.

### FR-4: Deterministic orientation

Identical persisted inputs and semantic versions must produce identical orientation output and explanation traces.

### FR-5: Bounded queues

The orientation engine must support bounded `now` and `next` sets. Effective limits must be visible in every run.

### FR-6: Policy-aware recommendation

Hard constraints, soft preferences, precedence, exceptions, and indeterminate outcomes must be explicit and inspectable. Non-overridable hard constraints cannot be bypassed silently.

### FR-7: Explanation

Each placement and material exclusion must include a concise reason and enough detail to inspect score, confidence, freshness, policy effects, and tie-breaking.

### FR-8: Human authority and disposition

Recommendations must not become accepted canonical decisions without explicit maintainer action. A run may be accepted, accepted with valid overrides, rejected, or deferred. Overrides and exceptions must capture actor, rationale, time, and referenced run.

### FR-9: Historical integrity

Orientation runs and dispositions must be immutable historical records. Corrections create superseding records rather than silently rewriting history.

### FR-10: Comparison

The maintainer must be able to compare runs and decisions and identify changes in inputs, semantic versions, scores, policies, placements, and dispositions.

### FR-11: Optional read-only external signals

The first external integration, if present, must be read-only and scoped. Imported signals must be attributable and remain distinct from canonical Calathea state.

### FR-12: Optional AI assistance

AI is not required for the MVP. If present, it may assist with drafts, summaries, and candidate findings. AI output is untrusted recommendation data, validated against versioned contracts and clearly distinguished from deterministic output and maintainer decisions.

### FR-13: Private local persistence

The core workflow must persist private data in a user-controlled local location without requiring a source repository, hosted account, telemetry endpoint, AI provider, or external project integration.

### FR-14: Export and recovery

The maintainer must have documented mechanisms to back up, restore, and export Calathea-owned records without weakening provenance or historical integrity.

## Non-functional requirements

### Correctness

- deterministic behavior for deterministic inputs;
- explicit invariants and validation failures;
- no silent policy relaxation;
- stable tie-breaking;
- versioned evaluation, scoring, orientation, and policy semantics;
- recommendation placement and project lifecycle remain separate;
- rejected/deferred runs do not replace current accepted orientation.

### Security and privacy

- least-privilege access to external systems;
- no credentials in prompts, traces, evidence, or project records;
- untrusted repository content cannot become executable instruction;
- sensitive data supports explicit redaction;
- deterministic core requires no network egress;
- optional outbound transfers are visible, scoped, and user-initiated;
- effectful integrations require an explicit governance boundary.

### Auditability

- runs, dispositions, overrides, exceptions, and corrections are attributable;
- replay inputs and semantic versions are retained or referenced immutably;
- historical records are never silently mutable;
- replaceable current views are rebuildable from immutable records.

### Reliability

- failed imports or AI calls do not corrupt canonical state;
- partial operations fail visibly and support safe recovery;
- repeated run requests and imports are idempotent or explicitly distinct;
- backup and restore preserve record identity, history, and causal links.

### Maintainability

- domain rules remain independent of CLI, storage, AI provider, and adapter choices;
- interfaces are explicit and testable;
- v0 avoids a generic policy DSL or plugin platform.

### Portability

The deterministic core must run locally without mandatory hosted services, network access, or AI-provider access.

## Non-goals for v0

- replacing GitHub Issues, Jira, Linear, or repository workflows;
- autonomous code or repository mutation;
- automated implementation agents;
- automatic learning that mutates scoring heuristics;
- real-time team collaboration;
- multi-user evaluation aggregation;
- generic plugin marketplace;
- broad notification/reminder system;
- bidirectional synchronization with external project tools;
- hosted or cloud persistence;
- mandatory telemetry or cloud backup;
- hosted SaaS control plane;
- generalized organizational portfolio management;
- prescribing a database, event-sourcing architecture, or deployment platform prematurely.

## Success measures

### Release acceptance measures

The primary user can:

1. register and evaluate at least ten projects;
2. generate an orientation with at most three `now` items;
3. understand the principal reason for every placement and material exclusion without reading implementation code;
4. accept, override, reject, or defer a run explicitly;
5. reproduce accepted recommendations exactly from recorded deterministic inputs and semantic versions;
6. compare two runs and explain why placements or dispositions changed;
7. complete UC-01 with network access disabled;
8. complete UC-01 without AI or external repository access;
9. keep private portfolio data outside the source repository by default;
10. use optional AI assistance without granting canonical mutation authority.

Required integrity measures:

- 100% of dispositions reference an immutable orientation run;
- 100% of material policy effects appear in the decision trace;
- identical versioned inputs produce identical deterministic output;
- core operation produces no outbound network request when integrations are disabled;
- v0 contains no repository-write capability;
- rejected/deferred runs never replace current accepted orientation.

### Product outcome measures

Initial dogfooding targets, revisable only with recorded baseline evidence:

- review ten current projects in 15 minutes;
- reduce recurring orientation effort by at least 25% versus the documented manual baseline;
- explain or intentionally override every `now` recommendation;
- make stop-investing decisions explicit rather than indefinitely deferred.

## MVP release criteria

The initial release is ready when:

- [UC-01](use-cases.md#uc-01-orient-a-portfolio-of-existing-projects) is implemented end to end;
- only the minimum UC-02 registration subset needed by UC-01 is included;
- the conceptual domain model and state/source-of-truth semantics are accepted;
- evaluation and orientation semantics are reconciled and covered by golden tests;
- a local CLI can register projects, capture evaluations, run orientation, explain output, and record all supported dispositions;
- orientation runs and dispositions are immutable and comparable;
- current accepted-orientation views are rebuildable;
- deterministic workflows operate with network access disabled;
- private project data is stored outside the source repository by default;
- backup, restore, export, and recovery are documented and validated;
- AI, if present, is optional, validated, and non-authoritative;
- external integrations, if present, are optional, read-only, scoped, and attributable;
- security assumptions and the Calathea–Anthesis boundary are documented;
- no known path permits silent canonical mutation by AI or imported data;
- no lifecycle transition occurs merely because an item receives a `kill` recommendation.

## Risks and mitigations

### Product scope expands into general project management

Require every v0 capability to support UC-01 directly.

### Scoring appears objective while encoding subjective judgment

Expose rationale, confidence, calibration guidance, policy effects, exceptions, and overrides. Do not present confidence as calibrated probability unless validated.

### Excessive planning delays validation

Accept only foundational contracts required for the vertical slice and defer implementation-specific RFCs.

### Governance duplicates Anthesis

Keep Calathea responsible for domain orchestration and recommendations; delegate authorization and effect governance where Anthesis is integrated.

### Repository content manipulates AI recommendations

Treat imported content as untrusted data, constrain context assembly, validate output, and retain deterministic fallback behavior.

### Private portfolio data leaves user control

Use local user-controlled storage, no mandatory telemetry/cloud service, explicit network egress, scoped integrations, and data minimization.

### Local history is corrupted or silently rewritten

Use immutable runs/decisions, content identity where appropriate, append-only correction, documented backup/restore, and rebuildable views.

### `kill` is interpreted as destructive mutation

Model orientation and lifecycle independently, require explicit lifecycle authority, and preserve history.

## Decision ownership

The previously open foundational decisions are now captured by accepted contracts:

- conceptual entities and terminology: RFC 0000;
- canonical/imported/derived/history/retention semantics: RFC 0005;
- project lifecycle states and legal transitions: RFC 0006;
- policy representation, exceptions, and composition: RFC 0007;
- evidence and explanation contracts: RFC 0008;
- AI orchestration and governance boundary: RFC 0004 plus the architecture integration contracts;
- implementation architecture and runtime: `docs/architecture/` and `docs/adr/`;
- phased implementation sequence: [MVP roadmap](mvp-roadmap.md).

## Decision summary

For v0, Calathea is a privacy-preserving, local-first, single-user, human-approved portfolio orientation product—not a general project-management replacement or autonomous execution platform.

Its central invariant is:

> Calathea may observe, evaluate, derive, explain, and recommend. Canonical decisions and effectful changes require explicit authority.
