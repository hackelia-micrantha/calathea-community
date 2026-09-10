# Calathea Community Kernel Roadmap

## Status

Public implementation roadmap for the reusable kernel.

This is **not** the canonical Calathea product roadmap. Product sequencing, strategy, planning/review/lifecycle direction, AI paved-road choices, and dogfood priorities are owned by the private `hackelia-micrantha/calathea` repository.

## Goal

Deliver a small, independently useful, deterministic kernel with explicit process/file/schema contracts and enough local persistence/replay behavior to support reliable composition.

## Principle

Public work should follow demonstrated compatibility needs rather than trying to anticipate the complete private product architecture.

```text
concrete public behavior
        ↓
versioned contract
        ↓
implementation + synthetic tests
        ↓
compatibility/recovery evidence
```

## K0 — Repository boundary contraction

### Goal

Complete the transition from "public product specification" to "public reusable kernel".

### Work

- README and repository-boundary rewrite;
- product docs narrowed to kernel scope/use cases/roadmap;
- RFC governance narrowed to kernel behavior/compatibility;
- classify RFC 0000–0008 as retain/split/narrow/retire;
- audit ADR/architecture documents for product-strategy leakage;
- preserve public build/test and standalone usability.

Tracked by issue #48.

## K1 — Deterministic orientation foundation

### Goal

Provide reproducible orientation behavior through supported process contracts.

### Capabilities

- validation of public project/evaluation/policy inputs;
- deterministic evaluation/scoring behavior required by the public contract;
- bounded orientation queues;
- exclusions/diagnostics;
- stable tie-breaking;
- inspectable explanation/trace output;
- machine-readable output suitable for tests and automation.

### Exit criteria

- deterministic tests cover supported behavior;
- malformed/unsupported inputs fail explicitly;
- kernel runs without AI/network requirements;
- public fixtures are synthetic;
- automation does not depend on human-readable prose or internal Go identifiers.

## K2 — Local persistence and replay

### Goal

Provide reusable user-controlled persistence required by supported kernel workflows.

### Capabilities

- versioned persisted formats;
- immutable/versioned records where promised;
- rebuildable projections/materialized views;
- comparison/replay primitives;
- migrations;
- backup/restore/recovery behavior;
- crash/partial-write failure handling.

### Exit criteria

- clean-checkout tests exercise persistence and recovery using synthetic data;
- replay behavior is deterministic when required historical semantics are available;
- migrations have explicit compatibility tests;
- private data/infrastructure is unnecessary.

## K3 — Complete process compatibility surface

### Goal

Make cross-repository and script composition depend only on intentional public interfaces.

### Capabilities

- stable command/flag/input contracts where committed;
- documented exit/error semantics;
- versioned machine-readable output;
- file/schema compatibility policy;
- clear deprecation/migration rules;
- no accidental API promise for `internal/` Go packages.

### Exit criteria

- process-level tests exercise packaged behavior;
- breaking changes require explicit compatibility review;
- docs distinguish machine contracts from human output.

## K4 — Hardening and packaging

### Goal

Make the kernel straightforward to build, test, package, and validate independently.

### Capabilities

- reproducible local quality gate;
- static analysis and security checks;
- fuzz/property coverage where useful;
- packaging/install documentation;
- release/version metadata;
- corruption/recovery tests;
- privacy/no-network validation.

## K5 — Generic extensions only on demand

Potential later public work includes:

- a narrow exported Go facade if a concrete in-process consumer exists;
- generic source adapters when independent community use justifies them;
- generic structured AI/invocation adapters if they can be specified without private model/prompt policy;
- additional schemas or extension points with concrete compatibility needs.

Do not create a generic plugin platform, policy DSL, hosted service, or broad public product-management framework speculatively.

## Release gates

A public-kernel release should satisfy:

### Independence

- builds/tests without private repositories or private data;
- documented standalone workflow works from a clean checkout.

### Correctness

- promised deterministic behavior is reproducible;
- public schemas/process semantics are versioned;
- failures are explicit and covered.

### Compatibility

- machine-facing contracts are documented;
- migrations/deprecations are reviewed;
- internal implementation details are not accidentally promised as APIs.

### Security/privacy

- no secrets/private data in fixtures or traces;
- deterministic workflow can run without required network access;
- optional integrations are explicit and safely fail;
- imported external content remains untrusted data.

## Relationship to private product roadmap

The private product may create requirements for additional kernel capabilities. Those are not automatically added here wholesale.

The public roadmap records only the extracted implementation/compatibility work that this repository deliberately supports.
