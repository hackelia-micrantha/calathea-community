# Public/private repository boundary

## Status

Accepted repository-ownership rule for the Calathea public/private split.

The executable-core migration is complete. Reusable product and architecture contracts are being promoted under the documentation migration slice.

## Repositories

- `hackelia-micrantha/calathea-community` is the public reusable core.
- `hackelia-micrantha/calathea` is the private composition and dogfood repository.

The split is an ownership boundary, not a periodic mirror.

## Target ownership

### `calathea-community` owns

Reusable Calathea product semantics and implementation that can be developed in public, including:

- deterministic domain and orientation logic;
- generic application services and application-owned ports;
- the public CLI and reusable local persistence implementation;
- public schemas, migrations, fixtures, golden tests, and conformance tests;
- public PRD/use-case/roadmap material that defines the reusable product;
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
                | process / files / versioned contracts
                v
calathea-community (public reusable core)
```

The public repository must not depend on the private repository for normal build, test, or documented core workflows.

Private composition pins a reviewed public-core revision. A reusable capability needed by both repositories should normally move into `calathea-community` rather than be copied into the private repository.

## Go module and package boundary

The public module is:

```text
github.com/hackelia-micrantha/calathea-community
```

The user-facing binary remains `calathea`.

The reusable Go implementation remains primarily under `internal/`. This is intentional: repository topology did not justify turning the existing implementation tree into a broad external library API.

For v0, the supported cross-repository surface is:

- the `calathea` executable;
- explicitly documented CLI behavior and exit semantics;
- versioned file/schema contracts as they are introduced.

The private repository must not import or vendor `calathea-community/internal/...` packages.

A future exported Go facade requires a concrete in-process consumer and a separate compatibility decision. See [ADR 0005](../adr/0005_public_go_process_boundary.md).

## Source-of-truth rule

There must be exactly one canonical repository for a maintained path or semantic contract.

During migration:

1. A path remains private-canonical until its public replacement is reviewed and validated.
2. The public artifact is promoted with private-specific data/references removed or generalized.
3. The corresponding private duplicate is removed or replaced with an explicit public reference/extension.
4. Only then does the public artifact become canonical.
5. Subsequent reusable changes are made publicly and consumed privately through the documented boundary.

Do not maintain equivalent source files in both repositories by manual synchronization.

## Promotion rule

A private capability or contract may be promoted when all of the following hold:

- it is reusable without private portfolio data;
- its public API/contract and behavior can be documented independently;
- tests do not require private credentials, infrastructure, or data;
- sensitive defaults and identifiers can be removed or parameterized;
- the licensing and dependency chain is compatible with the public repository;
- security review finds no private operational detail that should remain undisclosed.

Promotion should preserve accepted semantic history where practical. Correctness and a clean ownership boundary take precedence over preserving private issue-number topology or an awkward repository layout.

## Private-data boundary

The public core must preserve Calathea's local-first privacy invariants:

- user portfolio data lives outside the source checkout by default;
- deterministic core operation requires no network access;
- telemetry and hosted accounts are not required;
- external-source and AI integrations are explicit and optional;
- credentials are resolved out of band and are never durable project/evidence/prompt data.

These properties are product behavior, not reasons to keep the reusable implementation private.

## Migration status

### Slice 1 — repository boundary

**Complete.**

- public repository identity and MPL-2.0 licensing established;
- ownership/source-of-truth rules documented;
- coordinated public/private migration tracking established.

### Slice 2 — executable deterministic core

**Complete.**

- public Go module and `calathea` executable established;
- reusable CLI/application/domain foundation, tests, fixtures, and quality tooling promoted;
- process/file/schema boundary selected instead of a speculative broad Go API;
- private duplicate Go implementation removed;
- private composition pins the reviewed public revision.

### Slice 3 — reusable contracts and documentation

**In progress under public issue #3.**

Promotes:

- PRD, use cases, and MVP roadmap;
- RFC 0000–0008 and RFC governance/template;
- ADR 0001–0004 accepted history plus ADR 0005 for the public-module/process decision;
- domain/runtime/system-context and optional Invokrum/structured-invocation architecture contracts.

Private-specific issue topology, data, and annotations are not part of the public contracts.

### Slice 4 — thin private composition

The private repository should retain only composition, private data/configuration, dogfood operations, and genuinely non-public extensions. Reusable docs migrated in Slice 3 are replaced privately by public references or explicit private extensions rather than maintained as copies.

## CI and release expectations

`calathea-community` is independently buildable and testable from a clean checkout without access to the private repository or private test data.

Public CI may use ordinary organization execution infrastructure, including self-hosted runners, but its source, dependencies, fixtures, and required secrets must not depend on the private `calathea` repository.

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
