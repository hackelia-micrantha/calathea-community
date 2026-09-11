# Calathea v0.1.0-alpha.1

This is the first packaged Calathea Community CLI prerelease.

## Included

- public `calathea` executable and stable `version` process anchor;
- standalone Nix flake package, app, checks, development shell, and formatter;
- release-time version injection while ordinary development builds continue to identify as `dev`;
- deterministic release archives for Linux and macOS on amd64 and arm64;
- per-archive SHA-256 checksums;
- deterministic SPDX release metadata and GitHub build/SBOM attestations;
- explicit immutable-tag release workflow following the Micrantha release boundary;
- release-tag creation gated on successful CI for the exact current `main` commit.

## Scope

This prerelease validates packaging, distribution, and consumer integration for the public deterministic community kernel. It does **not** claim completion of private Calathea product milestones, strategy, or higher-level project-management workflows.

## Integration

Nix consumers should use the public flake through the immutable release tag, for example:

```nix
calathea.url = "github:hackelia-micrantha/calathea-community/v0.1.0-alpha.1";
```

The consumer's `flake.lock` is expected to bind the exact release commit and NAR content identity. Consumer repositories remain responsible for their own compatibility and deployment checks.
