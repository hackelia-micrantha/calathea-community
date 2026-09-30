# ADR 0006 — SQLite v0 Local Persistence

## Status

Proposed

- **Owner:** Calathea maintainers
- **Date:** 2026-08-13
- **Tracked by:** #14
- **Governing constraints:** ADR 0002; RFC 0005; RFC 0008; UC-01; MVP roadmap
- **Supersedes:** none
- **Superseded by:** none

## Context

Calathea v0 requires local, user-controlled persistence for sensitive portfolio state while preserving immutable decision history, deterministic replay inputs, explicit idempotency and optimistic-concurrency behavior, rebuildable projections, schema migration, backup/restore, redaction/tombstones, and offline operation.

ADR 0002 establishes those logical persistence semantics but deliberately does not choose a physical storage technology. RFC 0005 also explicitly avoids requiring event sourcing or a database.

The physical v0 implementation should minimize mechanism. A local CLI should not need to build its own transaction manager, index engine, multi-file recovery protocol, or hosted service merely to persist a bounded graph of immutable records and rebuildable projections.

The repository's public build contract also does not currently require CGO or a C compiler. A storage choice that silently adds that requirement would widen packaging and CI scope.

## Proposed decision

Use **one SQLite database per Calathea data root**, accessed only through application-owned persistence ports.

Prefer a **pinned CGO-free Go SQLite driver** if the #14 spike proves the candidate's build, dependency, feature, and upgrade posture acceptable. `modernc.org/sqlite` is the initial `database/sql` candidate because it provides a CGO-free SQLite implementation; the spike may select a different CGO-free driver if it is materially simpler or safer.

This ADR remains **Proposed** until the #14 validation spike demonstrates the required transaction, idempotency, concurrency, migration, projection rebuild, backup, restore, integrity, and packaging behaviors. Failure of that spike is grounds to reject or revise this ADR rather than forcing SQLite into the design.

### Storage boundary

The database contains physically distinct categories even if they share one file:

1. **authoritative immutable records/history**;
2. **operation/idempotency metadata** required for safe externally retryable writes;
3. **rebuildable projections/current views**;
4. **persisted schema/migration metadata**;
5. later redaction/tombstone/integrity metadata required by RFC 0005/#23.

Domain packages do not import SQL packages, know database paths, or expose SQLite row/schema concepts. Application-owned ports define persistence behavior; the SQLite adapter implements them.

### One database, not one file per aggregate

Use one database file for a Calathea data root rather than separate files/databases per record category. UC-01 authoritative commands can span related records, idempotency metadata, and required current references. One SQLite transaction/integrity boundary is simpler and safer than coordinating multiple files.

### Conservative transaction/concurrency model

Optimize v0 for correctness and a local CLI rather than high write concurrency:

- one writer transaction per authoritative command;
- a small/bounded connection pool;
- explicit handling of lock/contention failures;
- operation/idempotency identity remains the semantic retry mechanism;
- no unbounded transparent retry loop around authoritative commands;
- stale expected-version commands fail visibly rather than last-write-wins.

Start with SQLite's ordinary rollback-journal operating model. WAL is not a product requirement and should be enabled only if measured concurrency/performance evidence justifies its additional checkpoint/sidecar/recovery behavior.

Required SQLite settings are executable-owned. They must be tested/read back where practical rather than assumed from a successfully executed `PRAGMA`, because SQLite may ignore unknown pragmas without reporting an error.

Foreign-key enforcement must be enabled for every relevant connection before relying on database-level referential constraints.

### Authoritative atomic boundaries

When RFC 0005 requires a record plus a required current selection/reference to advance atomically, write them in one transaction together with the operation/idempotency record.

Other rebuildable projections may be updated after the authoritative commit according to #16/#21. Projection update failure does not permit rewriting or deleting successfully committed authoritative history.

### Idempotency

Every externally retryable authoritative command stores a stable operation/idempotency identity in the same authoritative transaction as its result.

The adapter/application contract must support:

- same identity + equivalent material request => return the original committed result/reference;
- same identity + materially different request => deterministic idempotency conflict;
- lost response after commit => retry resolves the original result;
- failure before commit => no authoritative partial result.

Material-request equivalence must use an explicit schema/canonicalization contract. Incidental Go struct formatting or ordinary JSON byte order is not a canonical fingerprint.

### Optimistic concurrency

Commands that require an expected current version/revision carry that precondition into the transaction. The transaction verifies it before creating the new authoritative record/current reference.

A mismatch returns deterministic expected/current references and does not silently merge semantic state.

SQLite locking is an implementation safety mechanism; it is not a substitute for the domain/application expected-version contract.

### Schema migration

Use explicit executable-embedded forward migrations with a Calathea-owned monotonically increasing persisted-schema version.

Requirements:

- schema compatibility version is distinct from SQLite's internal schema metadata and from domain semantic versions;
- migrations are ordered and attributable;
- migrations are transactional where SQLite permits;
- a database newer than the executable's supported schema fails closed;
- supported prior versions have automated migration tests;
- destructive migrations require explicit backup/recovery planning;
- migrations never silently reinterpret historical domain semantics.

Down migrations are not required for v0. Recovery to older application versions should rely on supported compatibility or an explicit backup/restore path rather than lossy reverse migration.

### Projection strategy

Projection/current-view tables are explicitly non-authoritative and rebuildable from authoritative records.

A rebuild operation may drop/clear and reconstruct projections without creating new domain decisions. Projection rows should carry enough source revision/version metadata to detect staleness where useful.

Indexes are implementation aids, not authoritative history.

### Integrity

Integrity is layered:

1. SQLite structural/index integrity;
2. database foreign-key integrity;
3. Calathea domain/history integrity.

Database-level validation should use appropriate SQLite integrity checks, but those are not sufficient on their own. `PRAGMA integrity_check` checks structural/index/constraint consistency but does not report foreign-key violations, so `foreign_key_check` (or equivalent) is required separately. Calathea must additionally validate supersession/causal links, operation identity, semantic-version support, tombstones/redactions, and other domain invariants.

The integrity goal is detection of accidental corruption, incomplete migration/restore, and semantically invalid history. This ADR does not claim tamper-proof history against a process/user with permission to modify the database file.

### Backup feasibility

Do not copy a live SQLite database file naively.

The #14/#22 implementation should select one supported consistent-snapshot mechanism exposed by the chosen driver, preferably:

- SQLite's online backup API; or
- `VACUUM INTO` for the small local v0 database when its simpler whole-snapshot behavior is sufficient.

A backup is not reported successful until the output is durably completed and minimally verified. The backup candidate must be reopened separately and pass persisted-schema, SQLite integrity, foreign-key, and Calathea domain/history validation.

`VACUUM INTO` is attractive for a compact snapshot and removal of unused/deleted pages from the generated copy, but it does not itself replace Calathea's backup verification, encryption guidance, retention, or failure protocol.

### Restore feasibility

Restore never writes an unvalidated candidate directly into current authoritative state.

The physical storage model must support:

1. stage candidate separately;
2. validate SQLite format/integrity;
3. validate foreign keys;
4. validate Calathea persisted-schema/semantic compatibility;
5. validate record/history identity invariants;
6. detect conflicts and produce a restore plan;
7. require explicit maintainer confirmation for replacing/destructive activation;
8. activate safely;
9. rebuild projections;
10. revalidate.

The exact activation/rollback mechanism belongs to #22.

### Redaction and deletion

SQLite does not redefine RFC 0005 redaction/deletion semantics.

Historical records are immutable by default. Redaction/deletion is represented by explicit authoritative metadata/tombstones and must update replay/explanation availability. Stable identities are never reused.

Deleting a row must not be described as forensic erasure. Physical page reuse, backups, filesystem snapshots, and storage media may retain prior bytes. #23 must define any required compaction/secure-delete/encryption workflow and its residual guarantees.

### Data location and permissions

The database lives under a user-controlled Calathea data root outside source checkout. #15/#17 will define platform-specific default path/configuration behavior.

The adapter should create files/directories with restrictive permissions where the platform supports them and avoid logging the database contents or sensitive SQL parameters.

Credentials and plaintext secret values are not Calathea domain records and must not be introduced merely because SQLite can store them.

### Encryption boundary

Selecting ordinary SQLite does not imply application-level database encryption.

For v0, document at-rest protection explicitly. The baseline may rely on OS/user-volume encryption plus restrictive file permissions and encrypted-backup guidance. A future database-encryption decision must be separate, including key custody, recovery, migration, portability, and supply-chain consequences.

## Driver decision

### Preferred candidate: CGO-free `database/sql` SQLite

A CGO-free driver preserves the repository's current Go-focused build boundary and avoids making a C compiler/runtime library an implicit dependency of every developer, CI runner, and release target.

The #14 spike must record:

- exact pinned driver/module versions;
- embedded SQLite version and upgrade policy;
- transitive module count and licensing;
- vulnerability/static-analysis results;
- Nix/mise build reproducibility;
- supported OS/architecture matrix;
- transaction/concurrency behavior needed by #15;
- backup/integrity feature access needed by #22/#25.

`modernc.org/sqlite` is the first candidate, not a permanently mandated dependency. Its dependency/upgrade footprint must be treated as a real cost.

### CGO comparison: `mattn/go-sqlite3`

`mattn/go-sqlite3` is mature and `database/sql` compatible, but it requires CGO and a C compiler. Selecting it would require an explicit revision to the public build/CI/toolchain contract. Do not add that mechanism solely to satisfy persistence unless the spike demonstrates a material benefit over the CGO-free path.

## Alternatives considered

### Structured files/directories

**Not selected as the leading hypothesis.**

Benefits:

- no database dependency;
- potentially human-inspectable artifacts;
- simple isolated-record reads/writes.

Costs for Calathea's actual requirements:

- custom multi-record transaction/commit protocol;
- crash recovery around fsync/rename semantics across platforms;
- custom idempotency and lookup indexes;
- concurrent-writer coordination;
- multi-file schema migration;
- referential/history integrity scanning;
- consistent backup while commands may run;
- projection indexing/rebuild bookkeeping.

This is more mechanism than one embedded transactional database unless the SQLite spike finds a concrete blocker.

### Event-sourced physical store

**Rejected for v0.**

The domain already has explicit immutable versions/decisions. Recasting every mutation as a generic event stream would add event schemas, reducers, snapshot/versioning rules, and debugging surface without a demonstrated need. Rebuildable projections do not by themselves require event sourcing.

### Hosted/server database

**Rejected for v0.**

It violates the offline, user-controlled, no-account baseline and creates network authentication, deployment, availability, synchronization, and secret-management mechanism unrelated to UC-01.

### Multiple SQLite databases

**Rejected for v0.**

They complicate cross-record authoritative atomicity, backup/restore consistency, migration, and integrity validation without providing a concrete isolation requirement.

## Consequences

### Benefits

- transactional atomicity maps naturally to RFC 0005 command boundaries;
- relational constraints/indexes support stable identities, idempotency, and exact version lookup;
- one portable local database simplifies data-root ownership and staged validation;
- projections remain queryable and rebuildable without creating a separate cache engine;
- SQLite provides established integrity and consistent-backup mechanisms;
- no server/account/network dependency;
- a CGO-free driver can preserve straightforward Go packaging.

### Costs / risks

- adds a significant third-party dependency/transitive dependency set to a currently dependency-free module;
- SQLite/driver versions become part of release maintenance and vulnerability response;
- schema migrations become a durable compatibility commitment;
- concurrent-process lock behavior needs deterministic application handling;
- ordinary SQLite is not application-level encryption;
- database corruption can still occur from filesystem/hardware/process misuse and requires recovery tooling;
- careless SQL/schema evolution can violate historical semantics even when the database remains structurally valid.

## Security and privacy impact

Positive:

- keeps sensitive portfolio data local by default;
- supports parameterized queries and strong structural constraints;
- enables staged integrity validation before restore activation;
- provides one data artifact that can receive restrictive filesystem permissions and backup policy.

Required controls:

- parameter binding for untrusted values;
- executable-owned migrations/schema only;
- foreign keys enabled and verified;
- no credentials/plaintext secrets in domain storage;
- no sensitive payloads in diagnostics/logging;
- explicit at-rest/backup encryption guidance;
- supply-chain review and pinned driver versions;
- corruption/restore tests under #9/#24/#25.

Residual risk:

- a compromised local account/root can read or alter the database unless independent OS/storage protections prevent it;
- integrity checks detect corruption/inconsistency but are not cryptographic proof of authorship;
- deleted/redacted bytes may persist in storage/backups until an explicit privacy workflow removes them.

## Compatibility and migration

This decision introduces the first persisted physical schema. Therefore:

- persisted schema versioning starts independently of Go/domain semantic versioning;
- migrations become part of release compatibility;
- future storage replacement must import/validate authoritative records without changing their domain identity/history;
- portable export remains the long-term technology-neutral escape hatch rather than treating raw SQLite layout as the public domain API.

No current user data migration is required because v0 physical persistence has not yet been released.

## Validation

ADR acceptance requires the #14 spike to demonstrate:

- authoritative record/current reference/idempotency atomic transaction;
- lost-response idempotent retry;
- conflicting payload under reused idempotency identity rejected;
- expected-version conflict under stale/concurrent commands;
- projection drop/rebuild equivalence;
- forward migration from a synthetic prior schema;
- newer unsupported schema fail-closed behavior;
- consistent backup snapshot, reopen, structural/FK/domain validation;
- corrupt/truncated restore candidate rejection without current-state mutation;
- selected driver passes formatting/static/unit/integration/vulnerability/build gates;
- package/dependency/license/toolchain impact is documented;
- no network/telemetry/hosted dependency is introduced.

The spike should also record whether rollback-journal mode remains sufficient. WAL should not be enabled without a measured need.

## Evidence / references

Primary implementation references used for the proposal:

- SQLite backup API: https://www.sqlite.org/backup.html
- SQLite `VACUUM INTO`: https://www.sqlite.org/lang_vacuum.html
- SQLite PRAGMA/integrity documentation: https://www.sqlite.org/pragma.html
- `modernc.org/sqlite` package documentation: https://pkg.go.dev/modernc.org/sqlite
- `mattn/go-sqlite3` repository: https://github.com/mattn/go-sqlite3

## Implementation links

- Issue: #14
- Persistence implementation: #15
- Projections/concurrency: #16
- Recovery: #21
- Backup/restore: #22
- Redaction/deletion: #23
- Integration/security gate: #24
- Hardening/doctor: #25
- Evidence PR: pending
