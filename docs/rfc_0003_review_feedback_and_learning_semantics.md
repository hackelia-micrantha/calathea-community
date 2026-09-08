# RFC 0003 — Public Review Extension Boundary

## Status

Deferred/narrowed public-kernel contract under issue #48.

The complete Calathea review, feedback, drift, and learning model is private product semantics. The public kernel does not currently promise a general review-management framework.

## Decision

Do not expose the private review taxonomy merely because generic record/persistence mechanisms could represent it.

A future public review feature must start from a concrete independently useful process/file/schema contract and extract only the minimum interoperable records and behavior.

## Current public guarantees

Existing deterministic/persistence surfaces may support generic properties useful to later review extensions:

- stable/versioned record identity;
- evidence/provenance references;
- immutable/versioned history where promised;
- rebuildable projections;
- machine-readable diagnostics/traces;
- safe failure without silent canonical mutation.

Those primitives do not constitute a public `ReviewCycle`, `Finding`, calibration, or learning API.

## Future public review records

If a concrete public use case is accepted, any exposed review records should be:

- explicitly versioned;
- attributable to caller/source/time;
- separable into observed/recommended versus caller-accepted state as needed by the interface;
- evidence-linked;
- migration-compatible;
- testable with synthetic fixtures.

The public RFC must state exactly which fields and transitions are compatibility promises.

## Learning boundary

The public kernel must not silently learn or mutate evaluation weights, policy, heuristics, or defaults from outcomes, model output, or prior caller behavior.

A future public calibration feature would require its own explicit contract.

## AI boundary

AI-based review synthesis is optional and not part of deterministic kernel requirements. A generic adapter, if introduced, is governed by RFC 0004's public safety constraints; the private product owns review prompting, taxonomy, routing, and disposition semantics.

## Private product boundary

Private Calathea owns:

- review triggers/cadence;
- observations/findings/recommendations/dispositions;
- drift taxonomy;
- implementation-to-planning feedback;
- calibration observations and policy-change workflow;
- stakeholder/project review strategy;
- AI-assisted review orchestration.

These concepts become public only when a concrete standalone kernel surface requires a deliberately extracted subset.
