# Optional Invokrum Interoperability Boundary

## Status

Deferred public adapter contract.

There is currently no requirement for Invokrum in deterministic `calathea-community` operation. This document records only the minimum interoperability constraints that would apply if a concrete public Invokrum adapter is implemented.

The private Calathea product owns model-routing, prompt/profile selection, AI paved-road policy, product-specific context assembly, and the decision to use Invokrum for particular workflows.

## Public-kernel rule

The kernel must not build a second Invokrum protocol or expose Invokrum implementation details through its domain model.

If a public adapter is added, it should depend on a narrow application-owned abstraction and translate to an intentionally supported Invokrum host surface.

Conceptually:

```text
InstructionResolver.resolve(request) -> ResolvedInstructions
```

A public adapter contract may include only the information required to preserve interoperability and attribution, for example:

- exact resolved instruction bytes;
- instruction artifact digest/identity;
- manifest/lock evidence required by the supported Invokrum contract;
- compatibility/protocol version;
- explicit resolution/verification failure.

The exact public types should be introduced only with a concrete implementation and compatibility need.

## Exact-artifact invariant

If an Invokrum result claims a digest/identity for instruction bytes, that evidence covers exactly the represented instruction artifact—not arbitrary post-processing, runtime context, provider transport envelopes, or later appended instruction text.

Any semantic transformation that changes the authoritative instruction content creates a distinct artifact and must not be represented as covered by the original Invokrum identity.

## Runtime data boundary

External/project data supplied alongside instructions remains data, not automatically trusted instruction material.

A public adapter must not allow untrusted runtime data to:

- choose arbitrary instruction roots/profiles;
- widen source traversal;
- request broader capabilities;
- authorize external effects;
- rewrite kernel-owned compatibility state.

Product-specific rules for selecting profiles/context belong outside this public kernel contract.

## Failure semantics

Invokrum is an optional integration.

When enabled, resolution/verification failure should fail closed for that optional operation and leave deterministic kernel state unchanged.

Examples include:

- unsupported protocol/capability;
- invalid pack/profile;
- verification drift;
- malformed/inconsistent evidence;
- resolution limit failure.

## Security boundary

Invokrum composition evidence does not authenticate a maintainer, authorize a Calathea product decision, approve outbound context, authorize tools/effects, or make model output authoritative.

Those concerns remain separate from this interoperability boundary.

## Testing requirement if implemented

A public adapter should have synthetic contract tests for:

- supported capability negotiation;
- exact-byte preservation;
- resolution success/failure mapping;
- verification drift blocking where supported;
- evidence preservation;
- malformed/unsupported response rejection;
- no credential leakage;
- no dependency on private Calathea prompt packs or portfolio data.

## Non-goals

This document does not:

- make Invokrum mandatory;
- select Invokrum as the private product's AI strategy;
- publish private instruction/profile packs;
- define private model routing or escalation;
- define complete Calathea AI workflow semantics;
- authorize external effects.

A richer public contract should be added only after a concrete independently useful adapter exists.
