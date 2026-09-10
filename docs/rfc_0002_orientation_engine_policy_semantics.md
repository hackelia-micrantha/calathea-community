# RFC 0002 — Deterministic Orientation Contract

## Status

Accepted public-kernel contract, narrowed under issue #48.

The private Calathea product owns portfolio strategy, capacity defaults, stop-investing philosophy, and policy calibration. This RFC defines the deterministic orientation behavior exposed by the public kernel.

## Supported placements

Where exposed by the public schema/process contract, orientation output may use:

- `now`;
- `next`;
- `later`;
- `kill` / stop-investing.

These values are recommendation/output labels. They do not themselves mutate project lifecycle, delete data, archive a project, or perform an external effect.

## Inputs

A deterministic orientation request identifies exact supported versions of:

- candidate/project inputs;
- evaluation inputs;
- policy/evaluator inputs;
- queue bounds and other configuration that materially affects output;
- semantic version of orientation behavior.

Hidden mutable state, network data, AI output, or wall-clock state must not influence a deterministic result unless explicitly represented by a versioned input.

## Eligibility and exclusions

The kernel must explicitly distinguish candidates that are eligible from those excluded by supported rules.

Every material exclusion represented by the public output contract should have a stable machine-readable reason/diagnostic.

The kernel must not silently relax a hard constraint merely to fill a queue.

## Bounded queues

If `now` or `next` is bounded, the effective bound is an explicit input or documented versioned default and is visible in output/trace where required by the public contract.

The public kernel does not define the private product's preferred production values.

## Tie-breaking

Tie-breaking is deterministic and based on stable documented values. It must not depend on map iteration, nondeterministic provider output, process timing, or unspecified ordering.

## Traceability

A supported orientation result should expose enough structured information to inspect material behavior, including as applicable:

- input/semantic versions;
- eligibility/exclusion;
- base evaluation result;
- policy/evaluator effects;
- queue bounds;
- final placement;
- tie-break reason;
- stable diagnostics.

Human-readable wording is not a machine compatibility surface unless explicitly documented.

## Dispositions and overrides

If the public kernel exposes persisted dispositions/overrides, their schema and validation are separate from the immutable OrientationRun/result they reference.

Changing a caller disposition must not rewrite the underlying deterministic run.

The private product owns richer authority and workflow semantics around acceptance/rejection/deferral unless they are intentionally exposed by a public process contract.

## Replay

Reproducing an orientation requires the exact versioned deterministic inputs and semantic behavior. If a required version is unavailable, replay fails explicitly rather than substituting current policy/input data.

## Private product boundary

The following remain private by default:

- strategic queue-size defaults;
- portfolio prioritization philosophy;
- stop-investing thresholds/rationale;
- production weights/penalties/bonuses;
- policy calibration and dogfood-derived heuristics;
- interpretation of placements as product-management decisions beyond the public output contract.
