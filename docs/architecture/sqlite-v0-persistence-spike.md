# SQLite v0 Persistence Spike Evidence

## Status

**In progress — executable driver evidence pending repository CI.**

- **Tracked by:** #38
- **ADR under test:** proposed ADR 0006 / PR #37
- **Candidate:** `modernc.org/sqlite` v1.56.0
- **Pinned ABI dependency:** `modernc.org/libc` v1.74.4
- **Project toolchain:** Go 1.26.5 through `mise`
- **Candidate minimum Go line:** Go 1.25
- **Candidate SQLite build:** SQLite 3.53.3 for Linux amd64 in v1.56.0 documentation
- **Driver model:** `database/sql`, CGO-free
- **Candidate license:** BSD-3-Clause

This report records the executable evidence required before ADR 0006 may move from **Proposed** to **Accepted**. The spike is intentionally a small representative durability harness, not the production persistence adapter or final table schema.

## Primary references

- `modernc.org/sqlite` package documentation: <https://pkg.go.dev/modernc.org/sqlite>
- canonical modernc SQLite repository: <https://gitlab.com/cznic/sqlite>
- v1.56.0 module metadata: <https://github.com/modernc-org/sqlite/blob/v1.56.0/go.mod>
- SQLite `VACUUM INTO`: <https://sqlite.org/lang_vacuum.html>
- SQLite PRAGMA/integrity documentation: <https://sqlite.org/pragma.html>

The package documentation identifies the driver as a CGO-free SQLite port, documents `database/sql` usage, validated DSN settings including foreign keys/journal/synchronous behavior, and warns that the exact `modernc.org/libc` version from its `go.mod` must remain aligned.

## Representative schema

The spike models only enough physical state to test RFC 0005 invariants:

- `calathea_meta` — Calathea-owned persisted schema version;
- `authoritative_records` — immutable representative authoritative versions;
- `current_records` — one required atomic current reference;
- `operations` — stable operation identity, versioned material-request fingerprint, and committed result reference;
- `record_projection` — disposable/rebuildable projection introduced by the synthetic v0→v1 migration.

The schema is deliberately not a proposal for all #15 domain tables.

## Canonical test request fingerprint

The idempotency experiment does not hash Go formatting or ordinary JSON bytes. It defines an explicit spike-only fingerprint schema, `sqlite-spike.write-request.v1`, and hashes a fixed ordered sequence of length-delimited fields:

- entity identity;
- requested record identity;
- payload;
- expected revision;
- new revision.

`OperationID` scopes the idempotency lookup and is not itself used to decide whether a retry payload is materially equivalent.

Production command fingerprints remain a #15 application-contract decision and must be separately versioned.

## Connection / durability hypothesis under test

Every candidate connection uses driver-validated settings:

- foreign keys: enabled;
- busy timeout: 250 ms;
- journal mode: rollback journal (`DELETE`), not WAL;
- synchronous: `FULL`;
- double-quoted string literal compatibility: disabled.

The harness creates the data file with mode `0600` and its parent with mode `0700` where the platform supports Unix permissions.

The spike separately classifies SQLite `BUSY`/`LOCKED` result families as an operational storage-contention condition. This is distinct from a Calathea semantic expected-revision conflict.

## Executable experiment matrix

| Experiment | Evidence in harness | Current status |
| --- | --- | --- |
| CGO-free build | `CGO_ENABLED=0 go test` and `go build` through `mise run sqlite-spike` | Pending CI |
| Foreign keys/settings | read back `foreign_keys`, journal mode, synchronous level | Pending CI |
| Atomic authoritative command | injected pre-commit failure leaves record/current/operation counts at zero | Pending CI |
| Lost-response retry | repeated operation + equivalent fingerprint returns original result without another record | Pending CI |
| Conflicting retry | reused operation identity + different material payload returns idempotency conflict | Pending CI |
| Storage contention vs semantic conflict | active write lock yields storage-busy classification; stale expected revision yields expected-conflict classification | Pending CI |
| Projection rebuild | corrupt projection, rebuild from authoritative records, compare semantic rows and authoritative counts | Pending CI |
| Schema migration | synthetic persisted schema 0 → 1, repeat is recognized, newer unsupported schema fails closed | Pending CI |
| Structural integrity | `PRAGMA integrity_check` | Pending CI |
| Foreign-key integrity | `PRAGMA foreign_key_check` | Pending CI |
| Domain/history integrity | structurally/FK-valid cross-entity current reference is rejected by Calathea validator | Pending CI |
| Corrupt restore candidate | truncated copy fails validation while original reopens unchanged | Pending CI |
| Consistent backup | parameter-bound `VACUUM INTO` creates independent staged snapshot and full validation succeeds | Pending CI |
| Private local file | no group/other file permission bits in Linux test | Pending CI |

## Backup choice under test

The first spike uses `VACUUM INTO` because SQLite defines it as a consistent snapshot of a live source database and allows the destination filename to be a scalar SQL expression, so the path can remain parameter-bound.

The destination is not trusted merely because `VACUUM INTO` returns success. The harness reopens it independently and requires:

1. supported Calathea persisted schema;
2. SQLite structural integrity;
3. foreign-key integrity;
4. Calathea domain/history integrity.

The production #22 design may still prefer the online backup API if incremental behavior, interruption handling, or measured cost makes it superior.

## Dependency / supply-chain evidence

The selected v1.56.0 module declares Go 1.25 and directly requires:

- `github.com/google/pprof`;
- `golang.org/x/sys`;
- `modernc.org/fileutil`;
- `modernc.org/libc` v1.74.4;
- `modernc.org/mathutil`;

with additional indirect dependencies in the module graph.

Before recommending ADR acceptance, repository CI must produce the actual Calathea module graph and `go.sum`, and the PR review must record:

- complete module delta after `go mod tidy`;
- license review for the resolved graph;
- `govulncheck` result under #9;
- static-analysis result;
- supported Calathea release target build evidence;
- driver/embedded-SQLite upgrade ownership.

The dependency footprint is therefore a measured acceptance cost, not assumed negligible merely because the driver is CGO-free.

## Journal mode result

**Pending executable evidence.**

The proposal starts in rollback-journal mode. WAL should be adopted only if a later measured local workload demonstrates a need that outweighs checkpoint, sidecar-file, backup, and recovery complexity.

## Current recommendation

**Continue the SQLite hypothesis; do not accept ADR 0006 yet.**

The architecture fit remains strong: SQLite directly supplies the transaction, uniqueness/index, migration, consistent-snapshot, and integrity mechanisms that a structured-file implementation would otherwise need to recreate. However, the driver and full module graph have not yet passed the repository's pinned Go 1.26.5 quality/security pipeline on the self-hosted runner.

### Acceptance decision rule

After the #38 branch executes:

- **Accept** ADR 0006 if all invariant experiments pass, CGO-disabled build succeeds, and dependency/security/package impact is reasonable.
- **Revise** if SQLite works but driver, journal, backup, or packaging details need a different bounded choice.
- **Reject** if the physical/driver mechanism introduces disproportionate correctness, portability, supply-chain, or recovery complexity compared with structured files.
