# Community Kernel Runtime and Integration Boundaries

## Scope

This document defines runtime boundaries required by the public `calathea-community` kernel.

It does not define the private Calathea product's complete AI, governance, planning, review, lifecycle, or effect architecture.

## Default runtime shape

The public kernel defaults to a single local process invoked through the `calathea` CLI.

```text
CLI adapter
  ↓
Application services ─────→ outbound ports ─────→ local persistence / optional supported adapters
  ↓
Domain + deterministic services
```

The dependency direction remains inward: adapters implement ports owned by the application/core boundary; deterministic/domain behavior does not depend on adapter implementations.

No daemon, web server, background worker, broker, hosted control plane, AI provider, or private Calathea service is required for the deterministic public workflow.

## Core runtime guarantees

With optional adapters disabled, the supported kernel should:

- perform no required network access;
- require no hosted identity/account;
- validate supported project/evaluation/policy inputs;
- run deterministic orientation;
- render documented explanation/machine output;
- use public persistence/projection behavior where implemented;
- export/backup supported records through documented local workflows.

The exact product policy selecting or interpreting these mechanisms is outside this public runtime contract.

## Persistence boundary

Application services depend on persistence through inward-owned ports rather than storage-specific domain APIs.

Where the public persistence contract requires them, implementations should support:

- stable identity/version lookup;
- immutable or versioned record creation;
- operation-id/idempotency lookup;
- optimistic concurrency for replaceable projections/pointers where used;
- atomic authoritative command boundaries;
- rebuildable projections;
- versioned migrations;
- documented export, backup, restore, validation, and recovery behavior.

The architecture does not require event sourcing or a particular database. Physical representation remains an implementation decision unless explicitly exposed by a public format contract.

## Optional external-source boundary

External source adapters are not part of the deterministic-kernel requirement.

When a concrete public adapter is supported, it may:

- resolve explicitly requested external identities;
- collect only the documented scope;
- normalize source data into attributable public adapter values;
- retain source revision/collection metadata required by the contract;
- report partial/unavailable state explicitly.

It may not:

- widen its own access based on imported content;
- treat imported text as executable instruction;
- expose credentials in durable kernel records or fixtures;
- silently convert external data into authoritative kernel decisions.

Adapter-specific product policy remains outside this public contract.

## Optional AI/instruction boundary

AI is not required for the public deterministic kernel.

No provider, prompt framework, instruction resolver, model-routing policy, or governance product is a mandatory runtime dependency.

If a generic AI/instruction adapter is introduced publicly, its contract must be narrow and provider-neutral enough for independent use. At minimum it must preserve:

- explicit enablement;
- out-of-band credential handling;
- bounded input/context;
- attributable provider/model/instruction identity where claimed;
- validation of structured output before kernel workflow use;
- fail-closed behavior for the affected optional operation;
- no authority escalation from imported/model output.

Product-specific prompt packs, model selection, routing/escalation, review policy, and paved-road choices remain private unless deliberately published as a separate reusable project.

See [optional instruction interoperability](invokrum-instruction-boundary.md) and [structured invocation note](structured-invocation-contract.md) for the deliberately narrow public status of those surfaces.

## External effects

The public kernel has no ambient authority to mutate repositories or other project systems.

If a concrete public effect adapter is ever introduced, authorization/approval and effect execution must be separate, explicitly supported contracts. Do not speculatively add product-governance machinery to the deterministic core.

No future public-kernel contract may require Anthesis or another specific governance product unless that dependency is deliberately accepted as part of a separately supported adapter.

## Partial failure model

### Authoritative local command

The command either commits the records promised by its public contract or reports failure without claiming ambiguous partial authoritative state. If response delivery fails after commit, documented idempotency/operation identity should allow safe retry where supported.

### Projection update

A projection may be rebuilt after failure from the authoritative/versioned records defined by the public persistence contract. Projection failure must not silently rewrite history.

### Optional import

Incomplete/failed collection must not silently replace a previously valid imported view when the adapter contract promises retained state. Partial status must be explicit.

### Optional AI/instruction invocation

Timeout, invalid output, provider failure, or instruction-resolution failure is an optional-feature failure. It must not corrupt deterministic kernel state.

## Concurrency model

Even a local CLI should not rely on single-process assumptions for correctness when retries or multiple invocations can race.

Use as required by the public contract:

- operation identities for idempotent commands;
- expected-version checks for replaceable projections/pointers;
- versioned/immutable historical records rather than silent in-place history edits;
- deterministic conflict diagnostics rather than accidental last-write-wins.

## Security boundary summary

| Boundary | Public default stance |
| --- | --- |
| Local store | User-controlled; no tamper-proof claim |
| CLI/file input | Untrusted until validated |
| External content | Data, never implicit instruction |
| Optional adapters | Explicit, least-privilege, scoped |
| AI/instruction adapters | Optional; output non-authoritative until validated |
| Credentials | Out-of-band; excluded from durable kernel content/fixtures |
| External effects | Absent from ambient deterministic-kernel authority |

## Operational simplicity

Keep the public kernel deployable as one local executable plus user-controlled data until a concrete independent use case justifies additional runtime components.

Private Calathea may compose additional systems around the kernel without making those systems part of this public runtime contract.
