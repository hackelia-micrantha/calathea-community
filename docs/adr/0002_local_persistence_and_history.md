# ADR 0002 — Local Persistence, Immutable History, and Rebuildable Projections

## Status

Accepted for v0 planning.

## Context

Calathea owns sensitive portfolio data and requires deterministic replay, explicit history, redaction, backup/restore, and offline operation. RFC 0005 requires immutable historical records without prescribing event sourcing or a database.

## Decision

Calathea v0 uses user-controlled local persistence behind application-owned storage ports.

Domain persistence semantics are:

- canonical historical versions and decisions are immutable after durable creation;
- deterministic recommendation records are immutable;
- corrections create superseding records;
- current views are rebuildable projections;
- authoritative command boundaries are atomic;
- operation identity supports idempotent retry;
- private data is not committed to the source repository by default;
- no hosted storage, telemetry, or cloud account is required.

Where a current selection cannot be derived from record creation order alone, Calathea retains an immutable maintainer-authored selection decision sufficient to rebuild it. Current policy selection therefore references an exact `PolicySetVersion`, and re-selecting an older policy version does not silently mutate historical meaning. The exact canonical entity name belongs in the policy/domain RFCs rather than being introduced by this architecture ADR.

This decision does **not** select SQLite, files, event sourcing, or any other physical storage technology.

## Consequences

Benefits:

- strong offline/privacy posture;
- historical decisions remain explainable;
- projections can be repaired without rewriting history;
- policy selection can be reconstructed even after rollback/re-selection;
- implementation remains free to choose a simple local storage technology.

Costs:

- backup, restore, migration, validation, redaction, and projection rebuilding must be treated as product features;
- supporting explicit selection records adds a small amount of history;
- local compromise remains outside the tamper-proof guarantee.
