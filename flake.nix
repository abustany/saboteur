{
  description = "Saboteur";

  inputs = {
    nixpkgs.url      = "github:NixOS/nixpkgs/bec27fabee7ff51a4788840479b1730ed1b64427";
    flake-utils.url  = "github:numtide/flake-utils/919d646de7be200f3bf08cb76ae1f09402b6f9b4";
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        overlays = [];
        pkgs = import nixpkgs {
          inherit system overlays;
        };
        rev = if (self ? shortRev) then self.shortRev else "dev";
      in
      with pkgs;
      {
        devShells.default = pkgs.mkShell {
          buildInputs = [
            # backend
            pkgs.go
            pkgs.gopls
          ];

          shellHook = ''
          export GOPATH="$(realpath .)/.go";
          export PATH="''\${GOPATH}/bin:''\${PATH}"
          '';
        };

        packages.default = pkgs.buildGo120Module {
          pname = "saboteur";
          version = rev;
          src = pkgs.lib.cleanSource self;
          vendorHash = "sha256-Fjm+bCGg9aBdSZxBo7rnGYmfiALaNlG2bK6VGGlt/TU=";
        };
      }
    );
}
