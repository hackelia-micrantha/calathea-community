# RFC 0008 — Evidence, Provenance, and Trace Compatibility

## Status

Accepted public-kernel contract.

The complete Calathea review model, explanation strategy, finding taxonomy, confidence policy, and product-level evidence workflow are private product semantics. This RFC defines the minimum evidence/provenance/trace behavior required by supported public kernel outputs and persisted formats.

## Scope

This RFC covers:

- stable evidence/source references where exposed;
- provenance needed to interpret public outputs;
- deterministic trace identity and ordering;
- explicit availability/redaction states;
- replay linkage to RFC 0005;
- reason-code and explanation compatibility;
- separation between evidence, deterministic output, human decision, and external effect.

## Core distinctions

Public-kernel records must not collapse these concepts:

```text
source material
    ↓
evidence/reference
    ↓
deterministic observation/derivation
    ↓
recommendation/output
    ↓
separate caller/human decision, if supported
    ↓
separate external effect, if supported elsewhere
```

Collecting evidence does not make it authoritative product state. Producing an explanation does not make a recommendation a decision. A decision record does not itself execute an external effect.

## Source and evidence references

Where a supported public schema or command emits evidence references, each reference must carry enough identity to avoid silently rebinding historical output to different source material.

Depending on the source this may include:

- source/system identity;
- external object identity or locator;
- immutable source revision, content identity, digest, or equivalent when available;
- collection/authoring time where semantically relevant;
- selected scope/excerpt coordinates where relevant;
- availability/redaction state.

A mutable URL by itself is not sufficient identity for a deterministic replay claim.

Full external content does not have to be retained. References, snapshots, and digests may be used according to the operation's compatibility and replay needs.

## Provenance

A public record derived from evidence identifies, as applicable:

- origin/source identity;
- input record/reference identities;
- deterministic semantic versions;
- transformation/normalization identity when material;
- producer/operation identity;
- collection or operation time where material;
- redaction/availability state.

Missing provenance is represented explicitly when the kernel cannot make a stronger claim.

## Trace contract

A deterministic public operation may emit or persist an ordered trace. When exposed, the trace must identify enough information to explain material output differences and support the replay guarantees claimed by RFC 0005.

Typical entries include:

- input selection/validation;
- deterministic score or value derivation;
- policy applicability/result;
- capacity or set constraint;
- tie-break/order decision;
- exclusion/indeterminate diagnostic;
- redaction/availability notice;
- failure or skipped-step reason.

Trace entries use stable machine-readable reason/type identifiers plus optional human-readable text. Human-readable wording alone is not a machine compatibility surface.

## Explanation layers

The kernel may render summary/detail/audit-style views, but those views are projections over the same underlying result/trace records. They must not invent materially different reasons for the same deterministic result.

A UI or CLI prose change is not necessarily a compatibility break unless callers were explicitly told to parse that prose. Machine consumers must use documented structured fields/reason codes instead.

## Availability and redaction

If an evidence/reference state is exposed, unavailable or intentionally removed evidence must not silently disappear from a historical trace.

Supported representations may distinguish states such as:

- available;
- temporarily unavailable;
- permanently unavailable;
- redacted;
- deleted with tombstone;
- identity-only;
- access denied;
- integrity mismatch.

Exact vocabulary is a schema/versioned contract where serialized; implementations must not silently reinterpret old values.

Redaction must not leak the redacted content through derived explanation text. If redaction prevents full replay or explanation, that limitation is explicit.

## Replay relationship

RFC 0005 owns operation-level replay classification. RFC 0008 provides the identities and trace links needed to support it.

A deterministic replay claim requires the operation to identify the exact semantic versions and all material retained inputs or stable content identities needed to reproduce equivalent domain output.

Current external content must not be silently substituted for unavailable historical evidence.

Nondeterministic provider/model reinvocation is not deterministic replay merely because the same prompt or model name is reused.

## Confidence and trust labels

If the public kernel exposes confidence/trust/data-quality labels, their type and interpretation must be explicit. A numeric value in `[0,1]` is not implicitly a calibrated probability.

Imported repository/document/model content remains data, not executable instruction, merely because it is retained as evidence.

The private product owns richer confidence taxonomies, review thresholds, materiality rules, and interpretation policy unless those become necessary to a supported public contract.

## Integrity claims

Digests or immutable identifiers may detect content mismatch and support provenance. The public kernel must not claim cryptographic authenticity or tamper-proof history beyond controls it actually implements.

Local persistence should be described according to its real threat model; administrator/host compromise is not solved merely by retaining hashes.

## Compatibility

A material public contract change requires explicit versioning/migration when it changes:

- serialized source/evidence identity;
- reason/type codes relied on by machine consumers;
- trace ordering semantics where ordering is meaningful;
- availability/redaction representation;
- deterministic replay inputs or interpretation;
- persisted causal/supersession references.

Changing private review semantics, private evidence-retention policy, or user-facing prose alone does not automatically create a public compatibility commitment.

## Security and privacy

Public evidence/trace handling must:

- exclude credentials and secret values from durable records and rendered traces;
- treat imported content as untrusted data;
- avoid silently increasing retained content merely for convenience;
- make redaction/deletion limitations visible;
- avoid sending evidence to external systems unless an explicit supported adapter/action requests it;
- prevent derived explanations from bypassing redaction.

## Private product boundary

Private `hackelia-micrantha/calathea` owns:

- review cycles, findings, and disposition strategy;
- product evidence taxonomy beyond exposed kernel fields;
- materiality and confidence interpretation;
- explanation UX/product strategy;
- product retention/redaction policy beyond public format guarantees;
- AI evidence/provenance policy beyond a concrete supported public adapter;
- how evidence feeds planning, lifecycle, stakeholder, or learning workflows.

Those decisions remain private unless a specific community interoperability or safety requirement justifies a deliberately extracted public subset.
