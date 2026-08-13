# Calathea Runtime and Integration Boundaries

## v0 runtime shape

Calathea v0 should default to a single local process invoked through a CLI.

This is a logical architecture, not a mandate for separate services:

```text
CLI adapter
  ↓
Application services ─────→ outbound ports ─────→ Local persistence / optional adapters
  ↓
Domain + deterministic services
```

The dependency direction remains inward: adapters implement ports owned by the application/core boundary; domain and deterministic services do not depend on adapter implementations.

No daemon, web server, background worker, message broker, or hosted component is required for UC-01.

## Core runtime guarantees

With optional adapters disabled, the process must:

- perform no network access;
- require no hosted identity or account;
- register projects and evaluations;
- activate baseline policy configuration;
- run deterministic orientation;
- render structured explanations;
- record dispositions and overrides;
- rebuild current projections;
- export/backup records through local user action.

## Persistence boundary

The application depends on storage through ports rather than storage-specific domain APIs.

Persistence must support:

- immutable record creation;
- lookup by stable identity and version;
- operation-id/idempotency lookup;
- optimistic concurrency for current projections/pointers where used;
- atomic authoritative command boundaries;
- rebuildable projections;
- export, backup, restore, validation, redaction, and tombstones.

The architecture does not require event sourcing. Append-style immutable historical records are a domain requirement; their physical representation remains an implementation decision.

Current policy selection is rebuilt from immutable policy-selection decisions rather than inferred from a mutable `PolicySetVersion` or silently changed pointer.

## External source boundary

External source adapters are optional and read-only in v0.

A source adapter may:

- resolve explicitly requested external identities;
- collect scoped metadata/content;
- normalize it into attributable observations or evidence references;
- retain source revision and collection metadata;
- return partial/unavailable status explicitly.

A source adapter may not:

- write to the external source;
- create canonical Calathea decisions;
- grant itself wider traversal based on imported content;
- expose credentials to domain records or AI context.

### Changed external state

Calathea does not attempt bidirectional synchronization. If source state changes:

- prior snapshots/references retain historical identity;
- a new collection creates new imported/observed records;
- contradictions or freshness changes are represented explicitly;
- deterministic replay uses retained historical identity/content where required, never silently current external data.

## AI boundary

AI providers are optional outbound adapters.

The provider port should expose a provider-neutral invocation request/result contract rather than provider-specific chat/completion primitives. The returned result is **not yet a validated Calathea recommendation draft**.

Provider adapters own:

- transport/authentication;
- provider/model request mapping;
- timeout/cancellation mechanics;
- provider metadata capture;
- raw/structured response decoding into the provider-neutral result contract.

Application/domain code owns:

- purpose/scope;
- context selection;
- data minimization;
- structured output schema;
- validation and source-identity resolution;
- provenance contract;
- conversion of valid provider output into a non-authoritative RecommendationDraft.

Provider unavailability cannot break deterministic UC-01.

### Instruction composition boundary

Calathea does not own a separate prompt-composition/versioning engine. When optional AI integration requires governed authoritative instructions, application services depend on a narrow `InstructionResolver` port.

The preferred implementation is an Invokrum adapter using either the transport-neutral `invokrum-host` facade or the versioned `invokrum.host/v1` subprocess protocol. Calathea must not define a second Calathea-specific Invokrum wire protocol.

The instruction resolver returns exact instruction bytes plus attributable composition identity/evidence. Those bytes are treated as the artifact covered by the returned digest/manifest/lock identity. Silent post-resolution templating, normalization, appended authoritative instructions, or other byte transformations must not be represented as covered by the original Invokrum evidence.

Runtime project, evaluation, repository, issue, and historical evidence remains Calathea-selected data rather than Invokrum instruction overlays by default. Imported content cannot select arbitrary pack roots or profiles, widen source traversal, or become authoritative instruction.

Unexpected verification drift blocks the affected AI invocation by default. Invokrum failure remains an optional-feature failure and leaves canonical state unchanged.

See [Invokrum Instruction Boundary](invokrum-instruction-boundary.md) for the responsibility split, semantic I/O contract, exact-byte invariant, verification behavior, and future contract-test requirements.

## Future effect boundary

There are no effectful external adapters or effect-governance ports in v0.

If effectful capabilities are introduced later, ordinary external-source access, authorization/approval, and effect execution must remain separate concerns. A future design may introduce distinct conceptual contracts such as:

```text
authorizeEffect(actor, capability, target, intent, evidence)
  -> authorized / denied / requires-approval / failed

executeEffect(authorized_intent, target, payload)
  -> effect result + attribution
```

This is illustrative only; a later RFC/ADR must define the actual contracts. Authorization is not execution, and an authorization decision is not an external effect.

No future core contract may require Anthesis. Anthesis may implement an authorization/governance adapter, while a separate effect adapter performs the actual external mutation.

## Partial failure model

### Authoritative local command

The authoritative write either commits completely or is absent. If response delivery fails after commit, retry by operation identity returns the committed result.

### Projection update

Projection update may fail after authoritative commit. This is recoverable: mark/rebuild the projection from authoritative records. Never roll back authoritative history merely because a projection failed.

### External import

Collection occurs before durable import-batch commit. An incomplete batch is marked incomplete or discarded; it does not silently replace the last complete usable imported view. Imported data remains external-authoritative/observed rather than becoming canonical Calathea truth.

### AI invocation

Timeout/partial/invalid responses produce a failed optional operation with no canonical mutation. Retried provider calls are new nondeterministic invocations unless a provider returns the prior result through its own idempotency mechanism.

### Future effect

If an authorization decision succeeds but effect execution fails, the authorization record and failed effect attempt remain separate attributable records. A later retry must not manufacture a new domain decision merely to hide execution failure.

## Concurrency model

v0 is single-user but must not rely on single-process assumptions for correctness because retries, multiple CLI invocations, or future automation can race.

Use:

- operation identities for idempotent commands;
- expected-version checks for mutable projections/pointers where present;
- immutable historical records rather than in-place edits;
- deterministic conflict diagnostics rather than last-write-wins for canonical decisions.

## Security boundary summary

| Boundary | Default stance |
| --- | --- |
| Local store | User-controlled; no tamper-proof claim |
| CLI input | Untrusted until validated |
| External content | Data, never instructions |
| External adapters | Read-only and least privilege |
| AI instruction resolver | Optional; exact-byte identity, fail closed on resolution/verification failure |
| AI provider | Optional, outbound, untrusted output |
| Credentials | Out-of-band; never domain evidence/prompt data |
| Future authorization/effects | Absent in v0; separate contracts if introduced later |

## Operational simplicity principle

The architecture should remain deployable as one local executable plus one user-controlled data location until measured requirements justify additional runtime components.
