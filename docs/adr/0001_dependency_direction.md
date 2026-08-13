# ADR 0001 — Dependency Direction and Core Isolation

## Status

Accepted for v0 planning.

## Context

Calathea must remain local-first, deterministic without network access, and independent of external repositories, AI providers, and governance systems. The existing RFCs define stable domain semantics that should not be coupled to infrastructure choices.

## Decision

Use inward dependency direction:

```text
Adapters → Application Services → Domain + Deterministic Services
```

Outbound ports required by application services are owned by the inward application/core boundary and implemented by outward adapters. Domain and deterministic services remain independent of infrastructure ports unless a future pure domain abstraction is itself part of domain semantics.

The domain and deterministic services must not depend on:

- CLI frameworks;
- storage engines or storage ports;
- GitHub or other source SDKs;
- AI-provider SDKs or provider ports;
- Anthesis or any governance product;
- network transports;
- hosted services.

Application services may depend on inward-owned ports such as persistence, clock, identity, read-only source, and optional AI-provider invocation interfaces.

## Consequences

Benefits:

- UC-01 remains runnable entirely locally;
- deterministic core behavior is easy to test;
- provider and storage choices remain replaceable;
- future integration choices cannot leak into domain semantics.

Costs:

- explicit ports and mapping code are required;
- application and domain boundaries must be kept distinct rather than treating every interface as a domain abstraction;
- some framework convenience is intentionally avoided at the core boundary.
