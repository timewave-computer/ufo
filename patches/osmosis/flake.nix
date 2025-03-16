{
  description = "Osmosis patches for UFO integration";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    osmosis-src = {
      url = "github:osmosis-labs/osmosis";
      flake = false;
    };
  };

  outputs = { self, nixpkgs, flake-utils, osmosis-src }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages = {
          default = self.packages.${system}.osmosis-patched;
          
          osmosis-patched = pkgs.stdenv.mkDerivation {
            name = "osmosis-patched";
            src = osmosis-src;
            
            nativeBuildInputs = [
              pkgs.git
              pkgs.go_1_22
            ];
            
            phases = [ "unpackPhase" "patchPhase" "buildPhase" "installPhase" ];
            
            patchPhase = ''
              echo "Applying UFO integration patches..."
              patch -p1 < ${./ufo.patch}
            '';
            
            buildPhase = ''
              export HOME=$(mktemp -d)
              go build -o osmosis-ufo ./cmd/osmosisd
            '';
            
            installPhase = ''
              mkdir -p $out/bin
              cp osmosis-ufo $out/bin/
            '';
          };
        };
        
        devShell = pkgs.mkShell {
          buildInputs = [
            pkgs.go_1_22
            pkgs.git
          ];
        };
      }
    );
} 