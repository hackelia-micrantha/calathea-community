# RFC 0005 — State, History, and Source-of-Truth Semantics

## Status

Accepted

- **Owner:** Calathea maintainers
- **Decision date:** 2026-08-03
- **Depends on:** RFC 0000, PRD, UC-01
- **Scope:** Authority, mutation, history, conflict, retention, and recovery semantics

## Summary

Calathea distinguishes authoritative records from imported facts, recommendations, projections, and external state. Durable decisions are append-only and attributable. Corrections create superseding records rather than rewriting history. Current views may be stored for efficiency, but they are rebuildable from authoritative records.

This RFC does not require event sourcing, a particular database, or a synchronization framework.

## Goals

- Define which system is authoritative for each record category.
- Define when records may be created, superseded, archived, redacted, or deleted.
- Preserve deterministic replay and decision auditability.
- Make imports, retries, conflicts, and partial failures deterministic.
- Keep private data local by default and make recovery practical.

## Non-goals

- Selecting a database, serialization format, or transaction engine.
- Defining project lifecycle states.
- Defining policy composition or evidence schemas in full.
- Supporting bidirectional synchronization in v0.
- Requiring an event log as the physical persistence model.

## Terminology

### Authority

The actor or system permitted to establish domain truth for a record.

### Source of truth

The authoritative system for a category of data. A copied record does not transfer authority unless an explicit maintainer decision creates a new Calathea-owned record.

### Authoritative record

A durable record whose interpretation governs Calathea behavior. Authoritative records are either maintainer-authorized Calathea records or externally authoritative references.

### Supersession

Creation of a new record that replaces an older record for current interpretation while preserving the older record historically.

### Projection

A rebuildable current or summary view derived from authoritative records.

### Snapshot

An immutable record of the inputs and outputs of a bounded operation. An orientation run is a snapshot in this sense, but `snapshot` is not the preferred domain name.

### Tombstone

A durable marker that an identity or record is unavailable, withdrawn, redacted, or deleted according to policy without silently reusing its identity.

## Authority model

| Record category | Source of truth | Calathea authority |
| --- | --- | --- |
| Portfolio identity and authored metadata | Calathea | Maintainer-authorized |
| Project identity and authored versions | Calathea | Maintainer-authorized |
| Evaluation versions | Calathea | Maintainer-authorized |
| Policy-set versions | Calathea | Maintainer-authorized |
| Orientation runs and diagnostics | Calathea deterministic core | Derived, immutable |
| Orientation dispositions and overrides | Calathea | Maintainer-authorized |
| Current accepted orientation | Calathea | Rebuildable projection |
| Review findings and dispositions | Calathea | Derived and maintainer-authorized components |
| Lifecycle decisions | Calathea | Maintainer-authorized |
| Repository, issue, PR, and CI state | External system | Read-only reference/import in v0 |
| AI output | AI provider/model | Untrusted recommendation draft only |
| Credentials and secret values | Secret-management facility | Never Calathea domain state |

## Record dimensions

RFC 0000 defines orthogonal dimensions. This RFC applies them as follows.

### Authority dimension

- **Maintainer-authorized:** may govern Calathea behavior.
- **Deterministically derived:** authoritative only as the result of specified inputs and semantics; never self-approved.
- **Externally authoritative:** owned by another system.
- **Non-authoritative:** draft, cache, working state, or advisory output.

### Derivation dimension

- **Authored:** directly supplied or approved by the maintainer.
- **Imported:** copied from an external source with provenance.
- **Observed:** normalized or inferred from authored/imported data.
- **Derived:** recomputable from retained inputs and semantic versions.
- **Projected:** rebuildable current or summary view.

### Durability dimension

- **Ephemeral:** may be discarded without changing durable truth.
- **Durable mutable pointer:** a replaceable convenience reference whose target is authoritative elsewhere.
- **Durable immutable:** never edited after successful creation.
- **Archived:** retained but excluded from normal active views.
- **Redacted:** content removed or masked while preserving required structural evidence.
- **Tombstoned:** identity retained while content is unavailable or deleted.

## Entity classification

| Entity | Authority | Derivation | Durability |
| --- | --- | --- | --- |
| Portfolio | Maintainer-authorized | Authored | Stable identity; versioned content |
| Project | Maintainer-authorized | Authored | Stable identity; versioned content |
| ProjectVersion | Maintainer-authorized | Authored | Immutable after acceptance |
| EvaluationVersion | Maintainer-authorized | Authored or promoted draft | Immutable after acceptance |
| PolicySetVersion | Maintainer-authorized | Authored | Immutable after activation/use |
| OrientationRun | Deterministically derived | Derived | Immutable |
| PlacementRecommendation | Deterministically derived | Derived | Immutable within run |
| EligibilityDiagnostic | Deterministically derived | Derived | Immutable within run |
| DecisionTraceEntry | Deterministically derived | Derived | Immutable within run |
| OrientationDisposition | Maintainer-authorized | Authored | Immutable |
| PlacementOverride | Maintainer-authorized | Authored | Immutable within disposition |
| AcceptedOrientation | Deterministically interpreted | Derived | Historical interpretation |
| CurrentAcceptedOrientation | Calathea projection | Projected | Rebuildable |
| Observation | Non-authoritative until used by decision | Imported/observed | Immutable once recorded |
| Finding | Derived recommendation | Derived | Immutable once recorded |
| ReviewDisposition | Maintainer-authorized | Authored | Immutable |
| LifecycleDecision | Maintainer-authorized | Authored | Immutable |
| EvidenceReference | Source-dependent | Authored/imported | Versioned or immutable reference |
| SourceReference | External-authoritative reference | Imported | Stable reference with availability state |
| AIInvocation | Non-authoritative provenance | Imported/generated | Immutable invocation record |
| RecommendationDraft | Non-authoritative | Generated/authored | Ephemeral or retained immutable draft |

## Creation and mutation rules

### Stable identities

Stable identities are never reused for a different logical entity, including after archival or deletion.

### Versioned authored content

Changing project, evaluation, or policy content creates a new immutable version. A current-version pointer may be updated atomically, but the prior version remains historical.

### Immutable operational records

Orientation runs, dispositions, lifecycle decisions, recorded review results, and AI invocation provenance are immutable after durable creation.

### Corrections

A correction must:

1. create a new record;
2. reference the corrected or superseded record;
3. identify actor, authority, time, and reason;
4. preserve the original record unless retention law or explicit privacy policy requires redaction/deletion;
5. rebuild affected projections deterministically.

A typo in an uncommitted draft may be edited. Once a record has influenced a durable decision or is declared accepted, corrections use supersession.

## Orientation semantics

### Orientation runs

An orientation run is an immutable deterministic record of:

- complete versioned inputs or immutable references;
- semantic versions;
- planning horizon;
- recommendations, diagnostics, and trace;
- operation identity and timestamps.

The run is never mutated by acceptance, override, rejection, or deferral.

### Dispositions

A disposition is a separate immutable maintainer decision referencing one orientation run.

Allowed kinds:

- `accepted`;
- `accepted_with_overrides`;
- `rejected`;
- `deferred`.

Rejected and deferred dispositions do not change the current accepted orientation.

### Effective accepted orientation

An accepted orientation is derived from:

- one accepted or accepted-with-overrides disposition;
- the referenced orientation run;
- validated overrides and policy exceptions.

The current accepted orientation is selected by deterministic supersession rules for a portfolio and planning scope. It is a projection, not the source of truth.

## Concurrency and optimistic locking

Commands that amend canonical authored state must include an expected version, revision token, or equivalent precondition.

On mismatch:

- Calathea rejects the write as a conflict;
- it does not silently merge semantic fields;
- it returns the expected and current references plus enough context for the user to retry;
- the caller must reload, reconcile, and submit a new explicit decision.

Pure append operations may use unique operation identities and idempotency keys instead of entity-version preconditions.

## Idempotency

Every externally retryable write operation must have a stable idempotency identity within a documented scope.

Repeated execution with the same identity and equivalent payload must return the original result or an equivalent success reference.

Reuse of the same identity with a materially different payload must fail as an idempotency conflict.

At minimum, idempotency applies to:

- project registration;
- evaluation acceptance;
- policy activation;
- orientation-run creation requests;
- orientation dispositions;
- imports;
- backup restore operations.

Generated IDs alone are insufficient when a caller can retry after losing the response.

## Deterministic replay

A deterministic operation is replayable only when Calathea retains or can resolve:

- exact input record versions;
- semantic and schema versions;
- policy-set version;
- planning-horizon value;
- deterministic ordering and tie-break rules;
- required imported observations or content identities;
- operation parameters.

Replay outcomes:

- **reproduced:** output identity/content matches expected deterministic result;
- **not reproducible:** a required input or semantic implementation is unavailable;
- **invalidated:** retained data fails integrity validation;
- **not applicable:** operation involved nondeterministic AI inference.

AI output may be re-invoked, but that is a new invocation rather than deterministic replay.

## Import and external-source semantics

### Read-only boundary

External integrations are read-only in v0. Import never grants Calathea authority over the external object.

### Import identity

An import record must retain:

- external source identity;
- external object identity;
- collection time;
- observed external revision or content identity where available;
- adapter and schema version;
- availability and freshness state;
- transformation provenance.

### Duplicate imports

Equivalent observations of the same external revision should deduplicate or link as repeated observations. They must not appear as independent corroborating evidence merely because they were imported twice.

### External changes

When an external object changes:

- create a new imported observation;
- preserve the prior observation historically if it influenced a decision;
- mark derived views stale where relevant;
- never rewrite a historical orientation run to use the newer data.

### External deletion or unavailability

Calathea records the source as unavailable, deleted, or inaccessible. It does not silently erase prior provenance or claim that referenced evidence still exists.

If copied evidence content is retained, retention policy governs that copy independently of external deletion.

### Conflicts

Imported state cannot conflict with Calathea-authored truth in the same sense as a bidirectional synchronization system because it is not promoted automatically.

When imported data disagrees with authored data:

- preserve both;
- classify the disagreement as an observation or diagnostic;
- require an explicit maintainer action to amend canonical authored state;
- retain the decision rationale and source references.

## Atomicity and partial failures

### Required atomic boundaries

The architecture must provide all-or-nothing durability for each individual authoritative command, including:

- accepting a project/evaluation/policy version and advancing its current pointer;
- recording an orientation disposition and its overrides;
- recording a lifecycle decision and advancing the current lifecycle projection.

An orientation run may be persisted independently before a disposition.

### Failure before durable commit

The operation is not reported as successful. Retriable working output may be offered, but it has no authoritative status.

### Failure after durable commit but before response

The caller retries using the same idempotency identity and receives the original result.

### Projection failure

Failure to update a projection does not invalidate successfully committed authoritative records. The projection is marked unavailable or stale and rebuilt.

### Multi-step imports

Partial imported batches are either:

- isolated under a batch identity and marked incomplete; or
- rolled back if atomic batch behavior is supported.

Incomplete batches must not silently influence orientation.

## Integrity

Durable records should support integrity verification appropriate to their storage representation.

The logical requirement is detection of accidental corruption or incomplete restore, not a mandate for blockchain, Merkle trees, or cryptographic signing.

Where content digests are used:

- the digest algorithm and canonicalization version are recorded;
- digests are not treated as proof of authorship;
- sensitive data is not exposed through unsafe low-entropy hashes.

## Retention model

### Retention categories

- **Required decision history:** records necessary to interpret accepted decisions.
- **Reproducibility inputs:** records necessary for deterministic replay.
- **Operational history:** useful diagnostics that may have bounded retention.
- **Imported evidence:** subject to source, privacy, and user-configured retention.
- **AI prompts/outputs:** minimized and not retained by default beyond required provenance.
- **Ephemeral working data:** deleted after completion or failure recovery window.

### Minimum v0 behavior

Calathea must document default retention behavior and allow the user to inspect what is retained.

It must not silently purge a record required to interpret a current accepted orientation.

Retention policy changes apply prospectively unless an explicit migration operation reports affected history and replay capability.

## Archival

Archival removes records from normal active workflows without destroying identity or history.

Archived projects:

- retain stable identity and history;
- are excluded from active orientation by explicit eligibility semantics;
- may be restored through an explicit maintainer action;
- are not equivalent to `kill` placement recommendations.

## Redaction

Redaction removes or masks sensitive content while preserving the minimum structural facts required for integrity and audit.

A redaction record must identify:

- target record and fields or content class;
- actor and authority;
- reason and time;
- policy/legal basis where applicable;
- impact on replay, explanation, and evidence availability.

After redaction, Calathea must not claim full replay or evidentiary completeness if required content is unavailable.

References to redacted content return an explicit redacted state rather than null or not-found ambiguity.

## Deletion

Deletion is distinct from archival and redaction.

### Allowed deletion

- ephemeral drafts and caches;
- imported copies not used by retained decisions, subject to policy;
- AI raw content beyond required provenance;
- user-owned content when explicit deletion requirements outweigh audit retention.

### Restricted deletion

Records that support a current or retained accepted decision should not be physically deleted by default. Prefer archival or redaction.

When physical deletion is required:

- retain a tombstone and deletion metadata where legally and operationally permitted;
- mark dependent explanations and replay as incomplete;
- rebuild projections;
- avoid identity reuse.

## Export

A complete export must distinguish:

- authoritative records;
- projections/caches;
- imported/external references;
- redacted or tombstoned content;
- semantic/schema version dependencies;
- integrity metadata.

Portable export must not include plaintext credentials.

An export may omit rebuildable projections if the restore process can regenerate them.

## Backup

Backups are user-controlled and encrypted by default in guidance.

A backup must include enough information to restore:

- stable identities;
- authoritative records and supersession links;
- semantic/schema references;
- required imported evidence copies or explicit missing-reference state;
- redaction/tombstone metadata;
- integrity metadata.

Calathea must not claim a backup succeeded until the backup is durably written and minimally verified.

## Restore

Restore is an explicit operation with a restore identity.

Required phases:

1. validate format and integrity;
2. identify semantic/schema compatibility;
3. stage without altering current authoritative state;
4. detect identity and history conflicts;
5. produce a restore plan;
6. require explicit maintainer confirmation for destructive or replacing actions;
7. commit authoritative records;
8. rebuild projections;
9. verify key invariants and report omissions.

A failed restore must leave the prior installation usable unless the user explicitly selected a destructive recovery mode.

## Restore conflict behavior

When restoring into non-empty state:

- identical record identities and content deduplicate;
- identical identities with different immutable content are corruption or history conflicts and must fail;
- new superseding records may be imported if causal links are valid;
- current pointers/projections are recomputed rather than accepted blindly;
- external references remain references and are revalidated separately.

## Recovery and rebuild

The following projections must be rebuildable from authoritative records:

- current project version;
- current policy-set version;
- current accepted orientation;
- current lifecycle state once lifecycle semantics exist;
- active/archive eligibility views;
- comparison indexes and summaries.

A rebuild operation:

- does not create new domain decisions;
- records operation identity and diagnostics;
- validates causal links and uniqueness constraints;
- reports orphaned or unavailable dependencies;
- can run without network access, except optional revalidation of external references.

## Privacy requirements

- Core data remains in a user-controlled local location.
- No automatic cloud synchronization or telemetry is required.
- Credentials remain outside domain records and exports.
- Outbound AI/integration data is explicitly scoped and minimized.
- Backup guidance defaults to encrypted user-controlled storage.
- Redaction and deletion limitations are visible before execution.

## Security considerations

Threats include:

- tampered local history;
- rollback to an older backup;
- replay of stale commands;
- idempotency-key collision or malicious reuse;
- imported evidence substitution;
- secret leakage through exports or retained AI context;
- projection corruption masquerading as authoritative truth.

Required mitigations include:

- immutable record identity and causal references;
- optimistic concurrency for canonical amendments;
- scoped idempotency identities;
- provenance and external revision capture;
- integrity verification during restore/rebuild;
- clear separation of projections from authoritative records;
- secret exclusion and redaction controls.

## Storage-neutral implementation constraints

An implementation may use mutable tables, append-only records, files, a relational database, or another local persistence mechanism if it preserves these semantics.

It must not expose physical storage behavior as domain truth. In particular:

- append-only domain history does not require event sourcing;
- a current-row representation does not permit historical mutation;
- database transactions do not replace actor, authority, and decision records;
- cached projections do not become authoritative merely because they are persisted.

## Consequences

### Benefits

- Clear authority and source-of-truth boundaries.
- Rejected and superseded decisions remain inspectable.
- Imports cannot silently overwrite authored state.
- Retries and recovery have deterministic outcomes.
- Storage technology remains replaceable.

### Costs

- More explicit identities, versions, and causal links.
- Retention and redaction require dependency analysis.
- Restore must validate history rather than copy files blindly.
- Projection rebuild paths must be tested.

## Follow-up requirements

- RFC 0006 defines lifecycle states and legal transitions.
- RFC 0007 defines policy exceptions and override legality.
- RFC 0008 defines evidence and explanation availability/redaction semantics.
- ADR 0002 and the runtime architecture define the persistence boundary consistent with this RFC; physical storage selection remains a later implementation ADR.
- The MVP roadmap requires backup, restore, replay, and projection-rebuild acceptance tests.

## Acceptance criteria

This RFC is accepted because:

- every RFC 0000 entity has authority, derivation, and durability classification;
- orientation runs and dispositions are immutable and separate;
- corrections use supersession rather than silent mutation;
- optimistic concurrency and idempotency are defined;
- imports preserve external authority and provenance;
- external changes never rewrite historical decisions;
- partial writes and projection failures have deterministic recovery;
- retention, archival, redaction, deletion, backup, and restore semantics are explicit;
- required audit history is not silently invalidated;
- the design remains storage-technology neutral.
