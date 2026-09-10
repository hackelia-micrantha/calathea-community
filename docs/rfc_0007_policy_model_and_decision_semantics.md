# RFC 0007 — Policy Evaluation and Composition Compatibility

## Status

Accepted public-kernel contract.

The complete Calathea policy philosophy, portfolio defaults, calibration, weighting strategy, exception policy, and product decision semantics are private product concerns. This RFC defines only policy behavior that independent users or integrations must be able to rely on when using the public kernel.

## Scope

This RFC covers:

- immutable/versioned policy-set identity where exposed;
- deterministic evaluator identity and versioning;
- supported policy-result classes;
- deterministic composition and ordering;
- validation and failure behavior;
- traceability needed to reproduce public-kernel results;
- compatibility rules for evolving the exposed policy surface.

It does not define which policies the private Calathea product should use or how they should be calibrated.

## Public contract

A supported public policy configuration is data interpreted by known kernel evaluators. Configuration must not embed arbitrary executable code or become an unrestricted expression/automation language.

Where policy is exposed through CLI, file, schema, or persistence surfaces, a policy instance identifies at least:

- stable policy/instance identity;
- evaluator type and semantic version;
- configuration/schema version;
- declared parameters required by that evaluator;
- deterministic ordering information where multiple policies compose.

Unknown required evaluator versions or invalid configuration fail visibly. The kernel must not silently substitute a different evaluator or reinterpret an older configuration under new semantics.

## Result classes

Public evaluators may emit typed results such as:

- `allow`;
- `deny`;
- `adjust`;
- `require_review` where a supported workflow uses it;
- `not_applicable`;
- `indeterminate`.

`indeterminate` and evaluator failure are distinct. Missing required evidence must not silently become `allow`.

An `allow` from one policy does not implicitly erase an applicable `deny` from another. A soft adjustment cannot make an otherwise invalid result legal.

## Deterministic composition

For identical validated inputs, evaluator versions, configuration, and kernel semantic versions, policy evaluation and composition must produce equivalent results and trace ordering.

When ordering affects results, ordering is explicit and stable. Priority may establish execution/composition order but must not silently redefine the semantic strength of unrelated result classes.

Numeric adjustments that affect deterministic output use exact/defined arithmetic semantics rather than implementation-dependent floating-point behavior.

If two effects cannot be composed under a documented combinator, the kernel reports a conflict/failure rather than choosing arbitrarily.

## Invariants versus policy

Kernel correctness invariants are not configurable policies. A policy cannot disable invariants promised by another public contract, including record immutability, output cardinality/shape constraints, or explicit separation between a recommendation and an external side effect.

Configuration that attempts to violate a kernel invariant is invalid.

## Exceptions and overrides

Exception/override records become part of this public contract only when a supported public workflow exposes them.

If exposed, they must be:

- explicit rather than implicit bypasses;
- attributable;
- scoped to a concrete policy/result/subject;
- separately recorded from the immutable policy-set version and original deterministic result;
- validated against non-exceptionable kernel invariants.

The private product owns the broader policy for when exceptions are appropriate, who should grant them, and how they affect planning or governance.

## Trace and replay

A public operation whose output is affected by policy must retain or reference enough information to identify:

- policy-set/configuration version;
- evaluator identities and versions;
- material policy results;
- composition/order information when relevant;
- deterministic input identities required by the operation.

The public trace contract is defined further by RFC 0008 and persistence/replay behavior by RFC 0005.

## Compatibility

A change requires explicit schema/semantic versioning or migration when it alters any supported public behavior such as:

- evaluator meaning;
- result classification;
- arithmetic or ordering semantics;
- configuration interpretation;
- composition/conflict behavior;
- persisted policy/result representation.

Adding a private Calathea policy or changing private calibration does not by itself create a public compatibility commitment.

## Security boundary

Public policy configuration is untrusted until validated. It must not:

- execute arbitrary commands or scripts;
- initiate undeclared network access;
- promote imported text or AI output into executable policy code;
- contain credentials or secret values;
- authorize external effects merely because a policy evaluation returned an affirmative result.

Any future effect authorization/execution surface requires its own explicit public contract.

## Private product boundary

Private `hackelia-micrantha/calathea` owns:

- policy philosophy and governance;
- default portfolio policies;
- weight/threshold calibration;
- prioritization heuristics and experiments;
- review/lifecycle policy interactions;
- private exception/approval strategy;
- product-specific interpretation of policy results;
- automatic-learning or adaptation strategy, if ever adopted.

Those decisions should be published only when a concrete community compatibility or standalone-use requirement justifies extracting a minimum generic contract.
