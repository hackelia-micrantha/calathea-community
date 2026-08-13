# ADR 0004 — Go Runtime for Calathea v0

## Status

Accepted

- **Owner:** Calathea maintainers
- **Date:** 2026-08-11
- **Governing constraints:** PRD, UC-01, ADR 0001, ADR 0002, MVP roadmap Phase 1
- **Supersedes:** none
- **Superseded by:** none

## Migration note

This ADR was accepted before the public/private repository split. Its original module path was `github.com/hackelia-micrantha/calathea`. ADR 0005 changes the reusable public module path to `github.com/hackelia-micrantha/calathea-community` while preserving the Go runtime, binary name, package layout, and other decisions here.

## Context

Calathea v0 needs a runtime for a local-first CLI with deterministic domain behavior, fast tests, straightforward cross-platform distribution, and a clean inward dependency direction.

The runtime should make it inexpensive to keep the domain and deterministic services independent from storage, networking, AI providers, GitHub, Anthesis, and other infrastructure. It should also support a later local persistence adapter without requiring a hosted runtime or service.

## Decision

Use **Go** for Calathea v0.

The repository targets the Go 1.26 language/toolchain line and pins the development toolchain to Go 1.26.5 through `mise.toml`.

Initial conventions:

- one Go module; the current public module path is defined by ADR 0005;
- one local executable under `cmd/calathea`;
- application orchestration under `internal/application`;
- domain model under `internal/domain`;
- deterministic orientation services under `internal/orientation`;
- outbound/application-owned contracts under `internal/application/ports`;
- standard library first; dependencies are added only for concrete requirements;
- `mise` is the task/toolchain entry point;
- no Makefile is required;
- no framework, daemon, web server, plugin system, or dependency-injection container is introduced by this decision.

Go package boundaries enforce dependency direction by convention and tests/review: CLI and adapters depend inward on application/domain packages; domain and deterministic packages must not import infrastructure packages.

## Alternatives considered

### Rust

Rust offers stronger compile-time modeling and memory-safety guarantees, and it remains appropriate where low-level safety or hostile-input processing justifies the additional complexity. For Calathea v0, the primary risks are semantic correctness, provenance, deterministic policy/orientation behavior, and historical-state integrity rather than unsafe memory access. Rust would increase implementation and review friction without a demonstrated v0 requirement that offsets that cost.

### Python

Python would minimize initial code volume, but it introduces an interpreter/environment dependency for a product intended to ship as a simple local executable. Its dynamic type system also provides less compile-time support for the domain distinctions Calathea is deliberately preserving.

### TypeScript / Node.js

TypeScript provides strong developer tooling and would be attractive for a future web-facing application. The v0 product has no web requirement, and Node.js would add a runtime/dependency footprint that is unnecessary for the local CLI vertical slice.

## Consequences

### Benefits

- simple native executable distribution;
- fast compilation and test feedback;
- standard tooling for formatting, vetting, tests, and builds;
- interfaces and small packages fit the application-port architecture without a framework;
- mature local persistence ecosystem is available when Phase 2 selects a storage technology;
- cross-platform builds remain straightforward.

### Costs / risks

- Go's type system is less expressive than Rust's for encoding some invariants, so constructors, unexported fields, value types, and tests must be used deliberately;
- package discipline is required to keep infrastructure from leaking inward;
- a future requirement for low-level cryptographic or memory-safety-critical behavior may justify a bounded Rust component, but not a rewrite by default.

## Security and privacy impact

The runtime choice does not add network access, telemetry, hosted identity, or provider dependencies. The core remains capable of running entirely offline.

Dependency additions require review because third-party modules increase supply-chain surface. Credential-bearing values remain outside domain records and traces regardless of runtime.

## Compatibility and migration

Accepted domain/RFC semantics remain language-independent and must not be changed to accommodate Go implementation convenience. Repository/module relocation is governed by ADR 0005.

## Validation

This ADR is validated when:

- the module builds as one local executable;
- unit tests run with only the Go toolchain;
- the local `mise run check` quality gate covers formatting, vetting, tests, and build;
- the package skeleton reflects ADR 0001 dependency direction;
- no optional integration is required for build or test.

The public-core extraction satisfied this baseline; ongoing CI remains the repository quality gate.
