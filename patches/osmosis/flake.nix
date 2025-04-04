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
              pkgs.go_1_23
            ];
            
            phases = [ "unpackPhase" "patchPhase" "buildPhase" "installPhase" ];
            
            patchPhase = ''
              echo "Applying UFO integration patches..."
              patch -p1 < ${./ufo.patch}
            '';
            
            buildPhase = ''
              export HOME=$(mktemp -d)
              
              # Platform-specific setup
              ${if pkgs.stdenv.isDarwin then ''
                # macOS specific setup
                echo "Building Osmosis on Darwin (macOS)"
              '' else if pkgs.stdenv.isLinux then ''
                # Linux specific setup
                echo "Building Osmosis on Linux (x86)"
                export CGO_ENABLED=1
                export CGO_CFLAGS="-I${pkgs.glibc.dev}/include"
                export CGO_LDFLAGS="-L${pkgs.glibc.out}/lib"
                export LD_LIBRARY_PATH="${pkgs.stdenv.cc.cc.lib}/lib:${pkgs.zlib}/lib:$LD_LIBRARY_PATH"
              '' else ''
                echo "Building Osmosis on unsupported platform"
              ''}
              
              go build -o osmosis-ufo ./cmd/osmosisd
            '';
            
            installPhase = ''
              mkdir -p $out/bin
              cp osmosis-ufo $out/bin/
            '';
            
            meta = with pkgs.lib; {
              description = "Osmosis with UFO integration patches";
              homepage = "https://github.com/osmosis-labs/osmosis";
              license = licenses.asl20;
              platforms = with platforms; linux ++ darwin;
            };
          };
        };
        
        devShell = pkgs.mkShell {
          buildInputs = [
            pkgs.go_1_23
            pkgs.git
          ] ++ pkgs.lib.optionals pkgs.stdenv.isLinux [
            # Linux specific dependencies
            pkgs.glibc.dev
          ];
          
          shellHook = ''
            export GOROOT="${pkgs.go_1_23}/share/go"
            export PATH="$GOROOT/bin:$PATH"
            export CGO_ENABLED=1
            
            # Platform-specific setup
            ${if pkgs.stdenv.isDarwin then ''
              # macOS specific setup
              echo "Osmosis dev shell on Darwin (macOS)"
            '' else if pkgs.stdenv.isLinux then ''
              # Linux specific setup
              echo "Osmosis dev shell on Linux (x86)"
              export CGO_CFLAGS="-I${pkgs.glibc.dev}/include"
              export CGO_LDFLAGS="-L${pkgs.glibc.out}/lib"
              export LD_LIBRARY_PATH="${pkgs.stdenv.cc.cc.lib}/lib:${pkgs.zlib}/lib:$LD_LIBRARY_PATH"
            '' else ''
              echo "Osmosis dev shell on unsupported platform"
            ''}
          '';
        };
      }
    );
} 