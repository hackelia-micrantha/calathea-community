# Public/private repository boundary

## Status

Accepted repository-ownership rule for the Calathea public/private split.

## Repositories

- `hackelia-micrantha/calathea-community` is the public repository.
- `hackelia-micrantha/calathea` is the private repository.

The split is an ownership boundary, not a periodic mirror.

## Target ownership

### `calathea-community` owns

Reusable Calathea product semantics and implementation that can be developed in public, including:

- deterministic domain and orientation logic;
- generic application services and application-owned ports;
- the public CLI and reusable local persistence implementation;
- public schemas, migrations, fixtures, golden tests, and conformance tests;
- public PRD/use-case material that defines the reusable product;
- public RFCs and ADRs that define reusable semantics or architecture;
- public contracts for optional adapters such as Invokrum integration;
- generic documentation, examples, and development tooling.

### `calathea` owns

Private composition and data that should not be required to use or contribute to the public core, including:

- real portfolio/project/evaluation/evidence data;
- private policy calibration, weights, heuristics, and experiments;
- private instruction/profile packs and model configuration;
- private integration and deployment configuration;
- private dogfood/operations material;
- proprietary or experimental extensions that have not been intentionally promoted to the public core.

Credentials and secret values belong in neither repository.

## Dependency direction

The intended dependency direction is:

```text
calathea (private composition / extensions)
                |
                v
calathea-community (public reusable core)
```

The public repository must not depend on the private repository for normal build, test, or documented core workflows.

Private code may depend on a released or pinned public-core version. A reusable capability needed by both repositories should normally move down into `calathea-community` rather than being copied upward into the private repository.

## Go module and package boundary

The current private implementation cannot be copied mechanically.

Today its module path is `github.com/hackelia-micrantha/calathea`, and much of the reusable implementation lives under Go `internal/` packages. A sibling repository cannot import packages beneath `calathea-community/internal/...`, and the public repository cannot retain the private repository's module path as its canonical import identity.

Before the executable-core migration lands:

- select the public module path, expected to be `github.com/hackelia-micrantha/calathea-community` unless a deliberate vanity path is introduced;
- define the smallest exported library/facade needed by private composition and external consumers;
- keep implementation details under `internal/` where they do not need cross-repository use;
- do not expose every current internal package merely to make the split compile;
- keep the user-facing binary name `calathea` independent of the Go module/repository name;
- migrate imports and tests deliberately and validate that the private repository can consume only supported public surfaces.

A subprocess/CLI contract may be appropriate for isolated integrations, but it should not be used solely to avoid defining a coherent public Go API when in-process composition is required.

## Source-of-truth rule

There must be exactly one canonical repository for a maintained path or semantic contract.

During the migration from the existing private-first repository:

1. A path remains private-canonical until its migration slice is complete.
2. A migration slice copies or reconstructs the public artifact, validates it, and removes the private duplicate or replaces it with a dependency/reference.
3. Once that slice lands, the public artifact becomes canonical.
4. Subsequent changes are made in the canonical repository and consumed by the other repository through an explicit dependency, generated artifact, or reference.

Do not maintain equivalent source files in both repositories by manual synchronization.

## Promotion rule

A private capability may be promoted when all of the following hold:

- it is reusable without private portfolio data;
- its public API and behavior can be documented independently;
- tests do not require private credentials, infrastructure, or data;
- sensitive defaults and identifiers can be removed or parameterized;
- the licensing and dependency chain is compatible with the public repository;
- security review finds no private operational detail that should remain undisclosed.

Promotion should preserve history where practical, but correctness and a clean ownership boundary take precedence over preserving an awkward repository layout.

## Private-data boundary

The public core must preserve Calathea's local-first privacy invariants:

- user portfolio data lives outside the source checkout by default;
- deterministic core operation requires no network access;
- telemetry and hosted accounts are not required;
- external-source and AI integrations are explicit and optional;
- credentials are resolved out of band and are never durable project/evidence/prompt data.

These properties are product behavior, not reasons to keep the reusable implementation private.

## Migration sequence

### Slice 1 — establish the boundary

- update the public repository identity and README;
- add an explicit public license;
- document ownership and transition rules;
- create tracked migration work in both repositories.

### Slice 2 — define the public Go surface and move the executable deterministic core

Resolve module/package ownership first, then move the Go module, CLI, reusable domain/orientation implementation, tests, generic application boundary, and CI/tooling needed to build it independently.

The slice is complete only when the corresponding private copies are removed or converted into public-core consumption. Code that remains internal implementation detail should stay behind the new public API rather than being exported by default.

### Slice 3 — move reusable contracts and documentation

Move the reusable PRD/use cases, RFCs, ADRs, architecture contracts, schemas, fixtures, and conformance material. Private-specific annotations should remain private rather than contaminating the public contracts.

### Slice 4 — make the private repository thin

The private repository should contain only composition, private data/configuration, and genuinely non-public extensions. It should pin the public-core version it consumes and document any temporary exceptions.

## CI and release expectations

`calathea-community` should eventually be independently buildable and testable from a clean checkout. Public CI must not rely on private repositories, private runners beyond ordinary execution infrastructure, or private test data.

The private repository may run additional integration/dogfood validation against a pinned public-core revision.

## Licensing

Public Calathea source is licensed under MPL-2.0 unless a file or bundled dependency states otherwise.

The private repository is not made public or relicensed merely because it depends on MPL-2.0 public code. Changes to MPL-covered public files remain subject to the license terms.

## Non-goals

This split does not:

- publish private portfolio data;
- make private experiments part of the public support contract;
- require a hosted service or SaaS boundary;
- create two editions that independently implement the same core semantics;
- require Anthesis or Invokrum for deterministic Calathea operation.
