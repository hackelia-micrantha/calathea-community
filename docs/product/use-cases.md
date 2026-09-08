# Calathea Community Kernel Use Cases

## Status

Accepted public-kernel usage scenarios.

These are deliberately narrower than the private Calathea product's end-to-end project-management use cases. They exist to define and validate supported public behavior, not to publish the complete product workflow model.

---

## KC-01 — Run deterministic portfolio orientation

### Intent

Use the public kernel to derive a reproducible orientation from valid versioned inputs.

### Inputs

- project identities/metadata required by the public schema;
- accepted evaluation values required by the public schema;
- supported policy/evaluator inputs;
- semantic/version information required for deterministic behavior.

### Flow

1. Provide valid inputs through a supported CLI/file surface.
2. Kernel validates schema, ranges, references, and supported semantic versions.
3. Kernel calculates deterministic evaluation/orientation results.
4. Kernel applies supported queue bounds, exclusions, policy effects, and stable tie-breaking.
5. Kernel returns human-readable explanation plus machine-readable output where documented.
6. The caller decides what to do with the recommendation.

### Required properties

- identical supported deterministic inputs produce identical promised outputs;
- every excluded/placed candidate has an inspectable reason where the contract promises explanation;
- malformed/unsupported inputs fail explicitly;
- the kernel does not require AI or network access;
- orientation output does not itself perform an external repository mutation.

---

## KC-02 — Persist, compare, and replay supported records

### Intent

Use public persistence/history behavior without depending on private Calathea data or infrastructure.

### Flow

1. Store records through a supported public persistence surface.
2. Preserve version/identity information required by the public contract.
3. Rebuild replaceable views from authoritative stored records where specified.
4. Compare supported run/record versions.
5. Replay deterministic output when required historical inputs and semantic versions are available.
6. Apply documented migrations when persisted formats change.

### Required properties

- public formats are versioned;
- migrations/failure behavior are documented;
- history promised immutable by the contract is not silently rewritten;
- backup/restore/recovery behavior is testable without private data;
- fixtures remain synthetic.

---

## KC-03 — Automate against the `calathea` process contract

### Intent

Use the executable from scripts or another composition without importing internal Go packages.

### Flow

1. Invoke a documented command.
2. Provide inputs using documented flags/stdin/files.
3. Read documented machine output when automation-safe output is required.
4. Interpret documented exit/error semantics.
5. Avoid scraping human prose or depending on internal Go identifiers.

### Required properties

- supported machine contracts are explicit and versioned;
- invalid input and unsupported-version failures are distinguishable where promised;
- automation does not require access to `calathea-community/internal/...`;
- a future incompatible change follows the repository's compatibility/migration policy.

See [CLI/process compatibility](../architecture/cli-process-contract.md).

---

## KC-04 — Use an optional generic adapter safely

### Intent

Integrate an explicitly supported optional source or AI/invocation adapter without changing deterministic-kernel authority.

### Flow

1. Enable/configure the adapter explicitly.
2. Resolve credentials out of band.
3. Bound and validate adapter input/output according to the public contract.
4. Keep imported external content distinguishable from kernel-owned state.
5. On failure, surface diagnostics and leave deterministic state unchanged unless the documented transaction contract says otherwise.

### Required properties

- integrations are optional;
- credentials are excluded from durable kernel records, prompts/evidence/traces, and fixtures;
- imported content is treated as untrusted data;
- AI/provider output is not automatically authoritative;
- disabled optional integrations produce no required network dependency.

The private product may apply stricter routing, model, policy, or approval rules. Those are not public-kernel requirements unless exposed through a generic supported interface.

---

## KC-05 — Implement or extend a compatible kernel surface

### Intent

Allow contributors/integrators to change the public implementation without access to private product documentation.

### Flow

1. Identify the public behavior or compatibility contract being changed.
2. Update implementation and synthetic tests.
3. Update the corresponding public RFC/ADR/schema/process contract when the change is durable and externally observable.
4. Document compatibility/migration consequences.
5. Keep product-specific rationale, calibration, and dogfood out of the public contract unless deliberately released.

### Required properties

- clean checkout build/test remains independent of private repositories;
- public behavior is testable from synthetic inputs;
- internal refactors do not become compatibility promises accidentally;
- a product need is translated into the minimum generic public contract.

---

## Cross-cutting kernel invariants

All supported public-kernel use cases should preserve, where applicable:

- deterministic/versioned semantics;
- explicit validation failures;
- stable machine-facing contracts;
- no mandatory hosted account or telemetry;
- no secret material in durable content or fixtures;
- no hidden network dependency for deterministic workflows;
- attributable imported data;
- safe failure/recovery;
- compatibility and migration clarity.

## Out of scope

The complete private Calathea product use cases for planning, review, implementation feedback, lifecycle/milestones, archiving, stakeholder strategy, AI paved road, and governed effects are intentionally not replicated here.

A public kernel capability may support one step of those workflows without making the full workflow a public specification.
