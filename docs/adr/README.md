# Calathea Architecture Decision Records

ADRs record concrete accepted architecture or technology choices made within the product and RFC constraints.

They do not replace the PRD, use cases, or RFCs.

## ADR statuses

- **Proposed** — concrete architecture choice under review.
- **Accepted** — effective architecture decision.
- **Superseded** — replaced by a named later ADR, retained for history.
- **Rejected** — considered but deliberately not adopted.

The Calathea maintainer is the v0 decision authority. Accepted ADRs should link their governing RFC/use-case constraints and implementation evidence.

## Current ADR index

| ADR | Status | Scope |
| --- | --- | --- |
| [0001 — Dependency Direction and Core Isolation](0001_dependency_direction.md) | Accepted | Inward dependency direction; domain/deterministic isolation |
| [0002 — Local Persistence, Immutable History, and Rebuildable Projections](0002_local_persistence_and_history.md) | Accepted | Technology-neutral local persistence semantics |
| [0003 — Optional Integration Boundaries and No Anthesis Dependency](0003_optional_integration_boundaries.md) | Accepted | Optional read-only/provider boundaries; no v0 effect/governance port |
| [0004 — Go Runtime for Calathea v0](0004_go_runtime.md) | Accepted | Go 1.26 line, single local executable, mise task/toolchain entry point |
| [0005 — Public Go Module and Process Boundary](0005_public_go_process_boundary.md) | Accepted | Public module identity, `internal/` encapsulation, process/schema integration surface |

ADR 0005 amends ADR 0004 only where repository/module identity changed during the public-core migration. It does not supersede the Go runtime choice.

## ADR rules

1. State the concrete decision, not only background analysis.
2. Cite the PRD/use case/RFC constraints the decision must satisfy.
3. List meaningful alternatives and tradeoffs.
4. Record security/privacy/operability/migration consequences where material.
5. Link implementation issues and validation evidence where public tracking exists.
6. Do not silently rewrite accepted ADR history after implementation; supersede substantive changed decisions.
7. Provider-specific implementation detail stays outside core/domain ADRs unless it changes a durable architecture boundary.

## Traceability

Use the repository-wide chain defined in [RFC governance](../rfcs/README.md):

```text
PRD → use case → RFC / ADR → issue → PR → validation evidence
```

Historical bare issue numbers in pre-split decision records refer to the original private implementation backlog and are provenance, not public normative dependencies. New implementation work should be tracked in this public repository when it changes the reusable core.
