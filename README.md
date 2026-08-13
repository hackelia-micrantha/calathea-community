# Calathea

> Local-first portfolio orientation for governed human-and-AI planning.

[![Last Commit](https://img.shields.io/github/last-commit/hackelia-micrantha/calathea-community)](https://github.com/hackelia-micrantha/calathea-community/commits/main)
[![Issues](https://img.shields.io/github/issues/hackelia-micrantha/calathea-community)](https://github.com/hackelia-micrantha/calathea-community/issues)
[![Pull Requests](https://img.shields.io/github/issues-pr/hackelia-micrantha/calathea-community)](https://github.com/hackelia-micrantha/calathea-community/pulls)
[![License](https://img.shields.io/github/license/hackelia-micrantha/calathea-community)](LICENSE)

## Overview

Calathea helps a technical maintainer decide what to work on now, what should come next, what can wait, and what should be stopped.

The product is deliberately local-first and human-authoritative. Its deterministic core converts explicit project evaluations and policy constraints into explainable `now`, `next`, `later`, and `kill` recommendations. AI assistance and external repository signals are optional extensions rather than prerequisites.

A `kill` placement is a recommendation to stop investing. It is not an automatic lifecycle transition, deletion, archival action, or repository mutation.

## Public repository role

This repository is the canonical public home for reusable Calathea semantics and implementation.

Calathea uses a public-core/private-composition split:

- **`calathea-community`** owns reusable product contracts, deterministic implementation, public tooling, and supported integration boundaries;
- **`calathea`** retains private portfolio data/configuration, dogfood operations, and intentionally private extensions.

See [the repository boundary](docs/architecture/repository-boundary.md) for ownership rules and [ADR 0005](docs/adr/0005_public_go_process_boundary.md) for the Go/process boundary.

### Public core

`calathea-community` owns:

- deterministic domain and orientation logic;
- generic application services and application-owned ports;
- the `calathea` CLI and reusable local persistence implementation as it is added;
- schemas, migrations, fixtures, golden tests, and conformance tests;
- the public [PRD](docs/product/prd.md), [use cases](docs/product/use-cases.md), and [MVP roadmap](docs/product/mvp-roadmap.md);
- [RFCs](docs/rfcs/README.md), [ADRs](docs/adr/README.md), and architecture contracts;
- public contracts for optional integrations such as Invokrum;
- generic examples, documentation, CI, and development tooling.

### Private composition

The private `calathea` repository retains:

- real portfolio, evaluation, rationale, and evidence data;
- private policy calibration and experiments;
- private instruction/profile packs and model configuration;
- private integration/deployment configuration;
- dogfood and operational material;
- proprietary or experimental extensions not intentionally promoted to the public core.

Credentials and secret values belong in neither repository.

## Go and process boundary

The public module is:

```text
github.com/hackelia-micrantha/calathea-community
```

The executable remains named `calathea`.

The Go implementation stays under `internal/`; those package paths are not a supported external library API. For v0, cross-repository composition uses the executable plus explicitly versioned file/schema/CLI contracts. A future exported Go facade requires a concrete in-process consumer and a separate compatibility decision.

The supported process surface and its current stable anchors are defined in the [CLI and process compatibility contract](docs/architecture/cli-process-contract.md). Automation must not treat human-readable CLI prose or internal Go identifiers as machine contracts.

## Product boundary

Calathea's reusable core is designed to remain:

- local-first and usable without network access;
- deterministic without requiring AI;
- human-approved rather than autonomously authoritative;
- read-only toward external repositories and project systems by default;
- independent of hosted accounts, mandatory telemetry, and cloud persistence;
- independent of Anthesis or another governance platform for core orientation.

Future effectful integrations require an explicit authorization/approval boundary. Anthesis may be one such adapter, but its concepts do not belong in the deterministic Calathea core.

## System context

```mermaid
flowchart LR
    User[Maintainer] --> CLI[Local CLI]
    CLI --> App[Application Services]
    App --> Core[Domain + Deterministic Services]
    App --> Store[(Private Local Store)]
    App -. optional read .-> Sources[Read-only External Sources]
    App -. optional invocation .-> AI[AI Provider]
```

Private user data is intentionally separate from the public source repository.

## Repository split

```mermaid
flowchart TB
    Private[calathea\nprivate composition + data] --> Public[calathea-community\npublic reusable core]
    Public --> Domain[Domain + orientation engine]
    Public --> CLI[CLI + application ports]
    Public --> Contracts[PRD + RFCs + ADRs]
    Private --> PrivateData[Portfolio/evidence/config]
    Private --> PrivateExt[Private experiments/extensions]
```

The dependency direction is private-to-public. The public core must not require the private repository to build, test, or run its documented core workflow.

## Development

Go is pinned through `mise.toml`. The local quality gate is:

```text
mise run check
```

It verifies formatting, runs `go vet` and `go test ./...`, and builds the `calathea` executable. The core has no third-party Go module dependencies at this stage.

## Current status

- repository ownership and MPL-2.0 licensing are established;
- the executable deterministic Go foundation is canonical here after public PR #7 and the coordinated private cleanup;
- reusable product/RFC/ADR/architecture contracts are canonical here after public PR #8 and the coordinated private cleanup;
- the public process boundary deliberately avoids exporting a broad Go library API;
- the CLI/process compatibility contract defines the current supported automation boundary;
- private portfolio and dogfood data remain outside this repository.

## Security and privacy

Public Calathea development must preserve these invariants:

- private portfolio data stays outside source checkouts by default;
- deterministic workflows can run with network access disabled;
- optional outbound integrations are explicit and scoped;
- imported repository content is untrusted data, not executable instruction;
- credentials are excluded from project records, prompts, evidence, traces, and model context;
- failures in optional integrations do not silently mutate canonical state.

## License

MPL-2.0. See [LICENSE](LICENSE).
