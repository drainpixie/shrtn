{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    hooks.url = "github:cachix/git-hooks.nix";
  };

  outputs = {
    self,
    nixpkgs,
    hooks,
    ...
  }: let
    inherit (nixpkgs) lib;

    systems = [
      "aarch64-linux"
      "i686-linux"
      "x86_64-linux"
      "aarch64-darwin"
      "x86_64-darwin"
    ];

    forAllSystems = lib.genAttrs systems;

    baseVersion = "1.0.0";
    version =
      if self ? shortRev
      then "${baseVersion}+git.${self.shortRev}"
      else "${baseVersion}-dirty";
  in {
    packages = forAllSystems (system: let
      pkgs = nixpkgs.legacyPackages.${system};
    in {
      default = pkgs.buildGoModule {
        inherit version;

        pname = "shrtn";
        src = ./.;

        vendorHash = "sha256-aUOG8VYVc+fwAq+0XngmLapziM9ciGNc00jMeBajvQA=";

        nativeBuildInputs = [pkgs.pkg-config];
        buildInputs = [pkgs.sqlite];

        ldflags = [
          "-s"
          "-w"
          "-X main.version=${version}"
        ];
      };
    });

    checks = forAllSystems (system: let
      lib = hooks.lib.${system};
    in {
      pre-commit-check = lib.run {
        src = ./.;
        hooks = {
          alejandra.enable = true;
          convco.enable = true;
          gofmt.enable = true;
          statix = {
            enable = true;
            settings.ignore = ["/.direnv"];
          };
        };
      };
    });

    devShells = forAllSystems (system: let
      pkgs = nixpkgs.legacyPackages.${system};
      check = self.checks.${system}.pre-commit-check;
    in {
      default = pkgs.mkShell {
        inherit (check) shellHook;

        buildInputs =
          check.enabledPackages
          ++ (builtins.attrValues {
            inherit (pkgs) go sqlite pkg-config air delve gotools;
          });
      };
    });

    formatter = forAllSystems (
      system:
        nixpkgs.legacyPackages.${system}.alejandra
    );
  };
}
