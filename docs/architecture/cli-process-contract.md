# Calathea CLI and Process Compatibility Contract

## Status

Accepted compatibility contract for the v0 public executable boundary.

- **Governing decision:** [ADR 0005 — Public Go Module and Process Boundary](../adr/0005_public_go_process_boundary.md)
- **Product constraints:** [PRD](../product/prd.md), [UC-01](../product/use-cases.md#uc-01-orient-a-portfolio-of-existing-projects)
- **Tracked by:** public issue #5

## Purpose

The `calathea` executable is the supported public integration boundary for v0. This document defines what a caller may rely on without turning internal Go packages or incidental human-facing CLI text into accidental compatibility promises.

The contract is deliberately narrow while the UC-01 command surface is still being implemented.

## Compatibility layers

Calathea distinguishes four layers:

1. **Process identity** — binary name, command dispatch, exit status, stdout/stderr behavior.
2. **Human CLI** — interactive/help/diagnostic text intended for a maintainer.
3. **Machine contract** — explicitly documented structured output or input formats with independent schema/semantic identity.
4. **Go implementation** — packages under `internal/`; not a public compatibility surface.

A consumer must not infer machine stability from human-readable output or Go identifiers.

## Current stable process anchors

The following behaviors are compatibility commitments for the current v0 executable boundary.

### Binary name

The executable is named:

```text
calathea
```

Repository/module identity may differ from the executable name.

### Success status

Exit status `0` means the requested process operation completed successfully according to that command's contract.

### Invocation/usage failure status

Exit status `2` means the invocation is not a valid supported command invocation, including an unknown top-level command.

Current behavior for an unknown command is equivalent to:

```text
$ calathea unknown
unknown command: unknown
$ echo $?
2
```

The diagnostic text is human-facing; the stable machine fact is exit status `2`.

### Version command

```text
calathea version
```

is the stable command for querying the executable version identity.

It exits `0` and currently emits:

```text
calathea dev
```

The token after `calathea` is the version identity. `dev` denotes an unversioned development build; it is not a release version and does not imply cross-revision compatibility.

Release packaging may replace `dev` with a release/build version without changing the command contract.

### No-argument invocation

A no-argument invocation currently exits `0` and prints a human-facing product orientation message.

The exact text is **not** a machine compatibility contract. Automation must not parse it.

## Stdout and stderr

### Current rule

- successful primary command output is written to **stdout**;
- invocation errors and diagnostics are written to **stderr**;
- exit status communicates success/failure independently of text.

### Future structured-output rule

When a command supports an explicitly documented machine-readable mode:

- stdout must contain only the structured primary result for that mode;
- human diagnostics, warnings, and recovery guidance must go to stderr;
- the process must not prepend banners, progress text, or prose to structured stdout;
- callers must still inspect exit status rather than treating parseable stdout as proof of success.

## Exit-code compatibility

Only documented exit codes carry stable semantic meaning.

Currently:

| Code | Stable meaning |
| --- | --- |
| `0` | Successful command completion |
| `2` | Invalid/unsupported command invocation or usage |

Other non-zero exit codes are **not yet assigned stable public meanings**. Consumers must not depend on undocumented numeric distinctions.

When new stable exit codes are introduced:

- their semantic category must be documented before consumers are expected to rely on them;
- a command must not reuse an established code for an incompatible meaning;
- machine-readable diagnostics may provide finer detail without requiring a proliferation of process codes.

## Command compatibility

A command becomes part of the supported public process contract only when it is documented as such.

For a documented command, compatibility includes as applicable:

- command/subcommand name;
- required and optional argument semantics;
- documented flags and their value domains;
- exit-code meanings;
- machine-input/output schemas explicitly declared stable;
- side-effect/authority semantics;
- deterministic/offline guarantees where specified.

Implementation-only commands, experimental flags, undocumented environment variables, debug text, ordering of human prose, and whitespace in human-readable output are not compatibility commitments unless a contract explicitly says otherwise.

Renaming or removing a documented stable command is a breaking process-contract change.

## Human-readable output

Human-facing output is optimized for maintainers and may evolve for clarity.

Unless a command contract explicitly freezes a text field, callers must not depend on:

- prose wording;
- punctuation;
- whitespace or wrapping;
- line ordering intended only for display;
- help-text formatting;
- diagnostic wording.

Automation should use exit status and explicitly versioned machine formats instead.

## Structured machine contracts

Machine-readable output does **not** become stable merely because it happens to be JSON, YAML, or another parseable format.

Before structured output is a supported cross-repository contract, it must define:

- an explicit schema or semantic identifier/version;
- field meaning and required/optional status;
- unknown-field compatibility behavior;
- ordering semantics where ordering is meaningful;
- representation of partial/indeterminate/failure states;
- redaction and sensitive-data behavior;
- compatibility/migration rules.

The schema version is independent of the executable build version. A binary update does not silently redefine an existing schema version.

A breaking machine-contract change requires a new schema/semantic version or an explicitly documented pre-stable migration policy. Existing persisted/replayable records retain the semantic identities needed to interpret them.

## Input/configuration contracts

The same rule applies to files or structured stdin used across the repository boundary.

A file format is a supported compatibility surface only after its schema/version and validation behavior are documented. Private composition must not depend on incidental internal serialization layouts.

Unknown or unsupported semantic versions fail visibly rather than being interpreted as the nearest available version.

## Versioning expectations

### Development builds

`calathea dev` identifies a development build. Private dogfood/integration workflows should pin an exact reviewed public revision rather than infer compatibility from `dev`.

Documented stable anchors in this contract remain intentional even during development; changes to them require an explicit contract update and migration rationale.

### Tagged releases

When tagged releases are introduced, release versioning describes the executable release. Machine schemas and durable semantic records retain their own explicit versions.

Release compatibility policy must not substitute for domain semantic versioning: a new binary may support several historical schema/semantic versions, and a persisted record must remain interpretable by the version identities it carries.

### Breaking changes

A breaking public process change requires all of:

1. explicit documentation of the changed contract;
2. migration/compatibility impact recorded in the PR/release evidence;
3. appropriate release-version treatment once a release policy is established;
4. new machine schema/semantic identity when the structured contract itself changes incompatibly.

## Internal Go packages

Packages beneath:

```text
github.com/hackelia-micrantha/calathea-community/internal/...
```

are implementation details, even when a type or function is exported within its package.

No compatibility guarantee is made for:

- internal package paths;
- internal exported identifiers;
- struct layout;
- constructors;
- interfaces used solely for internal composition;
- internal serialization helpers.

External/private consumers must not vendor, copy, or otherwise bypass Go's `internal` boundary to create a de facto API.

## Criteria for a future exported Go API

An exported non-`internal` Go facade requires a separate ADR backed by a concrete in-process consumer.

At minimum, that decision must establish:

1. **Concrete need** — a real consumer requires in-process composition; repository topology or convenience alone is insufficient.
2. **Process-boundary insufficiency** — the use case explains why CLI/file/schema composition is materially inadequate, such as measured latency, embedding, transactional composition, or deployment constraints.
3. **Small facade** — expose the minimum capability-oriented surface, not the current internal domain tree wholesale.
4. **Ownership** — define which types/errors/behaviors become compatibility commitments and which remain internal.
5. **Versioning** — define source/binary/semantic compatibility expectations for the exported package.
6. **Security boundary** — preserve local-first, authority, privacy, and deterministic-core invariants.
7. **Conformance** — tests/examples prove the facade behaves consistently with the executable/domain contracts.
8. **Migration** — private/process consumers have a deliberate adoption path; no dual canonical implementations are introduced.

Until that ADR is accepted, there is no supported public Go library API.

## Testing and conformance

The current application tests cover the stable process anchors:

- no-argument invocation succeeds;
- `version` succeeds and emits the version identity;
- unknown top-level command writes the diagnostic to stderr and exits `2`.

As functional commands and structured formats are introduced, their public process contracts should gain explicit compatibility/conformance tests rather than relying only on implementation-level unit tests.

## Security and privacy

Compatibility must not freeze insecure disclosure behavior.

A stable machine contract must state redaction/sensitive-data semantics where relevant. Credentials and secret values never become valid public CLI output merely because an earlier implementation accidentally emitted them.

Security fixes may reduce disclosed information without preserving unsafe human-facing text. If a machine field itself is security-sensitive, the schema/migration decision must make the change explicit.

## Relationship to private composition

Private `hackelia-micrantha/calathea` consumes a pinned reviewed public revision and composes through this process/file/schema boundary.

Private automation should:

- invoke documented commands only;
- inspect exit status;
- avoid parsing human prose;
- use only explicitly versioned machine contracts;
- update its public-core pin deliberately after reviewing compatibility changes.

## Non-goals

This contract does not:

- freeze the full future UC-01 command naming before implementation;
- define a generic RPC protocol;
- introduce daemon/server requirements;
- promise stable human-readable prose;
- export internal Go packages;
- replace RFC semantic versioning with executable release versioning.
