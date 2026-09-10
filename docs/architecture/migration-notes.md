# Public-kernel boundary migration notes

This branch implements `#48`, coordinated with private `hackelia-micrantha/calathea#59`.

## Intent

`calathea-community` remains the independently useful public implementation and compatibility surface, but it is no longer the canonical home for Calathea's complete product strategy.

The current branch contracts broad product documents and RFCs into minimum public-kernel contracts while preserving historical filenames where existing links may rely on them.

## History

This is a prospective source-of-truth correction. It does not rewrite git history; earlier broader public product documents remain visible in historical commits.

## Compatibility

The migration is documentation/authority-focused. It does not intentionally change the current Go module identity, `calathea` executable name, or supported process boundary.

Any future implementation change required to align code with a narrowed public contract must be tracked independently and validated as a compatibility change.

## Private relationship

Private Calathea consumes this repository through supported process/file/schema surfaces. It must not vendor or import public `internal/` packages merely to recover product semantics that now live privately.
