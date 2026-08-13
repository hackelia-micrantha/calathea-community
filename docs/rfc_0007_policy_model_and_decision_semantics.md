# RFC 0007 — Policy Model, Composition, and Decision Semantics

## Status

Accepted

- **Owner:** Calathea maintainers
- **Decision date:** 2026-08-03
- **Depends on:** RFC 0000, RFC 0001, RFC 0002, RFC 0005, RFC 0006, PRD, UC-01
- **Scope:** Policy identity, configuration, deterministic evaluation, composition, exceptions, and trace semantics

## Summary

Calathea uses a constrained hybrid policy model:

- policy configuration is declarative, immutable, versioned data;
- each policy selects a built-in evaluator behind a stable interface;
- evaluators are deterministic for retained versioned inputs;
- every material result is recorded as a `PolicyDecision` in the operation trace;
- no unrestricted policy language, arbitrary expression evaluator, or user-provided executable policy code exists in v0.

Policies may constrain eligibility, capacity, readiness, score adjustment, set composition, review requirements, or recommendation behavior. They do not create maintainer decisions, mutate lifecycle state, execute tools, or authorize external effects.

A policy result is distinct from a policy exception. Exceptions are explicit, immutable maintainer decisions with narrow scope and cannot override system invariants or non-exceptionable policies.

## Context

RFC 0002 consumes one `PolicySetVersion` while generating an `OrientationRun`, but the original policy semantics were embedded in orientation-engine prose. That leaves several risks:

- policy changes could be hidden inside application code;
- priority might silently override a hard denial;
- missing evidence might be treated as permission;
- multiple soft adjustments could combine unpredictably;
- overrides might mutate an orientation run or bypass invariants;
- historical runs might become impossible to explain after configuration changes;
- Calathea might duplicate Anthesis authorization and effect-governance responsibilities.

This RFC makes policy a first-class, deterministic domain contract without turning Calathea into a general-purpose policy engine.

## Goals

- Define stable policy identity and immutable policy-set versions.
- Keep policy configuration inspectable and human-readable.
- Separate hard constraints, soft effects, review requirements, and diagnostics.
- Define deterministic applicability, result, composition, conflict, and failure behavior.
- Ensure every material policy effect is visible in orientation or review traces.
- Define explicit exception records without allowing silent bypass.
- Preserve the Calathea–Anthesis responsibility boundary.
- Provide a small initial built-in policy catalog for v0.

## Non-goals

- A general-purpose policy language or unrestricted DSL.
- Arbitrary user-supplied scripts, expressions, or executable plugins.
- Replacing Anthesis authorization, capability, approval, or effect policies.
- Automatically learning policy weights or parameters.
- Defining organization-wide role-based access control.
- Defining persistence, serialization, runtime, or dependency-injection technology.
- Supporting bidirectional synchronization with external policy systems.
- Allowing policy to mutate historical records or external systems.

## Decision principles

1. **System invariants are not policies.** They cannot be disabled or overridden.
2. **Policy configuration is data; policy behavior is versioned code.**
3. **Priority controls deterministic order, not truth.** A lower-priority hard denial is not erased by a higher-priority allow.
4. **Missing evidence is never silently treated as allow.**
5. **Soft effects are typed, bounded, and visible.**
6. **Every historical operation captures exact policy and evaluator versions.**
7. **Exceptions are separate maintainer decisions.** They do not edit a policy set or prior run.
8. **Policy decisions are recommendations or derivation inputs, not human authority.**
9. **Invalid or unsupported policy semantics fail visibly.**
10. **Calathea policy governs domain semantics; Anthesis governs capabilities and effects.**

## System invariants versus policy

System invariants are rules required to preserve Calathea domain correctness. They execute before policy and are not represented as configurable policy instances.

Examples include:

- one project has at most one placement recommendation per orientation run;
- orientation recommendations never mutate project or lifecycle state;
- accepted evaluations, policy-set versions, runs, dispositions, and lifecycle decisions are immutable;
- completed, cancelled, and archived projects cannot enter normal active orientation queues;
- queue output cannot exceed the capacity selected for that run;
- historical records cannot be silently rewritten;
- AI output cannot become canonical without a separate maintainer action;
- credentials cannot become domain evidence or prompt content.

A policy definition that attempts to violate a system invariant is invalid and cannot be activated.

## Conceptual model

### Policy

A stable semantic identity for one rule family, such as `orientation.capacity.now` or `orientation.evaluation.freshness`.

A policy identity is never reused for unrelated behavior.

### Policy evaluator

A deterministic implementation behind a stable interface. It owns the meaning of one evaluator type and version.

An evaluator consumes:

- one validated policy instance;
- one explicit subject;
- one versioned evaluation context;
- required evidence or input references.

It returns one `PolicyDecision` or an operational evaluation failure.

### Policy instance

One configured use of a policy evaluator within a policy-set version.

A policy instance supplies:

- policy identity;
- evaluator type and semantic version;
- effect class;
- workflow phase;
- scope and subject selector;
- parameters;
- required inputs;
- missing-input behavior;
- priority and conflict metadata;
- exceptionability;
- explanation metadata.

### PolicySet

The stable identity of a coherent policy configuration.

### PolicySetVersion

One immutable, fully resolved policy configuration used by deterministic operations.

It contains the exact ordered policy instances and defaults required for evaluation. Templates or profiles may help author a set, but activation materializes a self-contained effective version so later changes to a template cannot alter historical meaning.

### PolicyDecision

The immutable deterministic result of applying one policy instance to one subject in one operation.

It is part of an orientation or review trace. It is not a maintainer decision, authorization grant, lifecycle transition, or external effect.

### PolicyException

An immutable maintainer-authorized decision allowing one narrowly scoped deviation from an exceptionable policy.

It does not modify the policy-set version or the original policy decision.

## Constrained hybrid model

### Declarative configuration

Policy-set configuration may describe only supported fields and typed parameters. It must not contain:

- executable code;
- shell commands;
- arbitrary regular-expression execution against unbounded content;
- general boolean expression trees;
- dynamic network lookups;
- reflection or arbitrary function names;
- embedded prompts that define policy truth;
- provider/model calls.

### Built-in evaluator registry

Each policy instance selects a known evaluator type, for example:

```yaml
policy_id: orientation.capacity.now
instance_id: default-now-capacity
evaluator_type: capacity_limit
evaluator_version: 1
phase: set_constraints
effect_class: hard
parameters:
  placement: now
  maximum: 3
```

The evaluator registry is an implementation detail behind a stable application boundary. The domain requirement is that evaluator identity and semantic version are explicit and replayable.

### Stable interface

Conceptually, an evaluator behaves as:

```text
evaluate(policy_instance, subject, context) -> PolicyDecision | PolicyEvaluationFailure
```

The RFC does not prescribe a language interface, class hierarchy, dependency-injection framework, or module layout.

## Policy instance contract

Every policy instance contains or resolves:

- stable `policy_id`;
- stable `instance_id` within the policy set;
- evaluator type;
- evaluator semantic version;
- configuration schema version;
- workflow scope;
- evaluation phase;
- subject type;
- constrained subject selector;
- effect class;
- parameters;
- required input declarations;
- missing-input behavior;
- priority;
- conflict group or exclusive key where applicable;
- exceptionability and exception constraints;
- enabled state;
- rationale or policy intent;
- provenance and authoring metadata.

Unknown fields are rejected unless the policy schema explicitly permits extensions.

## Workflow scope

A policy declares the workflows in which it may run.

Supported conceptual scopes include:

- orientation candidate eligibility;
- orientation candidate adjustment;
- orientation set composition;
- orientation result validation;
- orientation-disposition validation;
- review diagnostics;
- lifecycle-transition validation, where RFC 0006 delegates a guard to policy.

A policy cannot run in a workflow it did not declare.

v0 implementation may support only orientation scopes while retaining the broader stable vocabulary.

## Evaluation phases

Policies execute in deterministic phases:

1. **Input validation** — verify required policy and subject inputs.
2. **Candidate eligibility** — determine whether a project may participate in a placement surface.
3. **Candidate adjustment** — apply bounded typed soft effects or diagnostics.
4. **Set constraints** — enforce queue capacity, required coverage, or combination rules.
5. **Result validation** — verify the proposed orientation satisfies all applicable hard constraints.
6. **Disposition validation** — validate overrides and referenced exceptions.
7. **Review diagnostics** — produce require-review or advisory results without mutation.

Not every evaluator supports every phase. The policy schema declares valid combinations.

Within a phase, evaluation order is stable by:

1. explicit priority, ascending;
2. policy identity;
3. instance identity.

Priority does not make a policy semantically stronger than another policy class.

## Effect classes

### Hard

A hard policy constrains legal output.

Applicable results may deny eligibility, constrain capacity, require a member or property in a selected set, or reject an invalid disposition.

Hard policies are never silently relaxed.

### Soft

A soft policy influences ranking, tie-breaking, or preferred composition while preserving legal alternatives.

Soft effects must use supported typed descriptors and declared bounds.

### Review-required

A review-required policy does not necessarily deny the operation, but it requires a visible unresolved diagnostic or explicit maintainer review before a configured later boundary, such as accepting an orientation disposition.

### Advisory

An advisory policy produces explanation or diagnostic data only. It cannot alter effective score, eligibility, legal output, or acceptance validity.

## Policy decision results

Each applicable evaluation returns exactly one result status.

### `allow`

The policy's applicable hard condition is satisfied, or its required validation succeeds without an additional effect.

An `allow` from one policy does not cancel a `deny` from another policy.

### `deny`

A hard condition is violated.

The decision records the prohibited subject or output, reason, and relevant evidence. An exception may be considered only if the policy is explicitly exceptionable.

### `adjust`

A soft policy emits one or more typed, bounded effects.

An adjustment is not permission to violate a hard policy.

### `require_review`

The operation may continue only according to the configured review boundary. The unresolved condition remains visible and cannot be treated as ordinary allow.

### `not_applicable`

The policy does not apply to the subject or scope. This is not the same as allow.

### `indeterminate`

The policy applies, but required evidence is missing, conflicting, stale beyond interpretation, or otherwise insufficient to reach the intended result.

The policy instance's declared indeterminate behavior determines whether this:

- excludes the subject;
- denies the operation;
- requires review;
- fails the operation;
- produces a diagnostic only, where safe and explicitly permitted.

Indeterminate never silently becomes allow.

## Policy evaluation failures

`PolicyEvaluationFailure` is distinct from an `indeterminate` decision.

Examples:

- evaluator implementation unavailable;
- evaluator semantic version unsupported;
- invalid configuration escaped activation validation;
- unexpected evaluator exception;
- deterministic dependency corrupted;
- output contract violation.

Default behavior:

- hard, soft, set-composition, or result-validation evaluator failure fails the operation before an authoritative run is persisted;
- an advisory evaluator may be skipped only if its policy instance explicitly declares non-blocking failure behavior, and the failure remains visible in diagnostics;
- failures never become allow, deny, or adjustment by guesswork.

## Required input and indeterminate behavior

Every policy declares the inputs it requires, such as:

- project version fields;
- lifecycle state;
- accepted evaluation version;
- evaluation age;
- confidence category;
- project tags or classifications;
- imported observations;
- prior accepted placement history;
- queue candidate set;
- planning horizon;
- policy-exception references.

Required input declarations support preflight validation and explanation.

Allowed missing-input behaviors are constrained and explicit:

- `deny`;
- `exclude_subject`;
- `require_review`;
- `fail_operation`;
- `diagnostic_only` for advisory policies.

`allow` is not an allowed missing-input behavior.

## Typed soft effects

v0 soft policies may emit only supported effect descriptors.

### Score multiplier

A versioned exact-decimal multiplier applied to effective score.

Requirements:

- fixed-point or exact-decimal semantics; binary floating-point behavior must not define domain truth;
- explicit minimum and maximum per policy type;
- pre-adjustment value, factor, and post-adjustment value in the trace;
- cumulative policy-set bounds;
- no use for hard denial;
- no raw multiplication of RFC 0001 ordinal confidence.

### Tie-break preference

A named, versioned preference applied only after effective score equality or a configured comparison tolerance.

### Set preference

A bounded preference for diversity, coverage, or balance. It cannot violate capacity or another hard constraint.

### Recommendation marker

A typed marker such as kill-candidate evidence or re-evaluation recommendation. It does not create lifecycle state or canonical decisions.

### Diagnostic annotation

A visible explanation-only effect.

New effect types require schema and evaluator versioning. Arbitrary effect names or generic executable expressions are invalid.

## Confidence policy

RFC 0001 confidence is ordinal evidence quality, not a calibrated probability or ratio scale.

Therefore:

- raw confidence must not be multiplied directly into base score;
- threshold gates may deny, exclude, or require review;
- tie-break use is permitted as defined by RFC 0002;
- a numeric adjustment requires an explicit versioned ordinal-to-adjustment mapping and calibration evidence;
- absent such a mapping, confidence produces diagnostics or tie-breaking only.

## Freshness policy

Freshness policy operates on an accepted evaluation version's time and planning horizon without mutating that evaluation.

A freshness policy may:

- allow as fresh;
- require review;
- exclude from selected queues;
- apply a bounded score multiplier;
- mark the project indeterminate when timing evidence is unavailable.

Freshness thresholds and adjustment mappings belong to the policy-set version and must appear in the trace.

## Composition semantics

### Policy-set activation validation

Before a `PolicySetVersion` becomes effective, Calathea validates:

- every evaluator and version is available;
- every configuration conforms to its schema;
- selectors and effect classes are supported;
- priorities and instance identities are unique where required;
- cumulative soft-effect bounds are valid;
- conflict groups are resolvable;
- required baseline policies are present;
- no policy attempts to override a system invariant;
- exception settings are internally consistent.

Activation failure creates no effective policy-set version.

### Hard-policy composition

For one subject and phase:

- any applicable unexcepted `deny` remains a denial;
- `allow` does not cancel `deny`;
- `not_applicable` has no effect;
- `indeterminate` follows its declared behavior;
- contradictory hard requirements make the policy set invalid or the operation overconstrained; priority does not choose a winner silently.

### Soft-policy composition

Soft effects combine only through the declared effect-type combinator.

For score multipliers:

1. apply factors in stable policy order;
2. use exact-decimal arithmetic;
3. record each intermediate value;
4. enforce configured cumulative bounds;
5. emit a conflict or failure rather than clamp silently unless the policy set explicitly defines visible clamping semantics.

Tie-break and set-preference effects use their own evaluator-specific deterministic combinators.

### Review-required composition

Any unresolved applicable `require_review` result remains present. A separate explicit disposition or exception resolves it where permitted.

### No implicit short-circuit

The engine should evaluate all independent applicable policies necessary to produce a complete explanation, even when one hard denial already prevents selection.

Evaluation may short-circuit only when later policies depend on output that cannot legally exist; skipped policies receive an explicit `not_evaluated_due_to_prior_denial` trace entry rather than disappearing.

## Conflict model

### Configuration conflict

A conflict detectable without operation data, such as duplicate exclusive policies or contradictory queue limits, invalidates policy-set activation.

### Runtime overconstraint

Valid policies may produce no feasible set for current candidates. The engine returns the partial or empty legal result plus conflict diagnostics. It never relaxes hard policy automatically.

### Soft conflict

Opposing soft effects require a declared deterministic combinator. If none exists, the operation fails policy evaluation rather than choosing by priority silently.

### Hard versus soft conflict

Hard policy wins. The soft effect is suppressed with an explicit trace reason.

### System invariant conflict

The policy set is invalid. No exception is permitted.

## Subject selectors

Selectors are constrained declarative data over supported indexed fields, such as:

- exact portfolio identity;
- project identity set;
- lifecycle state set;
- declared project tags/classifications;
- planning-horizon type;
- orientation placement surface;
- policy workflow phase.

v0 selectors do not support arbitrary nested expressions, unbounded content search, runtime scripts, or AI interpretation.

Selector evaluation is deterministic and versioned.

## Policy decision contract

A `PolicyDecision` records:

- decision identity within the operation trace;
- policy-set-version reference;
- policy and instance identity;
- evaluator type and semantic version;
- workflow and phase;
- subject identity or set identity;
- applicability result;
- result status;
- typed effects;
- required input references;
- evidence references;
- missing/conflicting evidence;
- rationale and reason code;
- priority and conflict metadata;
- exception reference, if applied;
- exact before/after values where relevant;
- deterministic operation reference;
- schema/semantic versions.

A policy decision is immutable within the containing run or review record.

## Policy exceptions

### Purpose

A `PolicyException` permits one explicitly authorized deviation from one exceptionable policy under narrow conditions.

It does not:

- edit the policy definition;
- edit the policy-set version;
- edit the original policy decision;
- edit an orientation run;
- override a system invariant;
- authorize an external effect;
- become reusable without declared scope.

### Exception contract

An exception records:

- stable exception identity;
- policy identity and compatible policy/evaluator versions;
- policy-set-version scope where required;
- subject or subject-set scope;
- permitted deviation;
- workflow/phase scope;
- actor identity and authority;
- rationale;
- supporting evidence/provenance;
- creation time;
- effective and expiry bounds;
- maximum uses or one-shot semantics;
- supersession/revocation references;
- related orientation disposition or lifecycle command where applicable.

### Exceptionability

A policy explicitly declares one of:

- `not_exceptionable`;
- `exceptionable_with_review`;
- `exceptionable_with_constraints`.

System invariants are always non-exceptionable.

### Orientation overrides

An `accepted_with_overrides` orientation disposition may replace a recommendation only when:

- the replacement does not violate a system invariant;
- every affected hard policy remains satisfied or has a valid scoped exception;
- the exception is referenced explicitly;
- original and replacement placements remain visible;
- the original orientation run remains unchanged.

### Revocation and expiry

Revocation or expiry prevents future use. It does not retroactively rewrite a historical run or accepted disposition that validly referenced the exception at decision time.

A new review may identify that a current accepted orientation relies on an expired exception and recommend reorientation.

## Authority

### Policy authoring and activation

For v0, only the maintainer may activate a canonical `PolicySetVersion` or create a `PolicyException`.

Deterministic components may validate policy configuration. AI may draft configuration or rationale but cannot activate it.

### External sources

Imported external policy-like data remains external-authoritative observation until a maintainer explicitly creates a Calathea policy-set version from selected content.

### Ownership versus authority

Project or portfolio ownership metadata does not grant authority to activate or override policy.

## Calathea–Anthesis boundary

Calathea policies govern domain-specific questions such as:

- may this project participate in this orientation surface?
- what queue capacity applies?
- does stale or low-confidence evaluation require review?
- what bounded ranking or set preference applies?
- is a proposed placement override legal within Calathea semantics?

Anthesis policies govern capability and effect questions such as:

- may this actor invoke this tool or provider?
- is approval required for an external side effect?
- may a repository write occur?
- what evidence and attribution are required for an effect?
- has a capability been revoked?

Anthesis authorization does not turn a Calathea recommendation into canonical domain truth. A Calathea policy exception does not authorize an Anthesis-governed external effect.

Where both systems participate, traces reference each other without duplicating decisions.

## Initial v0 policy catalog

### Required baseline policies

#### `orientation.capacity.now`

- evaluator: `capacity_limit`;
- effect: hard set constraint;
- default maximum: `3`;
- exceptionable: no beyond changing the policy-set version before a run.

#### `orientation.capacity.next`

- evaluator: `capacity_limit`;
- effect: hard set constraint;
- initial default maximum: `10`;
- configured value must be visible in every run;
- exceptionable: no within a completed run.

#### `orientation.evaluation.required`

- evaluator: `required_evaluation`;
- effect: hard candidate eligibility;
- missing or malformed accepted evaluation excludes the subject with diagnostics.

#### `orientation.lifecycle.eligibility`

- evaluator: `lifecycle_eligibility`;
- effect: hard candidate eligibility;
- follows RFC 0006;
- `approved`, `active`, and `paused` are eligible;
- `proposed` requires explicit policy permission;
- `candidate`, `completed`, `cancelled`, and `archived` are excluded from normal active orientation;
- cannot override RFC 0006/system invariants.

### Recommended baseline quality policies

#### `orientation.evaluation.confidence`

- evaluator: `confidence_gate`;
- default effect: diagnostic or require review for weak evidence;
- raw confidence is not multiplied into score;
- numeric adjustment disabled until an explicitly calibrated mapping is accepted.

#### `orientation.evaluation.freshness`

- evaluator: `freshness_rule`;
- default effect: visible freshness class and require-review behavior when configured;
- no normative score multiplier until golden scenarios calibrate it.

### Optional policies, disabled by default

#### `orientation.readiness.blocked`

Controls whether an observed blocker excludes `now`, requires review, or only emits a diagnostic. It never changes lifecycle state automatically.

#### `orientation.selection.diversity`

Produces a bounded set preference or hard cap over supported project classifications.

#### `orientation.selection.starvation`

Provides bounded, explicit prior-placement influence. It never overrides hard policy.

#### `orientation.selection.risk_coverage`

Prefers or requires configured risk-reduction coverage in `now` when feasible.

#### `orientation.recommendation.kill`

Marks projects as kill candidates under a versioned calibrated heuristic. Disabled by default until accepted golden scenarios establish safe behavior.

## Baseline policy-set behavior

A v0 baseline policy-set version must be sufficient to run UC-01 without AI or external repository input.

It includes:

- `max_now = 3`;
- one explicit bounded `max_next` value;
- lifecycle eligibility;
- accepted-evaluation requirement;
- deterministic ordering and conflict semantics;
- visible confidence and freshness diagnostics;
- no automatic kill heuristic unless explicitly enabled and calibrated;
- no automatic learning or parameter mutation.

## Versioning and historical capture

Changing any of the following creates a new `PolicySetVersion`:

- policy instance membership;
- evaluator version;
- selector;
- parameter;
- priority;
- conflict group;
- missing-input behavior;
- effect class or bound;
- exceptionability;
- cumulative adjustment bound;
- baseline default.

An `OrientationRun` references exactly one immutable policy-set version and records every material policy decision.

If an evaluator implementation is no longer available, replay reports `not_reproducible` or `unsupported`; it does not substitute a newer evaluator silently.

## Activation and current policy

Policy-set activation is an explicit maintainer action that creates or selects an immutable version.

A current-policy-set reference may be stored as a rebuildable or mutable pointer, but the referenced version remains immutable.

Changing the current pointer does not alter historical runs or dispositions.

## Failure and recovery

### Invalid policy configuration

Activation fails with structured validation diagnostics. No partial effective version is created.

### Unsupported evaluator version

The operation fails before an authoritative run is persisted.

### Missing policy inputs

The policy returns `indeterminate` and follows its declared missing-input behavior.

### Policy evaluator failure

The operation follows the evaluator-failure rules in this RFC. It never guesses a result.

### Overconstrained orientation

Return the partial or empty legal result and complete conflict trace. Do not relax hard policy.

### Lost response after durable run

Retry uses RFC 0005 idempotency semantics and returns the existing run.

### Projection failure

Policy-set current pointers and policy summaries are rebuildable from immutable activation/version records.

## Privacy and security

Policy configuration and traces may reveal sensitive strategy, capacity, risk, or project classifications.

Requirements:

- core policy evaluation remains local and network-free;
- policy evaluation performs no hidden network lookup;
- secrets and credentials are invalid policy parameters;
- selectors operate only on declared supported fields;
- imported repository text cannot define or modify policy;
- AI-drafted policy is untrusted until maintainer activation;
- exports distinguish policy configuration, decisions, exceptions, and projections;
- redaction reports impact on replay and explanation;
- exception records remain attributable and narrowly scoped.

Threats include:

- hidden evaluator behavior changes;
- malicious or accidental broad exceptions;
- policy priority used as silent bypass;
- indeterminate treated as allow;
- multiplier stacking that overwhelms base score;
- stale policy pointers;
- external content attempting to inject configuration;
- divergence between Calathea policy and Anthesis authorization.

Mitigations include immutable versions, evaluator semantic identity, activation validation, typed effects, cumulative bounds, explicit exceptions, complete traces, and responsibility separation.

## Explanation requirements

For every material policy decision, a user must be able to determine:

- which exact policy-set version applied;
- which policy/evaluator version ran;
- why the policy applied or did not apply;
- which inputs and evidence were used or missing;
- what result and typed effect occurred;
- how conflicts were resolved;
- whether an exception applied;
- what before/after score or set effect occurred;
- whether the result was hard, soft, review-required, advisory, or indeterminate.

Detailed evidence identity, availability, and redaction schemas are owned by RFC 0008.

## Testing and calibration

### Unit-level contract tests

Each evaluator requires tests for:

- valid and invalid configuration;
- applicability and selector behavior;
- every result status;
- missing-input behavior;
- deterministic output;
- typed-effect bounds;
- unsupported enum/value handling.

### Composition tests

Golden scenarios must cover:

- hard allow plus hard deny;
- contradictory hard policies;
- hard versus soft conflict;
- several bounded score multipliers;
- confidence gate without raw multiplication;
- freshness require-review behavior;
- indeterminate fail-closed behavior;
- overconstrained capacity/coverage rules;
- optional diversity and starvation interaction;
- kill policy disabled and enabled;
- stable ordering under equal priority.

### Exception tests

Golden scenarios must cover:

- valid one-shot exception;
- expired exception;
- revoked exception;
- wrong subject or workflow scope;
- attempt to override a system invariant;
- historical disposition referencing an exception valid at decision time;
- exception reuse beyond maximum use.

### Replay tests

Identical policy-set version, subjects, evidence, evaluator versions, and operation inputs must produce identical policy decisions and orientation output.

## Consequences

### Benefits

- Policy behavior is visible, versioned, and independently testable.
- Orientation-engine code no longer owns hidden business rules.
- Hard, soft, review-required, advisory, and indeterminate outcomes remain distinct.
- Exceptions preserve human authority without rewriting history.
- Calathea avoids becoming an unrestricted policy platform.
- Anthesis overlap is constrained.

### Costs

- More explicit configuration and trace records.
- Built-in evaluators require versioned maintenance.
- Policy-set activation needs validation and golden tests.
- Exceptions require careful scope and lifecycle handling.
- Some apparently simple preferences cannot be expressed until a supported evaluator exists.

These constraints are intentional: inability to express arbitrary policy in v0 is safer than hidden or irreproducible policy behavior.

## Deferred decisions

- User-defined evaluator plugins.
- General boolean or temporal policy expressions.
- Multi-user policy authoring and approval.
- Remote policy registries.
- Organization-wide policy inheritance.
- Automatically learned parameters or weights.
- Cross-portfolio policy federation.
- Bidirectional Anthesis policy synchronization.
- Generic lifecycle and external-effect policy execution.

## Acceptance criteria

This RFC is accepted because:

- policy configuration is declarative and evaluators are versioned built-ins;
- policy sets are immutable, self-contained effective versions;
- system invariants cannot be overridden;
- `allow`, `deny`, `adjust`, `require_review`, `not_applicable`, and `indeterminate` are distinct;
- evaluation failures are distinct from indeterminate decisions;
- hard and soft composition is deterministic and explainable;
- soft effects are typed, bounded, and fully traced;
- ordinal confidence is not used as an implicit linear multiplier;
- conflicts and overconstraint never cause silent hard-policy relaxation;
- every material policy decision appears in operation traces;
- exceptions require actor, rationale, scope, evidence, expiry/use bounds, and provenance;
- policy versions and evaluator versions are captured in historical runs;
- the v0 catalog supports UC-01 without AI or external integrations;
- no unrestricted DSL or executable policy extension exists;
- the Calathea–Anthesis responsibility boundary is explicit.
