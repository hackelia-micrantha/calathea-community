# ADR 0003 — Optional Integration Boundaries and No Anthesis Dependency

## Status

Accepted for v0 planning.

## Context

Calathea may later integrate with repositories, AI providers, or effectful external systems. Earlier planning risked coupling the architecture directly to Anthesis despite v0 being read-only and human-authoritative.

## Decision

All v0 external integrations are optional outward adapters behind stable inward-owned ports.

For v0:

- repository/source adapters are optional and read-only;
- AI-provider adapters are optional and non-authoritative;
- no effectful external adapter is required;
- no effect-governance or effect-execution port exists;
- Calathea has no dependency on Anthesis.

If effectful capabilities are added later, a new RFC/ADR must define their boundaries. Authorization/approval and effect execution must remain distinct: a governance adapter may decide whether an effect is permitted, while a separate effect adapter performs the external mutation and records its result.

Anthesis may implement a future governance/authorization adapter, but Anthesis-specific identities, policy objects, approval models, or effect semantics must not enter the Calathea core domain.

Removing every optional integration adapter must leave the deterministic UC-01 workflow functional.

## Consequences

Benefits:

- avoids speculative platform coupling;
- preserves local/offline operation;
- preserves the domain invariant that decisions are not effects;
- allows Anthesis integration later without making it foundational;
- allows alternative governance or execution implementations if requirements change.

Costs:

- effectful workflows require a later explicit architecture decision;
- future adapters may require translation between Calathea domain concepts and external governance/effect models;
- Calathea cannot assume Anthesis-specific convenience features in its core API.
