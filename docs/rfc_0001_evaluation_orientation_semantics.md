# RFC 0001 — Deterministic Evaluation Contract

## Status

Accepted public-kernel contract, narrowed under issue #48.

The private Calathea product owns evaluation philosophy, calibration, strategic interpretation, and deployment defaults. This RFC defines only the deterministic evaluation behavior required by supported public interfaces.

## Public evaluation input

A supported evaluation representation must identify, as required by its schema:

- project identity;
- evaluation values/fields;
- semantic/schema version;
- optional rationale/evidence references where the public format exposes them;
- freshness/confidence metadata only where they affect supported deterministic behavior.

The exact public schema is authoritative over this prose.

## Validation

Public evaluation inputs must fail explicitly when they contain:

- unsupported semantic/schema versions;
- missing required fields;
- invalid enum/range/value forms;
- unresolved required references;
- malformed numeric values or representations that would make deterministic behavior ambiguous.

The kernel must not silently substitute optimistic/default values for missing required inputs unless that behavior is explicitly part of the versioned public contract.

## Deterministic evaluation

Where the kernel exposes a score or derived evaluation result:

- identical versioned inputs produce identical promised output;
- rounding/normalization behavior is stable and versioned where material;
- no hidden network/model/provider state participates;
- diagnostics distinguish validation failure from a valid low/indeterminate result;
- public machine output identifies the semantic version needed to interpret the result.

## Confidence and freshness

Confidence/freshness are public only to the extent a supported schema or deterministic evaluator relies on them.

If freshness affects eligibility/score, the effective rule/version must be visible in the trace or documented contract. Stale input must not be silently represented as current.

The public kernel does not define the private product's calibration philosophy or claim that confidence is a calibrated probability.

## AI boundary

AI-generated evaluation drafts are not required for deterministic kernel operation.

If a future public adapter supplies candidate evaluation data, it remains untrusted input until it satisfies this contract and the caller accepts it through whatever higher-level workflow owns authority.

## Private product boundary

The following remain private by default:

- why Calathea chooses particular evaluation axes;
- scoring weights/defaults;
- calibration rubrics and dogfood conclusions;
- strategic interpretation of impact, optionality, urgency, or risk;
- AI prompting/model selection for evaluation drafts.

Only expose those details publicly if they become required to interpret a supported public compatibility surface.
