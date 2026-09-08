# Calathea Community Kernel Architecture Decision Records

Public ADRs record concrete architecture or technology choices made **within this repository's supported kernel scope**.

They do not own the complete private Calathea product strategy, planning/review/lifecycle model, roadmap, calibration, or AI paved-road policy.

## ADR statuses

- **Proposed** — concrete public-kernel architecture choice under review.
- **Accepted** — effective architecture decision for the stated public scope.
- **Superseded** — replaced by a named later ADR, retained for history.
- **Rejected** — considered but deliberately not adopted.

Accepted ADRs should identify the public behavior, compatibility need, RFC, issue, or implementation evidence they constrain.

A private Calathea product requirement may motivate a public ADR, but the ADR should describe only the extracted public implementation decision rather than restating the private product rationale wholesale.

## Current ADR index

| ADR | Status | Scope |
| --- | --- | --- |
| [0001 — Dependency Direction and Core Isolation](0001_dependency_direction.md) | Accepted | Inward dependency direction; public deterministic/core isolation |
| [0002 — Local Persistence, Immutable History, and Rebuildable Projections](0002_local_persistence_and_history.md) | Accepted | Public local persistence architecture required by supported kernel behavior |
| [0003 — Optional Integration Boundaries and No Anthesis Dependency](0003_optional_integration_boundaries.md) | Accepted / audit pending | Public optional-adapter architecture; issue #48 will remove any private-product strategy that is not required here |
| [0004 — Go Runtime for Calathea v0](0004_go_runtime.md) | Accepted | Go runtime and single local executable for this public implementation |
| [0005 — Public Go Module and Process Boundary](0005_public_go_process_boundary.md) | Accepted | Public module identity, `internal/` encapsulation, process/schema integration surface |

ADR 0005 amends ADR 0004 only where repository/module identity changed during the public-core extraction. It does not supersede the Go runtime choice.

## ADR rules

1. State the concrete public implementation decision, not only background analysis.
2. Identify the supported kernel behavior or compatibility need that justifies the decision.
3. Do not reproduce private product strategy merely to explain why a public mechanism exists.
4. List meaningful alternatives and trade-offs.
5. Record security/privacy/operability/migration consequences where material.
6. Link public implementation issues and validation evidence where applicable.
7. Do not silently rewrite accepted ADR history after implementation; supersede substantive changed decisions.
8. Provider/product-specific policy stays private unless a concrete public adapter requires a stable interoperable boundary.

## Traceability

Public implementation work should follow the chain in [RFC governance](../rfcs/README.md):

```text
public kernel use case / compatibility need
  -> RFC or ADR when durable
  -> issue
  -> PR
  -> synthetic validation / compatibility evidence
```

When the requirement originates in private Calathea:

```text
private product requirement
  -> minimum extracted public mechanism
  -> public ADR/RFC/issue/PR
  -> validation
  -> private reviewed public-core pin
```

Historical issue references remain provenance. New public implementation work belongs in this repository only when it changes the supported public kernel.
