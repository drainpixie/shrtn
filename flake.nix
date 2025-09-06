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

    nixosModules.shrtn = {
      config,
      lib,
      pkgs,
      ...
    }: let
      cfg = config.services.shrtn;
    in {
      options.services.shrtn = {
        enable = lib.mkEnableOption "shrtn";

        host = lib.mkOption {
          type = lib.types.str;
          default = "127.0.0.1";
          description = "Host shrtn listens on";
        };

        port = lib.mkOption {
          type = lib.types.int;
          default = 3000;
          description = "Port shrtn listens on";
        };

        database = lib.mkOption {
          type = lib.types.str;
          default = "/var/lib/shrtn/shrtn.db";
          description = "Path to the SQLite database file";
        };
      };

      config = lib.mkIf cfg.enable {
        systemd.services.shrtn = {
          description = "shrtn";
          after = ["network.target"];
          wantedBy = ["multi-user.target"];

          serviceConfig = {
            ExecStart = "${pkgs.shrtn}/bin/shrtn";
            Restart = "always";
            DynamicUser = true;
            StateDirectory = "shrtn";

            Environment = [
              "HOST=${cfg.host}"
              "PORT=${cfg.port}"
              "DATABASE=${cfg.database}"
            ];
          };
        };
      };
    };

    overlays.default = final: prev: {
      shrtn = self.packages.${prev.system}.default;
    };

    formatter = forAllSystems (
      system:
        nixpkgs.legacyPackages.${system}.alejandra
    );
  };
}
