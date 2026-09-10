# Public/private repository boundary

## Status

Accepted public-kernel ownership rule.

This document supersedes the earlier assumption that `calathea-community` should own the complete reusable Calathea product/domain specification.

The public repository remains the canonical implementation/compatibility home for the kernel it exposes. The private `hackelia-micrantha/calathea` repository is canonical for the complete Calathea product strategy and product-management semantics.

## Repositories

- `hackelia-micrantha/calathea-community` — public reusable deterministic kernel and compatibility surface.
- `hackelia-micrantha/calathea` — private Calathea product, strategy, composition, dogfood, and private extensions.

The split is an ownership boundary, not a periodic mirror and not two independently implemented editions.

## Public target ownership

`calathea-community` owns:

- deterministic reusable implementation;
- generic application services and application-owned ports used by the kernel;
- the public `calathea` executable;
- supported CLI/process/file/schema behavior and compatibility;
- reusable local persistence, migrations, fixtures, golden tests, and conformance tests;
- kernel-level safety/privacy invariants necessary for standalone use;
- generic extension and interoperability boundaries;
- implementation ADRs and RFCs required to preserve public behavior or compatibility;
- synthetic examples, packaging, CI, and contributor documentation.

It does **not** own the complete Calathea product PRD, strategy, roadmap, full use-case model, product policy philosophy, calibration, AI paved road, dogfood, or private operational semantics.

## Private product ownership

The private `calathea` repository owns:

- product PRD, goals, outcomes, metrics, and roadmap;
- complete end-to-end project-management use cases;
- prioritization/planning/review/lifecycle/milestone/archival/stakeholder policy;
- definitions of done and product phase policy;
- policy calibration, weights, heuristics, experiments, and product defaults;
- AI-assisted workflow strategy, model routing, private prompt/profile packs, and paved-road decisions;
- real portfolio/project/evaluation/evidence data;
- private deployment/integration configuration and dogfood;
- proprietary or experimental extensions not deliberately extracted into a public contract.

Credentials and secret values belong in neither repository.

## Dependency direction

The implementation dependency remains:

```text
calathea (private product / composition)
                |
                | process / files / versioned contracts
                v
calathea-community (public reusable kernel)
```

The public kernel must not depend on the private repository for build, test, packaging, or documented standalone workflows.

Private composition pins a reviewed public-kernel revision.

## Authority direction

Product/specification authority is separate from implementation dependency.

The private product may require a capability from the public kernel. That requirement does not make the complete product rationale, policy, workflow, or calibration a public contract.

Public documentation should expose the minimum mechanism an independent consumer must know to use, implement, test, or extend the supported kernel safely.

## Publication rule

A concept belongs in this repository when an independent community consumer or implementation must know it for at least one of:

1. compatibility;
2. safety/privacy of the public kernel;
3. contribution or maintenance of the public implementation;
4. useful standalone operation;
5. an intentionally supported extension/interoperability surface.

Otherwise it defaults to private product documentation.

Generic applicability or potential reuse is not sufficient by itself.

## Mechanism versus policy

Prefer the following boundary:

| Private product | Public kernel |
| --- | --- |
| Why/when Calathea uses a capability | Supported interface and deterministic behavior |
| Strategic defaults and calibration | Validation and error semantics |
| Complete workflow and policy | Versioning and compatibility |
| Dogfood-derived heuristics | Synthetic fixtures/conformance tests |
| Model routing/paved-road choices | Generic adapter/invocation contract |
| Product roadmap | Public kernel implementation roadmap |

## Go module and process boundary

The public module remains:

```text
github.com/hackelia-micrantha/calathea-community
```

The user-facing binary remains `calathea`.

The reusable Go implementation remains primarily under `internal/`. Repository topology is not sufficient reason to expose a broad Go library API.

For v0, the supported cross-repository and automation surface is:

- the `calathea` executable;
- explicitly documented CLI behavior and exit semantics;
- machine-readable output when explicitly versioned;
- versioned file/schema contracts as introduced.

Private composition must not import or vendor `calathea-community/internal/...` packages.

A future exported Go facade requires a concrete in-process consumer and separate compatibility review. See [ADR 0005](../adr/0005_public_go_process_boundary.md).

## Source-of-truth rule

There is exactly one canonical owner for a maintained decision:

- public kernel implementation/compatibility: this repository;
- complete Calathea product strategy/policy: private `calathea`;
- private portfolio data/dogfood/configuration: private `calathea`;
- public tests/examples: this repository and synthetic only.

Do not maintain equivalent canonical product documents in both repositories.

If a private product requirement requires a new kernel capability:

1. define the product need privately;
2. extract the minimum generic public interface/behavior;
3. review and implement it here;
4. validate it independently with synthetic tests;
5. let the private product consume a reviewed public revision.

## RFC/ADR source of truth

Public RFCs own durable kernel semantics and compatibility commitments. They should not be used as a default home for full product philosophy, strategy, or policy.

RFC 0000–0008 were narrowed in place under issue #48 according to public compatibility need while preserving historical filenames and git history.

Public ADRs remain appropriate for implementation choices made by this repository.

## Private-data boundary

Public-kernel development must preserve:

- user/private portfolio data outside source checkout by default;
- deterministic kernel operation without mandatory network access;
- no required hosted account or telemetry;
- optional external/AI integration that is explicit and scoped;
- credentials resolved out of band and excluded from durable records/prompts/evidence;
- imported content treated as untrusted data;
- synthetic public fixtures that do not derive from private data.

## Migration status

### Original extraction

The executable/core extraction remains valid:

- public Go module and `calathea` executable exist;
- reusable deterministic implementation lives here;
- process/file/schema boundary avoids speculative broad Go API exposure;
- private duplicate implementation is not required.

### Boundary contraction

Tracked by issue #48 and private `calathea#59` and implemented by PR #50.

This branch:

- repositions the README around kernel scope;
- narrows public product PRD/use-case/roadmap text to kernel-scoped documents;
- changes RFC governance from full product semantics to kernel/compatibility semantics;
- contracts RFC 0000–0008 in place;
- audits and narrows ADR/architecture docs that exposed private-product strategy;
- intentionally leaves public build/test/module/process behavior unchanged.

Historical public product text remains available in git history; history is not rewritten to simulate that the earlier boundary never existed.

## CI and release expectations

`calathea-community` remains independently buildable and testable from a clean checkout without access to the private repository, private data, private infrastructure contracts, or private product documentation.

The private product may run additional integration/dogfood validation against a pinned public-kernel revision.

## Licensing

Public source is MPL-2.0 unless a file or dependency states otherwise.

The private product repository is not made public or relicensed merely because it depends on this kernel. Changes to MPL-covered public implementation files remain subject to their public license terms.

## Non-goals

This split does not:

- make the public kernel a toy or documentation-only project;
- create duplicated public/private implementations;
- hide behavior required to safely use the public kernel;
- publish private portfolio data or product strategy;
- require Anthesis, Invokrum, an AI provider, or a hosted service for deterministic kernel operation.
