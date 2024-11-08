{
  description = "Saboteur";

  inputs = {
    nixpkgs.url      = "github:NixOS/nixpkgs/nixos-24.05";
    flake-utils.url  = "github:numtide/flake-utils";
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

	  GOTOOLCHAIN = "local";

          shellHook = ''
          export GOPATH="$(realpath .)/.go";
          export PATH="''\${GOPATH}/bin:''\${PATH}"
          '';
        };

        packages.default = pkgs.buildGo122Module {
          pname = "saboteur";
          version = rev;
          src = pkgs.lib.cleanSource self;
          vendorHash = "sha256-Kux6WYiBEEFRVQ+JR4d9mo/iqfD65wo2+XS99eKen9k=";
        };
      }
    );
}
