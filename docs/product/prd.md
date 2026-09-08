# Calathea Community Kernel Scope

## Status

Accepted public-kernel scope.

This is **not** the complete Calathea product PRD. The private `hackelia-micrantha/calathea` repository owns product strategy, complete use cases, roadmap, policy, and dogfood.

This document defines only what the public community kernel must provide as an independently useful and supportable artifact.

## Purpose

`calathea-community` provides deterministic project-orientation mechanisms and stable integration contracts that can be used standalone or composed by a larger product.

The public kernel should make these properties reliable:

- deterministic behavior for versioned deterministic inputs;
- explicit validation and error semantics;
- inspectable recommendation traces;
- stable automation-facing process/file/schema contracts;
- local operation without mandatory network access;
- user-controlled persistence and replay where supported;
- synthetic tests and examples sufficient to validate compatibility.

## Primary public user

An independent technical maintainer, integrator, or contributor who wants to:

- run the `calathea` executable locally;
- provide valid project/evaluation/policy inputs;
- obtain deterministic orientation output;
- inspect why a placement or exclusion occurred;
- persist/replay supported records;
- automate against explicitly versioned machine contracts;
- contribute to or extend the kernel without access to private Calathea product material.

## Public-kernel responsibilities

### Deterministic orientation mechanisms

The kernel may expose reusable mechanisms for:

- validated project/evaluation inputs;
- deterministic scoring/evaluation;
- bounded orientation queues;
- exclusions and diagnostics;
- stable tie-breaking;
- policy/evaluator behavior that is part of the supported public interface;
- machine-readable traces/explanations.

Product-specific calibration and strategic policy remain outside this scope.

### Process and compatibility surface

The kernel owns:

- documented CLI commands and exit behavior;
- explicitly supported machine-readable output;
- versioned file/schema contracts;
- migrations and compatibility rules for public persisted formats;
- failure behavior required for safe automation.

Human-readable prose and internal Go identifiers are not machine contracts unless explicitly documented as such.

### Persistence and replay

Where public persistence is implemented, the kernel owns reusable behavior for:

- local storage;
- immutable or versioned records as specified by the public contract;
- rebuildable projections/materialized views;
- migrations;
- backup/restore/recovery semantics required by the supported format;
- deterministic replay when required inputs and semantic versions are available.

### Safety and privacy

The public kernel must:

- work without mandatory hosted services;
- avoid required telemetry;
- support deterministic workflows with network access disabled;
- keep credentials outside project records, prompts, evidence, traces, and fixtures;
- treat imported external text as untrusted data;
- fail optional integration paths without silently mutating canonical kernel state;
- use synthetic public fixtures rather than private portfolio data.

## Public-kernel non-goals

This repository does not define or publish the complete Calathea:

- product strategy or business/product outcomes;
- strategic roadmap and sequencing;
- full project-management use-case taxonomy;
- prioritization philosophy or production calibration;
- planning/review/lifecycle/milestone/archival policy;
- stakeholder and long-term maintenance strategy;
- AI paved-road/model-routing/private prompt strategy;
- dogfood evidence, private operating assumptions, or portfolio data.

It also does not aim to become a complete replacement for issue trackers, CI/CD, source control, or autonomous implementation systems.

## Supported integration boundary

The public Go module is:

```text
github.com/hackelia-micrantha/calathea-community
```

The supported user-facing executable is:

```text
calathea
```

The Go implementation remains primarily under `internal/`; those packages are not a compatibility promise.

For v0, composition is process-oriented through documented CLI/file/schema contracts. An exported Go facade requires a concrete consumer and separate compatibility decision.

## Public acceptance requirements

A release-quality kernel should demonstrate:

1. a clean checkout can build/test without the private repository;
2. deterministic workflows run without AI or external network access;
3. identical versioned deterministic inputs produce identical promised outputs;
4. malformed or unsupported inputs fail explicitly;
5. machine-readable automation contracts are versioned and documented;
6. public fixtures are synthetic;
7. no private data/configuration is needed for tests or examples;
8. optional integrations cannot silently corrupt deterministic state;
9. compatibility/migration consequences are documented when a public contract changes.

## Relationship to private Calathea

The private Calathea product may select, compose, and calibrate these mechanisms for richer project-management workflows.

That product dependency does not make private strategy part of this public contract.

When a private product requirement needs a new reusable capability, only the minimum generic interface and behavior required by independent consumers should be promoted here.

See [repository boundary](../architecture/repository-boundary.md), [kernel use cases](use-cases.md), and [kernel roadmap](mvp-roadmap.md).
