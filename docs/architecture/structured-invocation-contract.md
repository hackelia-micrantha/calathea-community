# Optional Structured Invocation Interoperability

## Status

Deferred public extension note.

The deterministic community kernel does not require an AI runtime or a portable structured-invocation protocol.

This document records only the compatibility properties a future public adapter should preserve. The private Calathea product owns its complete workflow-state model, context selection, roles, model routing, review semantics, and AI paved-road policy.

## Public interoperability principle

A future structured invocation surface should keep three categories separate:

```text
caller-selected structured input
        ↓
optional instruction/provider adapter
        ↓
untrusted runtime output
```

The public kernel should not treat rendered prompt text, provider-private working state, or mutable chat history as an implicit canonical extension of kernel state.

## Minimal candidate contract

A concrete public adapter may eventually need a small versioned request containing only generic interoperability data, such as:

```yaml
invocation:
  operation: example.operation
  input_revision: example-revision
  selected_input_refs: []
  instruction_artifact_ref: optional
  constraints: {}
  output_contract: example/v1
```

These fields are illustrative and are **not** a committed schema.

A real public schema should be introduced only with an independently useful implementation and conformance tests.

## Revision binding

If an invocation claims to correspond to a specific versioned kernel input, the result must remain attributable to that input revision.

If material input changes before execution/review, the caller or adapter must not silently rebind the old invocation to the newer state while preserving the old identity.

The exact stale/retry policy is a caller/product concern unless explicitly standardized by the public adapter.

## Data versus instruction

Selected external excerpts, repository text, issue content, retrieved context, and similar inputs remain data unless deliberately introduced through a trusted instruction/configuration path.

Untrusted input must not:

- select a more privileged instruction configuration;
- widen requested capabilities;
- self-authorize external effects;
- become executable instruction merely by appearing in context.

## Runtime-private state

Provider/runtime intermediate computation is not automatically part of the public kernel state or audit contract.

A public adapter must not claim replayability of hidden provider state it cannot actually reproduce.

## Output boundary

Runtime/model output is untrusted until it satisfies the documented output contract and caller-side validation.

Validation does not itself grant product-level authority or permission for external effects.

## Security and failure properties

A future public adapter should demonstrate:

- explicit enablement;
- bounded inputs;
- versioned request/output contracts;
- out-of-band credential handling;
- attribution to claimed input/instruction/provider identities;
- stale/retry behavior that does not silently falsify provenance;
- invalid/partial output rejection;
- no mutation of deterministic kernel state on optional-adapter failure;
- no reliance on private Calathea data, prompts, profiles, or runtime topology in public tests.

## Relationship to Invokrum

If Invokrum is used by a concrete public adapter, follow the narrow [Invokrum interoperability boundary](invokrum-instruction-boundary.md). Do not create a second Calathea-specific Invokrum protocol.

The private product may use a richer structured invocation model. That model is not automatically a public compatibility commitment.

## Non-goals

This note does not define:

- the private Calathea AI workflow model;
- supervisor/specialist runtime architecture;
- product roles or requested capabilities;
- product context/evidence selection policy;
- model routing or escalation policy;
- feedback/learning strategy;
- authorization/effect governance;
- a committed portable invocation schema.

Promote only the minimum contract once a concrete public consumer exists.
