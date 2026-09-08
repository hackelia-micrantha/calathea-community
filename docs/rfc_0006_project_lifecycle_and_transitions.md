# RFC 0006 — Lifecycle Compatibility Boundary

## Status

Deferred/narrowed public-kernel contract under issue #48.

The complete Calathea lifecycle, milestone, phase, archival, maintenance, and revival policy is private product semantics. This RFC records only compatibility invariants required by the public kernel.

## Core invariant

Orientation placement and lifecycle state are distinct.

In particular, a public `kill`/stop-investing orientation result does **not** itself:

- delete a project;
- archive a project;
- cancel work;
- transition a lifecycle field;
- perform an external mutation.

Any lifecycle transition requires a separate supported command/record if the public kernel exposes such a feature.

## Public lifecycle representation

A lifecycle field or transition becomes a public compatibility contract only when it appears in a supported CLI/file/schema/persistence surface.

If exposed, the contract must define:

- versioned allowed values;
- stable serialization;
- validation of unknown/unsupported values;
- transition behavior actually enforced by the kernel;
- migration consequences when values/semantics change;
- whether the field is caller-authored, derived, or a projection.

Do not infer additional private product semantics from a serialized enum.

## Transition safety

Where public transition commands exist:

- the current/source state is explicit;
- requested target state is explicit;
- invalid transitions fail visibly;
- retries/concurrency follow the public persistence/idempotency contract;
- historical transition records promised immutable are not rewritten;
- a deterministic orientation result cannot silently trigger the transition.

## Milestones and phases

The public kernel does not currently promise a general milestone/phase management model.

A future public representation should be introduced only for a concrete independently useful surface and should specify the minimum versioned fields/validation needed for interoperability.

The private product owns definitions of done, phase entry/exit policy, exception/waiver semantics, maintenance obligations, archive rationale, and revival criteria.

## External systems

Repository/issue/CI state may be used as imported evidence by a separately supported adapter, but external state cannot implicitly mutate a public lifecycle record unless an explicit future contract says so.

## Private product boundary

Private Calathea owns:

- lifecycle state taxonomy beyond exposed compatibility fields;
- legal product transition policy;
- milestone/phase semantics;
- definitions of done;
- pause/archive/complete/stop distinctions;
- maintenance/stakeholder policy;
- revival criteria;
- how lifecycle decisions interact with planning/review.

These remain private unless a concrete public compatibility need requires a deliberately extracted subset.
