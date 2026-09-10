# RFC 0004 — Optional AI Adapter Safety Contract

## Status

Accepted/narrowed public-kernel contract under issue #48.

The private Calathea product owns its AI paved road, model routing, instruction profiles, context-selection strategy, and workflow-specific governance. This RFC defines only the safety and compatibility constraints for any optional AI adapter exposed by the public kernel.

## Core invariant

AI is optional. Deterministic community-kernel workflows must remain usable without an AI provider, instruction system, hosted account, or network access.

AI/provider output is untrusted data. It does not become caller-authoritative kernel state, product authority, or external-effect authority merely because it is valid or high-confidence.

## Adapter requirements

A supported public AI adapter must provide, as applicable:

- explicit enablement/configuration;
- provider-neutral versioned request/result contracts;
- bounded caller-selected input/context;
- out-of-band credential resolution;
- provider/model identity and relevant invocation provenance;
- explicit timeout/cancellation/failure behavior;
- structured-output validation when machine-consumed;
- no canonical deterministic-state mutation on failed/invalid invocation.

## Input boundary

Imported repository/document/external text remains untrusted data.

It must not:

- select a more privileged configuration;
- widen source traversal;
- inject authoritative instruction implicitly;
- request broader capabilities by self-assertion;
- expose credentials or secret-derived values.

## Output boundary

Before optional AI output can influence a supported kernel workflow, the public adapter/application layer validates the documented contract.

Validation may include:

- schema/version checks;
- enum/range checks;
- required reference resolution;
- rejection of unsupported fields/versions where the contract requires it;
- preservation of provider/invocation identity.

Schema validity does not itself grant product-level authority.

## Failure semantics

Provider/instruction failure is an optional-feature failure.

The kernel must not silently:

- broaden context;
- switch to a more permissive authority model;
- mutate deterministic state;
- reinterpret invalid output as valid;
- claim successful invocation when provenance/evidence is incomplete.

A caller/product may implement explicit fallback policy outside this contract.

## Instruction interoperability

If a public instruction adapter such as Invokrum is implemented, follow the narrow architecture contract in `docs/architecture/invokrum-instruction-boundary.md`.

Exact instruction-artifact identity must not be falsely extended to transformed content or unrelated runtime/provider data.

## Effects

This RFC grants no external-effect authority. Any public effect adapter requires a separate explicit authorization/execution contract.

## Private product boundary

The following remain private by default:

- model/provider selection and routing;
- local/frontier escalation rules;
- prompt/instruction packs and profiles;
- product operation-to-profile mapping;
- detailed data-minimization policy by workflow;
- AI tool allowlists/denylists and rationale;
- review/disposition semantics;
- Anthesis integration strategy;
- dogfood-derived AI policy.

Only extract a generic public contract when an independent public consumer needs it.
