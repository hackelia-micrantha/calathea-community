{
  description = "Calathea community-kernel development, CI, and CLI package";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/10cbfc4edbf7ef7000c8b50d2a581554c6b55644";
  };

  outputs = { self, nixpkgs }:
    let
      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs supportedSystems (system: f (import nixpkgs { inherit system; }));
      version = nixpkgs.lib.strings.removeSuffix "\n" (builtins.readFile ./VERSION);
    in {
      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            gh
            mise
            python3
          ];

          shellHook = ''
            export GOTMPDIR="$PWD/.cache/go-tmp"
            mkdir -p "$GOTMPDIR"
          '';
        };
      });

      packages = forAllSystems (pkgs:
        let
          calathea = pkgs.buildGoModule {
            pname = "calathea";
            inherit version;
            src = self;
            vendorHash = null;
            subPackages = [ "cmd/calathea" ];
            ldflags = [
              "-s"
              "-w"
              "-X github.com/hackelia-micrantha/calathea-community/internal/application.Version=${version}"
            ];
            doInstallCheck = true;
            installCheckPhase = ''
              runHook preInstallCheck
              test "$("$out/bin/calathea" version)" = "calathea ${version}"
              runHook postInstallCheck
            '';
            meta = with pkgs.lib; {
              description = "Deterministic local-first project-orientation kernel CLI";
              homepage = "https://github.com/hackelia-micrantha/calathea-community";
              license = licenses.mpl20;
              mainProgram = "calathea";
              platforms = platforms.unix;
            };
          };
        in {
          inherit calathea;
          default = calathea;
        });

      apps = forAllSystems (pkgs:
        let
          system = pkgs.stdenv.hostPlatform.system;
          calatheaApp = {
            type = "app";
            program = "${self.packages.${system}.calathea}/bin/calathea";
          };
        in {
          calathea = calatheaApp;
          default = calatheaApp;
        });

      checks = forAllSystems (pkgs:
        let
          system = pkgs.stdenv.hostPlatform.system;
        in {
          calathea = self.packages.${system}.calathea;
          default = self.packages.${system}.calathea;
        });

      formatter = forAllSystems (pkgs: pkgs.nixfmt-rfc-style);
    };
}
