# Invokrum Instruction Boundary

## Status

Accepted architecture refinement for future AI integration. Implementation remains deferred until Calathea implements an AI invocation path.

This decision refines RFC 0004 and the runtime integration boundary. It does not change the v0 requirement that deterministic portfolio orientation works without AI, Invokrum, Anthesis, or network access.

## Decision

Calathea will not build a separate prompt-composition/versioning subsystem.

Where governed AI instructions are required, Calathea should use Invokrum as the deterministic instruction-composition mechanism. Calathea retains ownership of domain-specific AI orchestration, runtime context/evidence selection, output contracts, validation, and recommendation-draft semantics.

Invokrum is an optional outer adapter/application dependency. It is not part of the deterministic orientation core.

## Responsibility split

| Concern | Owner |
| --- | --- |
| AI operation purpose and domain semantics | Calathea |
| Project/portfolio scope | Calathea |
| Runtime evidence/context selection | Calathea |
| Data minimization and outbound disclosure | Calathea |
| Prompt/instruction classes and profiles | Calathea-owned Invokrum pack |
| Deterministic overlay ordering and compatibility | Invokrum |
| Exact composed instruction bytes | Invokrum |
| Composition manifest, digest, and lock evidence | Invokrum |
| AI provider invocation | Calathea provider adapter |
| Structured output schema | Calathea |
| Model-output validation | Calathea |
| RecommendationDraft creation | Calathea |
| Capability/effect governance when introduced | Separate governance boundary; optionally Anthesis |

## Mental model

```text
Calathea AI application service
        |
        | operation/profile selection
        v
InstructionResolver port
        |
        +-- Invokrum adapter
        |      |
        |      +-- invokrum-host library facade, or
        |      +-- invokrum.host/v1 subprocess protocol
        |
        v
ResolvedInstructions
        |
        +-- exact instruction bytes
        +-- output digest
        +-- resolved manifest
        +-- canonical lock evidence
        |
        +------------------------+
                                 |
Calathea-selected runtime data  |
(projects/evidence/history)     |
        |                        |
        +------------+-----------+
                     v
              AI provider adapter
                     |
                     v
              untrusted output
                     |
                     v
              Calathea validation
                     |
                     v
             RecommendationDraft
```

Invokrum determines the identity and ordering of authoritative instruction material. Calathea determines which runtime facts and evidence are supplied alongside those instructions.

## Calathea-owned semantic contract

Calathea should own a narrow application port rather than expose Invokrum protocol DTOs throughout the application or domain model.

Conceptually:

```text
InstructionResolver.resolve(
  operation,
  instruction_profile,
  expected_lock?
) -> ResolvedInstructions

ResolvedInstructions:
  exact_bytes
  output_digest
  resolved_manifest
  canonical_lock
  resolver_compatibility
```

The exact application types remain an implementation decision. The contract expresses what Calathea needs, not a second wire protocol. The adapter may translate Invokrum transport DTOs into application-owned values, but it must not discard evidence needed to bind the resolved instruction artifact to an invocation.

## Wire and library integration

Calathea must not invent a Calathea-specific Invokrum transport protocol.

The adapter should use one of Invokrum's supported host surfaces:

1. `invokrum-host` when direct Rust embedding is practical; or
2. `invokrum.host/v1` through `invokrum rpc` when process isolation or implementation-language independence is preferable.

Transport choice is an adapter concern and must not leak into Calathea domain semantics.

Capability discovery should be used before relying on optional Invokrum host operations.

## Exact-byte invariant

The bytes returned by Invokrum are the instruction artifact covered by the returned digest/manifest/lock evidence.

Calathea must not silently claim that Invokrum evidence covers a transformed artifact.

If Calathea performs any post-resolution transformation, including:

- templating;
- appending instructions;
- normalization or newline conversion;
- provider-specific wrapping that changes represented instruction content;
- concatenation with other authoritative instruction text;

then the transformed value is a new artifact and requires its own attributable identity. The original Invokrum digest remains evidence only for the exact bytes Invokrum returned.

Provider transport serialization is a separate layer. JSON escaping, HTTP framing, or equivalent wire encoding does not make the entire provider request an Invokrum-attested artifact. The provider adapter must preserve the decoded instruction content represented by the resolved bytes and must not claim that the Invokrum digest covers transport envelopes, runtime evidence, or provider-added metadata.

The preferred design is to supply Invokrum-resolved content unchanged as the authoritative instruction layer and keep runtime portfolio/repository data logically separate.

## Runtime evidence boundary

Project metadata, evaluations, repository excerpts, issue text, prior orientation records, and other selected evidence are runtime data rather than Invokrum overlays by default.

This preserves the existing Calathea invariant that imported content is untrusted data, not instruction.

Runtime evidence must not:

- choose arbitrary Invokrum pack roots;
- select a more privileged instruction profile;
- add or replace instruction overlays;
- widen source traversal;
- authorize capabilities or effects.

Pack roots and profile mappings are selected by trusted Calathea configuration/application policy.

## Resolution and verification behavior

Two modes are useful:

### Resolve

Use `resolve` when no previously accepted instruction identity is pinned. The invocation records the returned instruction identity and evidence.

### Verify

Use `verify` when a workflow expects a pinned Invokrum lock/evidence identity.

Unexpected drift should block AI invocation by default. Drift is not a reason to silently generate a new accepted instruction identity.

A maintainer-controlled workflow may explicitly review and accept a new instruction identity separately.

## AIInvocation binding

When AI integration is implemented, `AIInvocation` should record enough information to establish which instructions were used without requiring indefinite retention of raw prompt/model content.

At minimum, where applicable:

- Calathea operation identity;
- Calathea instruction profile identity;
- Invokrum host/protocol or library compatibility version;
- Invokrum output digest;
- resolved manifest and canonical lock evidence, or durable integrity-preserving references to them;
- selected runtime context/evidence references or privacy-preserving digests;
- provider/model identity;
- structured-output contract version;
- validation result.

Raw instruction bytes need not be retained indefinitely when Calathea's retention policy permits an integrity-preserving durable reference to the exact artifact and its Invokrum evidence. Retention must still be sufficient for the reproducibility, audit, privacy, backup, and deletion guarantees Calathea claims.

## Failure semantics

Invokrum is optional infrastructure for optional AI assistance.

Failures must be fail-closed for the affected AI invocation and leave canonical Calathea state unchanged.

Examples:

| Failure | Calathea behavior |
| --- | --- |
| Pack/profile invalid | Reject AI invocation |
| Invokrum unavailable | Optional AI feature unavailable; deterministic core continues |
| Unsupported host protocol/capability | Reject invocation with explicit diagnostic |
| Resolution limit exceeded | Reject invocation; do not broaden limits silently |
| Verification drift | Block by default and require explicit review |
| Returned evidence malformed/inconsistent | Reject invocation |
| Provider fails after successful resolution | Record/link invocation failure according to retention policy; do not treat Invokrum resolution as model execution |

## Security and trust boundaries

Invokrum provides deterministic composition and integrity evidence. It does not authorize Calathea operations, authenticate the maintainer, approve outbound context, authorize tools/effects, or make model output trustworthy.

Calathea remains responsible for:

- authorizing the configured pack root and allowed profiles;
- context minimization and privacy controls;
- keeping credentials out of instruction/runtime evidence;
- binding the exact resolved instruction identity to the provider invocation;
- validating model output before constructing a RecommendationDraft.

Future Anthesis integration remains orthogonal: Anthesis may govern capabilities/effects, while Invokrum governs deterministic instruction composition.

## Compatibility and testing

When implementation begins, Calathea should add adapter-level contract tests that verify:

- capability negotiation;
- `resolve` success and exact-byte preservation;
- `verify` success and drift blocking;
- preservation of resolved manifest and canonical lock evidence;
- stable error mapping into Calathea diagnostics;
- rejection of malformed/unsupported responses;
- no post-resolution instruction mutation;
- provider mapping preserves the resolved instruction content while keeping transport/runtime data outside the Invokrum evidence claim;
- deterministic fixture behavior for a pinned Calathea Invokrum pack/profile.

Tests should target Invokrum's public host contract rather than implementation internals.

## Consequences

### Benefits

- avoids duplicate prompt composition/versioning machinery in Calathea;
- gives instruction sets deterministic identity and drift evidence;
- keeps prompt/instruction governance reusable across projects;
- preserves a clean distinction between trusted instructions and untrusted runtime evidence;
- keeps Calathea implementation-language choice independent of Invokrum through the subprocess contract;
- keeps optional AI out of the deterministic v0 core.

### Costs

- AI integration gains an additional adapter/dependency;
- exact-byte provenance requires discipline around provider request mapping;
- Calathea-owned instruction packs/profiles become versioned configuration requiring review;
- pinned verification introduces an explicit update/review workflow for instruction drift.

## Non-goals

This decision does not:

- introduce AI into UC-01;
- make Invokrum mandatory for deterministic Calathea operation;
- turn Invokrum into a runtime evidence templating engine;
- define a new Calathea/Invokrum wire protocol;
- authorize external effects;
- replace Anthesis or another future governance boundary;
- make model output deterministic or authoritative.
