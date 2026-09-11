# Packaging and release

## Status

This document defines the standalone package, release, and consumer-integration contract for the public Calathea Community CLI. It does not expand the CLI/process compatibility contract or transfer private Calathea product strategy into the public repository.

The first packaged build is intentionally a prerelease: `v0.1.0-alpha.1`.

The boundary follows the Micrantha distribution strategy: Calathea Community owns its public-kernel build artifacts and releases; consumers own their version pin, compatibility validation, and deployment. No meta, private-product, or host repository becomes a build dependency for the public kernel.

## Version identity

`VERSION` is the canonical application-package version for a reviewed release candidate.

Ordinary development builds retain the process-contract identity:

```text
calathea dev
```

Release/package builds inject `VERSION` into `internal/application.Version` at link time and must report:

```text
calathea <VERSION>
```

Application-package versions remain independent from persisted-schema, evaluation, policy, orientation, and machine-output semantic versions.

## Standalone Nix contract

The repository flake supports these systems:

- `x86_64-linux`
- `aarch64-linux`
- `x86_64-darwin`
- `aarch64-darwin`

Canonical outputs are:

```text
packages.<system>.default
packages.<system>.calathea
apps.<system>.default
apps.<system>.calathea
checks.<system>.default
checks.<system>.calathea
devShells.<system>.default
formatter.<system>
```

Typical standalone use:

```sh
nix build .#calathea
nix run .#calathea -- version
nix flake check
```

The package install check executes the built binary and requires its reported version to match `VERSION` exactly.

The Nix package is repository-owned and has no dependency on private Calathea data, Dubnium host configuration, credentials, or effect authority.

## Release publication

Release publication uses an immutable reviewed tag rather than a mutable release branch.

For version `<VERSION>`:

1. update `VERSION` and `RELEASE_NOTES_v<VERSION>.md` through ordinary review;
2. merge the reviewed release/package changes to `main`;
3. require a successful `CI` **push** run whose `head_sha` is exact current `main`;
4. explicitly dispatch `Create Release Tag`, which requires the target to resolve to that exact current `main`, verifies the successful exact-main CI evidence, and requires an operator confirmation;
5. that workflow validates the reviewed version metadata and creates annotated immutable tag `v<VERSION>`;
6. the tag-triggered `Release` workflow independently proves the tag commit is the triggered commit and remains on `main`;
7. the release workflow reruns the repository quality gate and standalone Nix package gate at the immutable tag;
8. only after those gates pass, create or reuse a **draft** GitHub prerelease;
9. build deterministic target archives, per-target SHA-256 checksums, and SPDX metadata from the tagged source timestamp;
10. attest each archive's build provenance and SPDX metadata;
11. upload all immutable release inputs while the release remains draft;
12. publish the prerelease only after every target build and attestation succeeds.

This exact-main-CI check is intentionally redundant with normal merge policy. It prevents release-tag creation if repository branch settings or auto-merge behavior admit a commit before its canonical post-merge CI has completed successfully.

`v0.1.0-alpha.1` is published as a prerelease. A later stable version can use the same boundary with an explicit publication-policy change rather than weakening provenance checks.

The tag is the canonical immutable source identity. GitHub release assets are derived distribution artifacts for non-Nix or direct-download consumers.

## Release assets

The initial release publishes pure-Go CLI archives for:

```text
linux/amd64
linux/arm64
darwin/amd64
darwin/arm64
```

Archive names are:

```text
calathea_<VERSION>_<os>_<arch>.tar.gz
```

Each deterministic archive contains:

- `calathea`
- `LICENSE`
- `README.md`
- `SBOM.spdx.json`

For each archive the release also publishes:

- `<archive>.sha256`;
- target-specific SPDX JSON;
- GitHub build provenance attestation;
- GitHub SBOM attestation.

Release archive timestamps are derived from the tagged source commit and archive ownership/modes are normalized by `scripts/release.py`.

## Consumer integration

The preferred Micrantha integration for a Nix consumer is the released public flake itself, pinned by immutable tag and lockfile.

Example:

```nix
inputs.calathea = {
  url = "github:hackelia-micrantha/calathea-community/v0.1.0-alpha.1";
  inputs.nixpkgs.follows = "nixpkgs";
};
```

The consumer commits its `flake.lock`, which resolves that human-readable tag to an exact source revision and NAR content identity. The consumer then selects `calathea.packages.${system}.default` or `.calathea` and owns any host/profile wiring.

This deliberately mirrors other Micrantha public-adapter integrations: the producer owns packaging and release provenance; the consumer does **not** repackage the binary merely to install it.

A consumer integration change should prove at least:

- the lock resolves the intended reviewed release tag;
- the package builds through the consumer's Nix graph;
- `calathea version` reports the expected release identity;
- the package is attached only to the intended host/profile;
- existing consumer checks remain green;
- installation grants no additional credential or effect authority.

Compatibility is consumer-owned evidence. A successful Calathea Community release does not by itself prove a particular Dubnium/dotfiles composition is compatible.

## Non-Nix consumers

Consumers without a Nix composition path may instead pin the immutable release archive and verify its published SHA-256/provenance material.

Do not:

- consume `main` or a mutable release-preparation branch as a production package source;
- duplicate Calathea package ownership into a consumer repository when the released flake is usable;
- require private Calathea repositories to build the public CLI;
- grant credentials or external-effect authority merely because the executable is installed;
- conflate application release version with persisted/domain semantic versions;
- treat a public-kernel release as completion evidence for private Calathea product milestones.

Dubnium-specific installation policy belongs to its consumer configuration, not to Calathea Community's reusable package definition.
