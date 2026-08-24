{
  description = "Calathea development and CI environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/10cbfc4edbf7ef7000c8b50d2a581554c6b55644";
  };

  outputs = { self, nixpkgs }:
    let
      supportedSystems = [ "x86_64-linux" ];
      forAllSystems = f: nixpkgs.lib.genAttrs supportedSystems (system: f (import nixpkgs { inherit system; }));
    in {
      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            mise
            staticcheck
            govulncheck
          ];

          shellHook = ''
            export GOTMPDIR="$PWD/.cache/go-tmp"
            mkdir -p "$GOTMPDIR"
          '';
        };
      });

      formatter = forAllSystems (pkgs: pkgs.nixfmt-rfc-style);
    };
}
