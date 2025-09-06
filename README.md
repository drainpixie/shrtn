![shrtn](https://socialify.git.ci/drainpixie/shrtn/image?description=1&font=Raleway&logo=https%3A%2F%2Fgo.dev%2Fimages%2Fgophers%2Fblue.svg&name=1&pattern=Circuit+Board&theme=Light)

## install

non-nix

```sh
$ git clone https://git.sr.ht/~pixie/shrtn
$ GOBIN=/usr/local/bin go install
```

nix

```nix
{
  inputs.shrtn.url = "github:yourname/shrtn";

  outputs = { self, nixpkgs, shrtn, ... }: {
    nixosConfigurations.host = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        shrtn.nixosModules.shrtn
        {
          services.shrtn.enable = true;
          services.shrtn.port = 8080;
        }
      ];
    };
  };
}
```

## cli

- [timeline](./systems/timeline) dell latitude 5490
- [incubator](./systems/incubator) netcup 500 g11s

## install

## layout

- `lib/` -> custom functions
- `pkgs/` -> custom derivations
- `overlays/` -> custom overlays
- `secrets/` -> secrets managed via `agenix`
- `systems/<hostname>` -> system-specific configuration
- `modules/` -> mixed `NixOS` and `home-manager` modules
