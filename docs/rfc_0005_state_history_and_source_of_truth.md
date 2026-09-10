# RFC 0005 — Public State, History, Replay, and Persistence Invariants

## Status

Accepted public-kernel contract, narrowed under issue #48.

The private Calathea product owns broader product authority, retention, planning/review/lifecycle history, and private data policy. This RFC defines only the persistence/history guarantees required by supported public formats and deterministic replay.

## Public record classes

### Versioned caller-authored records

Records accepted through a supported public command/file/schema surface and retained with stable identity/version where the contract promises history.

### Deterministically derived records

Outputs such as orientation runs, score/policy traces, comparisons, and projections derived from exact versioned inputs.

### Imported records

If a supported adapter exists, external data is retained with source identity/revision/time as required by that adapter contract. Importing does not transfer source authority.

### Projections

Replaceable current views that are rebuildable from the authoritative/versioned public records specified by the persistence contract.

## Historical integrity

When a public record is specified as immutable/versioned:

- material correction creates a new version/superseding record;
- prior versions used by historical deterministic results are not silently rewritten;
- projection updates do not replace the underlying history;
- unsupported historical versions remain explicit rather than being reinterpreted as current.

## Atomicity and idempotency

A supported authoritative write must either commit the records promised by its command boundary or report failure without claiming an ambiguous partial success.

Where the process contract supports idempotency, operation identity may be used so a retry can return the already committed result rather than duplicate it.

## Concurrency

Local CLI usage may still race through retries or concurrent invocations.

Use, where required by the public persistence contract:

- expected-version/precondition checks;
- immutable version creation;
- deterministic conflict diagnostics;
- rebuildable projections;
- explicit idempotency identifiers.

Avoid undocumented last-write-wins for history-bearing records.

## Replay

A deterministic replay claim requires the exact supported versions of:

- persisted/input records;
- policy/evaluator configuration that materially affected output;
- deterministic semantic versions;
- required retained external evidence/snapshots when a public adapter contract made them inputs;
- stable tie-break/rounding semantics.

If any required version/artifact is unavailable, replay fails explicitly rather than substituting current state.

Provider-private or model-internal computation is not part of deterministic replay unless a retained provider result is explicitly treated as historical input.

## Migrations

A public persisted format change must define:

- source and target version;
- deterministic migration behavior where feasible;
- validation/error handling;
- whether rollback/downgrade is supported;
- compatibility consequences for older binaries/records;
- tests using synthetic fixtures.

Unknown/unsupported versions fail rather than being silently coerced.

## Backup, restore, and recovery

Where supported, backup/restore should preserve:

- stable record identities;
- version/supersession relationships;
- semantic/version metadata needed for replay;
- provenance required by public adapter contracts;
- distinction between history-bearing records and rebuildable projections.

Restore validates integrity/version compatibility before claiming success.

Projection failure after historical/authoritative commit is repaired by rebuild; it does not justify rewriting committed history.

## Privacy/safety

- credentials are never durable public record/evidence content;
- public fixtures are synthetic;
- export/backup paths must not assume private Calathea data or infrastructure;
- optional adapter failure cannot silently replace valid deterministic state with incomplete data.

## Private product boundary

Private Calathea owns retention/deletion choices for private planning, review, AI, stakeholder, and portfolio records beyond the public formats defined here.

A broader source-of-truth policy is not a public compatibility requirement merely because the public kernel provides reusable persistence machinery.
