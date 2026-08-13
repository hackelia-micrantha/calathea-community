# RFC 0003 — Review, Feedback, and Calibration Semantics

## Status

Accepted

- **Owner:** Calathea maintainers
- **Decision date:** 2026-08-03
- **Depends on:** RFC 0000, RFC 0005, RFC 0006, PRD, UC-03, UC-04
- **Scope:** Review cycles, observations, findings, recommendations, dispositions, outcomes, and calibration signals

## Summary

Calathea reviews project and portfolio evidence through explicit, immutable `ReviewCycle` records.

A review may produce observations, findings, recommendations, and maintainer dispositions. It may also validly conclude that no material change is required.

Review never mutates evaluations, orientation placement, lifecycle state, policy, or external systems directly. Any resulting change occurs through its own explicit workflow and authoritative record.

Automatic heuristic or policy learning is deferred from v0. Calathea records calibration signals for later analysis but does not silently alter scoring behavior.

## Review cadence

Calathea supports:

- scheduled review;
- ad-hoc review triggered by changed evidence;
- targeted review of a project, orientation, policy effect, lifecycle state, or outcome.

No universal weekly cadence is normative in v0. Cadence is configuration or maintainer practice and must not create hidden state mutation.

## Review inputs

A review identifies exact versions or references for:

- review subject and scope;
- project and lifecycle records;
- evaluation versions;
- orientation runs and dispositions;
- policy decisions and diagnostics;
- imported observations;
- external evidence references;
- prior findings and dispositions;
- outcomes and calibration signals;
- review semantic version.

Missing, stale, inaccessible, or conflicting evidence must remain visible.

## Review record model

### Observation

An attributable statement about authored, imported, or derived data. It is factual or explicitly indeterminate and does not itself assert that corrective action is required.

### Finding

An evidence-backed interpretation that an expectation, assumption, evaluation, orientation, lifecycle state, or rationale may no longer hold.

A finding records:

- subject;
- supporting and conflicting evidence;
- confidence/data-quality semantics;
- materiality;
- review rule/version;
- explanation.

### Review recommendation

A proposed response to a finding. It has no mutation authority.

Examples:

- create an evaluation draft;
- run orientation again;
- consider a lifecycle transition;
- resolve missing ownership;
- record an outcome;
- take no action and retain current state.

### Review disposition

The maintainer response to a finding or review cycle, such as:

- affirm;
- dismiss;
- defer;
- accept recommendation;
- initiate a separate action;
- no change required.

A disposition records actor, authority, rationale, and time.

## Drift taxonomy

The review system may identify:

- **staleness drift:** relevant evaluation or evidence exceeds configured freshness expectations;
- **priority drift:** accepted orientation appears inconsistent with current evaluation or policy evidence;
- **execution drift:** observed project activity differs materially from accepted intent or placement assumptions;
- **estimation drift:** observed effort differs materially from evaluation assumptions;
- **ownership drift:** accountability metadata is absent or contradicted;
- **narrative drift:** project rationale or intended outcome no longer matches evidence;
- **lifecycle drift:** current lifecycle state appears inconsistent with evidence;
- **evidence drift:** a required source is unavailable, changed, or no longer supports the interpretation.

Drift detection creates observations or findings only. It never changes canonical state automatically.

## Valid review outcomes

A review is valid when it records one of:

- one or more findings and dispositions;
- observations without a material finding;
- explicit no-change conclusion with sufficient reviewed scope;
- indeterminate conclusion because required evidence is unavailable.

A review is not required to create action or mutation. Requiring change would reward churn and undermine audit quality.

## Action boundaries

A review may recommend or initiate a separate command, but must preserve separation:

- evaluation change → new `EvaluationVersion` acceptance workflow;
- orientation change → new `OrientationRun` and disposition;
- lifecycle change → new `LifecycleDecision`;
- policy change → new `PolicySetVersion`;
- external effect → separate authorization/effect workflow.

Failure of a follow-on action does not rewrite the review result.

## Completion and outcome review

Completion follows RFC 0006 and requires an outcome record or equivalent accepted outcome content.

A review may assess whether outcome evidence is sufficient, but it does not mark a project completed.

Outcome result values may include:

- successful;
- partial;
- unsuccessful;
- superseded;
- no longer needed.

Completion does not imply success.

## Calibration signals

Calathea may record immutable signals such as:

- estimated versus observed effort;
- predicted versus observed urgency consequences;
- accepted orientation versus later outcome;
- repeated missing evidence;
- rationale contradiction;
- evaluation confidence versus evidence quality.

A calibration signal is evidence for later analysis. It does not automatically alter:

- evaluation axes;
- confidence values;
- scoring formula;
- policy weights;
- freshness rules;
- future project estimates.

Any later automated learning model requires a separate RFC with bounded changes, rollback, validation, provenance, and explicit maintainer activation.

## AI use

AI may assist with:

- summarizing scoped evidence;
- drafting observations or findings;
- proposing recommendations;
- identifying possible contradictions.

AI output remains a non-authoritative recommendation draft. The review record must distinguish:

- source evidence;
- deterministic analysis;
- AI-generated content;
- maintainer-authored decisions.

Repository content and external text are untrusted data, not instructions.

## Evidence and confidence

A finding must not use confidence as an unexplained probability.

At minimum, it records:

- evidence coverage;
- unavailable or conflicting sources;
- confidence/data-quality category;
- material assumptions;
- freshness.

Detailed schemas and redaction behavior are owned by RFC 0008.

## Failure and recovery

### Evidence unavailable

Record an indeterminate observation/finding and do not fabricate certainty.

### Partial collection

Incomplete evidence batches are identified and cannot silently appear complete.

### Review persisted, follow-on action fails

The review remains valid. The failed action is separately recorded and retryable.

### Duplicate review request

Idempotency semantics return the existing review result for equivalent scope and inputs.

### Projection failure

Current review summaries may be rebuilt from immutable review records.

## Invariants

- reviews may validly produce no change;
- observations, findings, recommendations, dispositions, and mutations are distinct;
- review output never silently changes canonical state;
- findings retain evidence and semantic-version references;
- follow-on actions use their own authority and concurrency rules;
- automatic heuristic learning is absent from v0;
- outcomes do not retroactively rewrite historical evaluations or runs;
- AI output is visibly non-authoritative.

## Ownership boundary

RFC 0003 owns:

- review-cycle semantics;
- drift taxonomy;
- observations and findings;
- review recommendations and dispositions;
- outcome review;
- calibration-signal recording.

RFC 0003 does not own:

- evaluation scoring;
- orientation-run selection;
- lifecycle transitions;
- policy composition;
- automatic learning;
- external effects.

## Consequences

### Benefits

- Review reflects evidence rather than forcing churn.
- Findings and decisions remain auditable.
- Learning data can accumulate without silently changing behavior.
- Lifecycle and orientation remain authoritative in their own workflows.

### Costs

- More explicit records than a single review action list.
- Follow-on actions require separate commands.
- Calibration requires deliberate analysis before automation.

## Acceptance criteria

- no-op and indeterminate reviews are valid;
- review does not enforce mutations;
- drift taxonomy is owned here rather than RFC 0002;
- observation, finding, recommendation, disposition, and action are separate;
- completion follows RFC 0006;
- automatic confidence/effort/heuristic mutation is removed from v0;
- calibration signals are immutable evidence only;
- AI output remains non-authoritative;
- failures do not produce partial canonical mutation.
