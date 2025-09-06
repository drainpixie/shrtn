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

## usage

```
$ PORT=3005 shrtn
11:39AM INF server running addr=http://localhost:3005
11:39AM INF incoming request method=GET path=/

$ curl http://localhost:3005
shrtn @ http://localhost:3005
=============================
a small url shortener

get      /         index        ex: this website
get      /<id>     redirect     ex: https://google.com
post     /<url>    shorten url  ex: <id>

links    5
clicks   8
version  dev

$ shrtn get <id>
$ shrtn key <token>
$ shrtn all
$ shrtn del <id>
$ shrtn shorten <url>
```
