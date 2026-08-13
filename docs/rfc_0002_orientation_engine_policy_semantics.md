# RFC 0002 — Orientation Engine Semantics

## Status

Accepted

- **Owner:** Calathea maintainers
- **Decision date:** 2026-08-03
- **Depends on:** RFC 0000, RFC 0001, RFC 0005, RFC 0006, PRD, UC-01
- **Scope:** Deterministic orientation-run derivation, eligibility, queue selection, tie-breaking, and diagnostics

## Summary

Calathea converts accepted `EvaluationVersion` records and one effective `PolicySetVersion` into an immutable, deterministic `OrientationRun` containing placement recommendations:

- `now`;
- `next`;
- `later`;
- `kill`.

An orientation run is a recommendation record. It never becomes canonical by mutation. A separate maintainer-authored `OrientationDisposition` accepts, overrides, rejects, or defers the run.

This RFC defines engine behavior. Detailed policy composition and exception semantics are owned by RFC 0007, and explanation/evidence schemas by RFC 0008.

## Inputs

A run references exact versions of:

- one portfolio;
- considered project versions;
- one accepted evaluation version per eligible project;
- one policy-set version;
- planning horizon;
- imported observations used, if any;
- scoring, policy, orientation, and schema semantics;
- runtime queue limits and operation identity.

Lifecycle state is an explicit input. Placement is never read from or written to the project record.

## Candidate collection

A project is considered when it belongs to the portfolio and is not excluded by lifecycle or portfolio membership rules.

Default lifecycle eligibility:

- `approved`, `active`, and `paused`: eligible;
- `proposed`: eligible only when policy permits;
- `candidate`, `completed`, `cancelled`, and `archived`: excluded from normal active orientation.

Every considered project receives either a placement recommendation or an explicit exclusion/indeterminate diagnostic.

## Evaluation handling

- Missing evaluation: project is ineligible and receives `missing_evaluation`.
- Malformed evaluation: project is excluded and receives `invalid_evaluation`; malformed data is never coerced.
- Stale evaluation: policy may exclude, penalize, or permit it, but the behavior must be explicit.
- Low confidence: policy may exclude, diagnose, or apply an explicitly calibrated adjustment; the engine must not pretend precision.

The engine never edits evaluation records.

## Effective score

The provisional effective score is:

```text
effective_score = base_score × freshness_factor × policy_factor
```

Requirements:

- every factor and semantic version is retained;
- default factor is `1.0` when no adjustment applies;
- hard-policy decisions are not encoded as a near-zero multiplier;
- soft adjustments are bounded and explained;
- factor application is deterministic;
- effective score is not persisted back into the evaluation version.

Exact freshness and policy-factor defaults remain provisional until RFC 0007 and golden scenarios define them.

### Confidence handling

RFC 0001 defines confidence as an ordinal evidence-quality indicator, not a ratio-scale probability. Therefore, the raw `[0,1]` confidence value must not be multiplied directly into the base score merely because it is numeric.

By default, confidence is used for:

- eligibility or re-evaluation diagnostics;
- explicit policy gates;
- stable tie-breaking after effective score.

A policy set may define a bounded confidence adjustment only when it also defines and versions:

- the ordinal-to-adjustment mapping;
- calibration evidence or golden scenarios;
- threshold and indeterminate behavior;
- explanation requirements.

Absent that explicit mapping, confidence does not alter effective score.

## Deterministic pipeline

1. Validate complete versioned inputs.
2. Collect portfolio candidates.
3. Apply lifecycle and hard eligibility rules.
4. Derive effective scores and diagnostics.
5. Select `now` under capacity and policy constraints.
6. Select `next` from remaining eligible candidates.
7. Classify remaining eligible projects as `later` or `kill`.
8. Emit recommendations, exclusions, policy decisions, tie-breaks, and trace metadata.
9. Persist one immutable orientation run.

No stage may silently relax a hard policy.

## Placement meanings

### `now`

Immediate focus within the planning horizon. Bounded by `max_now`, default `3`.

### `next`

Near-term ready queue. Bounded by `max_next`; the configured value must be visible in the run.

### `later`

Eligible but not currently selected for execution pressure.

### `kill`

Recommendation to stop investing. It is not a lifecycle state, cancellation decision, archive action, deletion, or external effect.

## Selection semantics

### `now`

- must not exceed `max_now`;
- must satisfy hard policies;
- should favor readiness, value, and declared soft policy objectives;
- may remain partially empty when constraints prevent safe filling.

### `next`

- excludes all `now` projects;
- must not exceed `max_next`;
- must satisfy applicable hard policies;
- may tolerate lower readiness or confidence when explicitly diagnosed.

### `later`

All remaining eligible, non-kill projects are placed in `later`.

### `kill`

A kill recommendation requires a versioned policy or heuristic and an explicit reason. No default heuristic becomes normative until calibrated in RFC 0007.

## Tie-breaking

After effective score, the default stable sequence is:

1. higher confidence;
2. fresher evaluation;
3. higher risk reduction;
4. lower effort;
5. longer time since last accepted `now` or `next` placement, when starvation policy is enabled;
6. stable project identity.

Tie-breaks must be represented in the decision trace.

## Starvation handling

Starvation avoidance is optional policy behavior, not an implicit score mutation.

When enabled, it must:

- use bounded influence;
- never override hard policy;
- retain the prior-selection evidence used;
- appear explicitly in the trace.

## Policy behavior

This RFC consumes policy decisions but does not fully define the policy model.

At minimum, the engine distinguishes:

- hard eligibility/capacity constraints;
- soft preferences or bounded adjustments;
- indeterminate policies whose required data is unavailable;
- explicit policy exceptions referenced by an override.

Policy conflicts or indeterminate results must be surfaced. Hard policies are never silently downgraded.

## Explanation and diagnostics

Every considered project must expose enough structured data to answer:

- Was it eligible?
- What evaluation and lifecycle versions were used?
- What base and effective score components applied?
- Which policies affected it?
- Which queue constraint or tie-break mattered?
- Why did it receive its placement or exclusion?

Minimum run-level metadata:

- complete input references/digests;
- semantic versions;
- queue limits;
- policy-set version;
- operation identity;
- deterministic ordering version.

Detailed evidence and redaction behavior are owned by RFC 0008.

## Orientation disposition boundary

The engine creates no accepted orientation.

A maintainer may create a separate disposition:

- `accepted`;
- `accepted_with_overrides`;
- `rejected`;
- `deferred`.

Rejected or deferred runs never replace current accepted orientation. Overrides never edit the original run and must obey policy-exception semantics.

## Drift boundary

The engine may emit diagnostics such as stale input, missing owner, or low confidence. It does not own the drift taxonomy or create canonical review findings.

Review cycles, observations, findings, recommendations, and no-change outcomes are owned by RFC 0003.

## Failure behavior

### No eligible candidates

Return a valid run with empty `now`/`next`, explicit exclusions, and no fabricated placements.

### Policy overconstraint

Return the partial legal result and conflict diagnostics. Do not relax hard policy.

### Missing semantic implementation

Fail the run as not reproducible/unsupported before persisting authoritative output.

### Partial persistence

A run is either durably created as a complete immutable record or not created. Retry uses operation identity and idempotency semantics from RFC 0005.

## Invariants

- `len(now) <= max_now`;
- `len(next) <= max_next`;
- one project has at most one recommendation per run;
- every considered project has a recommendation or diagnostic;
- completed, cancelled, and archived projects do not enter active queues;
- every `now` project has a valid accepted evaluation;
- every `kill` recommendation has an explicit versioned reason;
- identical versioned inputs produce identical output;
- recommendation does not mutate project, lifecycle, or current accepted orientation.

## Deferred

- complete policy composition and exceptions: RFC 0007;
- evidence and explanation schema: RFC 0008;
- multi-user evaluation aggregation;
- learned scoring or policy mutation;
- blocked-specific queue policy beyond explicit configuration.

## Acceptance criteria

- run, recommendation, disposition, and lifecycle state are distinct;
- old `state`, `bucket`, `done`, and `killed` terminology is removed;
- lifecycle eligibility follows RFC 0006;
- drift ownership is removed from this RFC;
- raw ordinal confidence is not treated as a linear multiplier;
- policy relaxation and factors are explicit;
- every considered project is explainable;
- deterministic replay requirements follow RFC 0005;
- no automatic lifecycle or external mutation is possible.
