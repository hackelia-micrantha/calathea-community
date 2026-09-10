# ADR 0003 — Optional Adapter Boundaries

## Status

Accepted for the public community kernel.

## Context

The deterministic kernel must remain independently usable without coupling core/domain behavior to repositories, AI providers, instruction systems, governance platforms, or effect executors.

Earlier architecture text described several private-product integration directions in this public ADR. Under the private-product/public-kernel boundary, this ADR now records only the reusable implementation decision required here.

## Decision

External integrations are optional outward adapters behind stable inward-owned ports, introduced only when a concrete public use case requires them.

For the deterministic public kernel:

- no external-source adapter is mandatory;
- no AI/provider/instruction adapter is mandatory;
- no effectful external adapter is part of ambient kernel authority;
- no governance/effect platform is a required core dependency;
- removing optional adapters leaves the documented offline deterministic workflow functional.

Any supported public adapter must keep adapter-specific transport/configuration outside the domain core and preserve the kernel's validation, provenance, privacy, and failure invariants.

If effectful capabilities are ever added publicly, authorization/approval and effect execution must be explicit separate concerns rather than inferred from a recommendation or ordinary adapter access.

## Consequences

### Benefits

- preserves local/offline standalone operation;
- avoids speculative platform coupling;
- keeps provider-specific details out of deterministic/core semantics;
- allows independently useful adapters to evolve behind explicit contracts;
- prevents recommendation behavior from acquiring accidental effect authority.

### Costs

- each concrete adapter requires an explicit public contract and tests;
- private Calathea integrations may need translation around the public process/file/schema surface;
- convenience features from a particular provider cannot leak into the core API without compatibility review.

## Boundary with private Calathea

The private product may choose specific AI, instruction, governance, source, or effect integrations and may define stricter policies around them.

Those choices do not become public-kernel dependencies or ADR concerns unless a concrete independently useful public adapter is deliberately supported.
