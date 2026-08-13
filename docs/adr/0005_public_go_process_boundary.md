# ADR 0005 — Public Go Module and Process Boundary

## Status

Accepted for the public-core extraction.

- **Decision date:** 2026-08-12
- **Governing constraints:** repository boundary, ADR 0001, ADR 0004
- **Supersedes:** ADR 0004 only with respect to repository/module identity

## Context

The reusable Calathea implementation was developed in the private `hackelia-micrantha/calathea` repository under module path `github.com/hackelia-micrantha/calathea`. Most implementation packages are intentionally under Go `internal/`.

Moving those packages to sibling repository `hackelia-micrantha/calathea-community` creates an important boundary choice. Exporting the current domain/application package tree merely so the private repository can import it would convert implementation details into a compatibility promise without an actual external library use case.

Calathea already has a natural executable boundary: the local `calathea` CLI. The product is local-first, deterministic without AI, and designed to exchange versioned records and eventually structured command output rather than share mutable in-process state across repositories.

## Decision

### Module

The public source module is:

```text
github.com/hackelia-micrantha/calathea-community
```

The binary remains named:

```text
calathea
```

Repository name and binary/product name intentionally differ.

### Public integration surface

The supported cross-repository integration surface is the `calathea` executable plus explicitly versioned file/schema/CLI contracts as they are introduced.

There is **no supported exported Go library API in v0**.

The deterministic domain, orientation, and application implementation remain under `internal/`. Their Go identifiers may be exported within those packages for internal composition and testing, but their package import paths are not public compatibility contracts.

### Private repository consumption

The private `hackelia-micrantha/calathea` repository must not import `calathea-community/internal/...` and must not retain a duplicate fork of the deterministic core.

Private composition should use one or more of these supported boundaries:

- invoke a pinned/released `calathea` executable;
- provide user-controlled local data/configuration through documented paths or flags;
- consume structured output governed by explicit schema/version contracts;
- run private dogfood/integration validation against a pinned public revision.

A future need for in-process Go composition is not solved by exporting the existing internal tree. It requires a separate ADR backed by a concrete consumer and the smallest stable facade that consumer needs.

## Compatibility

Until structured commands are implemented, compatibility is intentionally narrow:

- binary name `calathea` is stable;
- documented command names and exit-code semantics become compatibility commitments when introduced;
- structured machine-readable output must carry or reference a schema/semantic version before it is treated as stable integration surface;
- `internal/` package paths and types carry no external compatibility guarantee.

Version pinning is preferred for private composition until a stable release/version policy is established.

## Dependency direction

```text
calathea (private data/config/extensions)
        |
        | CLI / files / versioned schemas
        v
calathea-community (public executable core)
```

The public repository must build and test without access to the private repository.

## Provenance and licensing

The initial executable-core files were promoted from private `hackelia-micrantha/calathea` at commit `f1b5bdd625c7db890be2d024031d02b675e9b7e4` and merged publicly in `calathea-community` PR #7.

The repository owner intentionally publishes the promoted reusable source in `calathea-community` under MPL-2.0. The extraction changes repository/module identity and removes private-only references where required; it does not intentionally change deterministic domain semantics.

## Consequences

### Benefits

- preserves Go `internal` encapsulation;
- avoids prematurely freezing a broad library API;
- prevents the repository split from dictating domain architecture;
- keeps the public executable independently buildable and useful;
- permits the private repo to remain thin without a duplicate core;
- leaves room for a deliberately designed library facade later.

### Costs

- private in-process extensions cannot directly call internal domain packages;
- process/schema integration needs explicit versioning and error semantics;
- some future extension use cases may justify a new public Go facade.

## Rejected alternatives

### Export the current domain/application packages wholesale

Rejected because repository topology is not sufficient justification for a large compatibility surface. It would constrain refactoring and make implementation-oriented types part of the public contract.

### Keep the Go core private and expose only copied binaries

Rejected because the deterministic reusable core is intentionally part of `calathea-community`; keeping source private would defeat the public-core ownership decision.

### Maintain matching Go source in both repositories

Rejected because it creates split-brain semantics and violates the repository source-of-truth rule.

## Follow-up

- public issue #5 defines the durable CLI/process compatibility contract as functional commands are introduced;
- public issue #3 migrates reusable product/RFC/ADR documentation;
- future exported Go APIs require a separate consumer-driven decision.
