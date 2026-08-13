# RFC 0001 — Evaluation Semantics

## Status

Accepted

- **Owner:** Calathea maintainers
- **Decision date:** 2026-08-03
- **Depends on:** RFC 0000, RFC 0005, PRD, UC-01
- **Scope:** Evaluation axes, scoring, confidence, freshness inputs, and calibration

## Summary

RFC 0001 defines how a project is assessed. It does not define portfolio placement, queue policy, lifecycle, or acceptance of orientation recommendations.

An accepted assessment is an immutable `EvaluationVersion`. Drafts produced by a maintainer or AI assistant remain non-authoritative until explicitly accepted.

## Evaluation axes

Each evaluation records integer values from 1 through 5 for:

- impact;
- effort;
- risk reduction;
- optionality;
- urgency.

Each axis must include a rationale. Scores without rationale are invalid.

## Calibration anchors

### Impact

- `1`: narrow or speculative benefit;
- `3`: material project or workflow benefit;
- `5`: portfolio-defining or strategically critical benefit.

### Effort

- `1`: small, bounded change with low coordination cost;
- `3`: multi-step work with meaningful uncertainty;
- `5`: large, cross-cutting, or poorly bounded work.

Effort defaults upward when uncertainty is material.

### Risk reduction

- `1`: little measurable reduction in operational, security, delivery, or governance risk;
- `3`: addresses a known material risk;
- `5`: removes or materially controls a critical risk.

### Optionality

- `1`: creates little reusable leverage;
- `3`: enables at least one concrete follow-on path;
- `5`: unlocks several credible strategic paths.

Optionality requires named future paths; vague flexibility does not qualify.

### Urgency

- `1`: delay has little consequence within the planning horizon;
- `3`: delay creates measurable cost or missed opportunity;
- `5`: delay creates immediate or irreversible harm.

Urgency must be justified against a stated planning horizon.

## Base score

The provisional v0 formula is:

```text
base_score = (impact × risk_reduction × optionality × urgency) / effort
```

The formula is intentionally simple and inspectable. It is provisional until validated against golden calibration scenarios.

The engine must retain:

- all axis values;
- rationale per axis;
- formula semantic version;
- exact computed base score.

## Confidence

Confidence is an ordinal confidence indicator in the closed interval `[0, 1]`.

It represents confidence in the quality and completeness of the evaluation inputs—not a calibrated probability that a project will succeed.

Confidence must include a rationale covering material uncertainty, missing evidence, and assumptions.

Default interpretation:

- `< 0.40`: weak evidence; orientation should normally request re-evaluation;
- `0.40–0.69`: usable with visible uncertainty;
- `0.70–0.89`: well-supported evaluation;
- `>= 0.90`: exceptional evidence; must not be assigned merely because the evaluator feels certain.

## Freshness

An `EvaluationVersion` records:

- evaluation time;
- planning horizon;
- freshness metadata;
- semantic version.

RFC 0001 does not prescribe freshness decay. RFC 0002 may use freshness as an orientation input, but it must expose the applied rule and must not mutate the evaluation version.

## Evaluation drafts and acceptance

A draft may be produced by:

- the maintainer;
- a deterministic assistant workflow;
- an AI invocation.

A draft becomes an accepted `EvaluationVersion` only through an explicit maintainer-authorized action.

Acceptance creates a new immutable version. It does not overwrite prior versions.

## Validation

An evaluation is invalid when:

- an axis is missing or outside `1–5`;
- rationale is missing;
- confidence is outside `[0,1]` or lacks rationale;
- semantic version is unavailable;
- project identity is unresolved;
- required provenance is missing.

Invalid evaluations must not be silently coerced or used as canonical orientation input.

## Calibration requirements

Before implementation acceptance, the project must define golden scenarios covering at minimum:

- high impact / high effort;
- urgent low-optionality work;
- security or risk-reduction work with modest direct impact;
- low-effort speculative work;
- stale but strategically important work;
- materially uncertain work.

Tests must verify score calculation and make disagreements in axis interpretation visible.

## AI constraints

AI may draft axis values, rationale, and uncertainty. AI output:

- is a non-authoritative `EvaluationDraft`;
- must identify selected context and provenance;
- must not assign canonical confidence automatically;
- must not create or mutate an `EvaluationVersion` directly.

## Ownership boundary

RFC 0001 owns:

- evaluation axes and rubrics;
- base-score semantics;
- confidence semantics;
- evaluation validation;
- calibration guidance.

RFC 0001 does not own:

- `now / next / later / kill` placement;
- queue limits;
- policy composition;
- accepted orientation;
- lifecycle transitions;
- review findings or learning automation.

## Consequences

### Benefits

- Small, explainable scoring model.
- Human judgment remains visible through rationale and confidence.
- Evaluation history is immutable and replayable.
- Orientation can evolve without redefining evaluation records.

### Risks

- The multiplicative formula may over-amplify axes.
- Ordinal scoring can appear more objective than it is.
- Different users may calibrate scales inconsistently.

Mitigation: golden scenarios, explicit rationale, versioned semantics, and later calibration evidence. Automatic weight or heuristic learning is deferred.

## Acceptance criteria

- evaluation and orientation responsibilities are separated;
- every axis has anchors and rationale requirements;
- confidence is explicitly non-probabilistic;
- accepted evaluations are immutable versions;
- drafts and AI output remain non-authoritative;
- malformed evaluations fail visibly;
- formula and calibration scenarios are versioned;
- no lifecycle, placement, or policy semantics are duplicated here.
