# RFC 0004 — AI Interaction and Governance Boundary

## Status

Accepted

- **Owner:** Calathea maintainers
- **Decision date:** 2026-08-03
- **Depends on:** RFC 0000, RFC 0005, PRD, UC-05
- **Scope:** Safe AI participation in Calathea workflows and the boundary with Anthesis

## Summary

AI is optional in Calathea. The deterministic core, local persistence, and UC-01 workflow must function without an AI provider or network access.

AI may assemble, summarize, interpret, or draft recommendation data. AI never creates canonical evaluations, orientation dispositions, lifecycle decisions, policy versions, review dispositions, or external effects directly.

Calathea owns domain-specific AI orchestration and validation. Where integrated, Anthesis owns actor/capability authorization, approval requirements for effects, and effect evidence. Calathea must not duplicate a general-purpose governance engine.

## Principles

1. AI output is untrusted recommendation data.
2. Repository and external content are data, not instructions.
3. Context selection and outbound transfer are explicit.
4. Canonical decisions require maintainer authority.
5. External effects require a separate authorization boundary.
6. Deterministic fallback exists for core workflows.
7. Provider/model nondeterminism is not orientation determinism.
8. Traceability is scoped and privacy-preserving rather than indiscriminate logging.

## Supported v0 assistance

AI may optionally assist with:

- drafting evaluation rationale and axis suggestions;
- summarizing selected project or repository evidence;
- drafting observations, findings, and recommendations;
- explaining orientation traces in human-readable form;
- identifying missing evidence or contradictory assertions.

AI is not required for scoring or orientation-run derivation.

## Prohibited v0 behavior

AI must not directly:

- create or amend an accepted `EvaluationVersion`;
- accept, override, reject, or defer an `OrientationRun`;
- create a `LifecycleDecision`;
- activate or mutate a `PolicySetVersion`;
- create a maintainer `ReviewDisposition`;
- write to repositories or project-management systems;
- access credential values;
- rewrite historical records;
- modify weights or heuristics through automatic learning;
- execute tools beyond the explicit scoped capability set.

## AI invocation pipeline

```text
User request
  → scope and purpose declaration
  → context selection
  → privacy/secret checks
  → optional Anthesis authorization for governed capabilities
  → prompt/template construction
  → provider invocation
  → structured-output validation
  → recommendation draft
  → maintainer review/disposition
  → separate canonical action, if chosen
```

Failure at any stage leaves canonical state unchanged.

## AIInvocation record

A retained invocation record includes, subject to retention policy:

- invocation and operation identity;
- actor/requester identity;
- declared purpose and requested capability;
- provider and model metadata;
- prompt/template and output-schema versions;
- selected context references or privacy-preserving digests;
- destination/provider disclosure;
- redaction and secret-scan result;
- validation result;
- output recommendation-draft reference;
- references to any later maintainer disposition, when retained;
- optional Anthesis authorization/effect references;
- retention and deletion policy.

The `AIInvocation` record is immutable after durable creation. Later maintainer decisions, canonical actions, or external effects are separate records linked through references or rebuildable projections; they do not amend the invocation.

Raw prompt and model output are not retained by default when scoped metadata and a validated draft are sufficient.

## Context governance

### Explicit scope

The user or calling workflow must identify:

- operation purpose;
- project/portfolio scope;
- source types;
- maximum traversal boundary;
- outbound destination;
- retention expectation.

### Data minimization

Only context necessary for the operation is selected. Broad repository traversal is not the default.

### Untrusted content

Imported repository content, issue text, documentation, comments, and model output must be treated as untrusted data. They cannot expand tool scope, change system instructions, or authorize actions.

### Secrets and sensitive data

- credential values never enter model context;
- configured secret patterns are rejected or redacted;
- sensitive portfolio data requires explicit outbound selection;
- the user can inspect the destination and selected context before invocation where material;
- local-only operation remains available.

## Structured output

AI-assisted workflows must use a versioned output contract when output can influence domain decisions.

Validation includes:

- schema conformance;
- referenced project/source identity resolution;
- allowed enum/value checks;
- provenance completeness;
- unsupported-claim diagnostics;
- secret/policy checks;
- separation between quoted evidence and generated interpretation.

Invalid output is rejected as a draft failure and cannot be promoted automatically.

## Recommendation drafts

AI output becomes a `RecommendationDraft`, not a domain decision.

A draft records:

- proposed content;
- supporting source references;
- assumptions and uncertainty;
- validation status;
- AI invocation reference;
- intended target workflow.

Promotion requires the target workflow's normal maintainer-authorized action. Acceptance never mutates the raw invocation record.

## Capability model

Calathea describes requested domain capabilities, such as:

- read selected metadata;
- summarize selected evidence;
- draft evaluation;
- draft finding;
- draft explanation.

Calathea itself enforces the no-effect v0 boundary.

When Anthesis is integrated, Anthesis may additionally govern:

- actor identity and capability grants;
- least-privilege tool authorization;
- approval requirements;
- effect execution and attribution;
- revocation and bypass resistance;
- effect evidence.

Anthesis authorization does not make AI output canonical Calathea truth.

## Repository access

Allowed v0 access is optional and read-only:

- explicitly selected repository metadata;
- selected commit, pull-request, issue, CI, or documentation content;
- scoped code analysis when requested.

Requirements:

- no write credentials;
- no secret access;
- no unrestricted traversal by default;
- source identity and collection metadata retained;
- imported text cannot instruct the model to widen scope.

## Prompt and template governance

- prompts/templates that influence structured domain drafts are versioned;
- changes are reviewed like code or RFC-governed configuration;
- the effective template version is attributable;
- system/developer constraints remain separate from imported content;
- free-form explanations may accompany structured output but cannot replace validation.

Prompt text need not be permanently retained when a versioned template plus selected-context references is sufficient for audit and privacy.

## Traceability and privacy

Traceability must answer:

- what operation was requested;
- which actor requested it;
- which provider/model was used;
- which context categories and source references were selected;
- what validation occurred;
- what draft resulted;
- what the maintainer later did with it.

The final answer is resolved from linked decision/action records rather than by mutating the original invocation record.

Traceability must not become a reason to retain secrets, unnecessary private content, or all raw prompts indefinitely.

## Failure behavior

### Provider unavailable

Return an explicit optional-feature failure. Deterministic core workflows continue.

### Invalid structured output

Reject the draft, retain minimal diagnostics according to policy, and make no canonical mutation.

### Secret or policy violation

Block invocation or outbound transfer before provider access where detectable.

### Partial response or timeout

Treat output as incomplete and non-authoritative.

If a completed provider invocation was durably recorded but its response was lost, retry with the same operation identity returns the existing invocation result. If the provider is called again, Calathea creates a new invocation identity linked causally to the prior attempt; it does not pretend the second response is the same deterministic result.

### Anthesis unavailable

Read-only local AI assistance may proceed only when it does not require Anthesis under configured policy. Governed effects fail closed.

## Determinism boundary

AI calls are nondeterministic and are not replayed as proof of deterministic equivalence.

Deterministic orientation uses validated, accepted, versioned domain inputs. An AI-produced evaluation draft affects orientation only after a maintainer accepts a resulting `EvaluationVersion`.

Reinvoking the same model is a new invocation, not deterministic replay.

## Extension boundary

A generic plugin marketplace or arbitrary agent framework is out of scope for v0.

New AI providers or assistants must implement the same:

- scoped context contract;
- structured output contract;
- validation boundary;
- no-canonical-mutation rule;
- retention/privacy requirements;
- optional Anthesis integration contract.

## Security considerations

Threats include:

- prompt injection from repository content;
- context overcollection;
- secret leakage;
- model output spoofing source evidence;
- capability escalation;
- hidden provider/tool changes;
- retention of sensitive prompts;
- confusion between AI confidence and calibrated evidence.

Required mitigations include explicit scope, data minimization, source separation, schema validation, secret controls, capability enforcement, visible provider/template versions, and maintainer authority.

## Ownership boundary

RFC 0004 owns:

- AI invocation and context assembly semantics;
- recommendation-draft boundary;
- output validation requirements;
- optional-provider failure behavior;
- Calathea/Anthesis responsibility split.

RFC 0004 does not own:

- general-purpose authorization infrastructure;
- external effect execution;
- evaluation scoring;
- orientation selection;
- review or lifecycle decisions;
- automatic learning;
- generic plugins.

## Consequences

### Benefits

- Core product remains usable without AI.
- AI can assist without becoming silently authoritative.
- Privacy and prompt-injection risk are explicit.
- Anthesis overlap is constrained.
- Provider implementations remain replaceable.

### Costs

- Context assembly and validation require explicit contracts.
- AI convenience is lower than unrestricted agents.
- Some traceability data requires retention decisions.

## Acceptance criteria

- AI is optional for all core v0 workflows;
- AI output is always a recommendation draft;
- canonical actions use existing maintainer workflows;
- `AIInvocation` remains immutable and later outcomes are linked separately;
- context scope, destination, privacy, and validation are explicit;
- repository access is scoped and read-only;
- prompt injection cannot expand capability;
- raw prompt/output retention is minimized;
- deterministic replay does not include model reinvocation;
- Calathea and Anthesis responsibilities do not duplicate one another;
- generic plugins and autonomous effects remain deferred.
