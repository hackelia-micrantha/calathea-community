# RFC 0000 — Community Kernel Record and Identity Model

## Status

Accepted public-kernel contract, narrowed under issue #48.

The complete Calathea product domain model is private. This RFC defines only the identity/state distinctions required by supported public process/file/schema behavior.

## Public concepts

### Stable identity versus version identity

A logical record/project may have a stable identity while material content changes through explicit versions.

Public formats that claim replay/history compatibility must distinguish the two.

### Caller-authored / authoritative input

State supplied or accepted by the caller through a supported kernel operation. The public kernel validates and persists it according to the documented contract; it does not infer product-level authority beyond that interface.

### Derived/recommended output

Deterministic results such as scores, orientation placements, exclusions, and traces derived from exact versioned inputs.

Derived output is not silently converted into an external effect or private-product decision.

### Imported data

If a supported adapter exists, external data retains source identity/revision/time and remains untrusted. Storage does not transfer source authority.

### Historical/versioned record

A retained prior input/run/decision required by the public persistence/replay contract. Material correction creates a new version/superseding record rather than silently rewriting history promised immutable.

### Projection

A replaceable/rebuildable current view derived from authoritative/versioned public records. A projection is not the sole history source.

## Minimum public entities

Only entities exposed by supported public contracts belong here. Current kernel work may require representations for:

- project identity/minimum metadata;
- evaluation version;
- policy/evaluator input version;
- orientation run and placement result;
- supported caller disposition/override where implemented;
- evidence/trace values used by deterministic output;
- persistence/migration metadata.

Private Calathea concepts such as complete planning, review, lifecycle, milestone, archival, stakeholder, AI workflow, and effect models are not part of this RFC unless a future concrete public surface extracts a subset.

## Identity/version invariants

- stable IDs are not silently reused for a different logical object;
- material versions are attributable and ordered/related according to the public schema;
- historical output identifies the input/semantic versions needed by its compatibility/replay claim;
- unsupported/unknown versions fail explicitly rather than being silently reinterpreted;
- corrections/supersession preserve the history guarantees stated by the public persistence contract.

## Recommendation/effect boundary

The public orientation kernel derives recommendation data. It has no ambient authority to mutate external repositories/project systems.

Any future public effect surface requires a separate explicit contract.

## Scope boundary

The private `hackelia-micrantha/calathea` repository owns the richer product terminology and product-level authority model.

A new concept is added to this public RFC only when an independent user/implementation must rely on it for compatibility, safety, contribution, or useful standalone operation.
