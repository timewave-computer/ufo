{
  description = "UFO - Universal Fast Orderer";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    osmosis-patches.url = "path:/Users/hxrts/projects/timewave/ufo/patches/osmosis";
  };

  outputs = { self, nixpkgs, flake-utils, osmosis-patches }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        
        # Import the modular nix files
        packages-module = import ./nix/packages.nix { inherit pkgs; };
        python-module = import ./nix/python.nix { inherit pkgs; };
        scripts-module = import ./nix/scripts.nix { inherit pkgs; inherit self; };
       
        # Extract values from modules 
        ufo-binaries = packages-module.ufo-binaries;
        all-binaries = packages-module.all-binaries;
        ufo-test-runner = packages-module.ufo-test-runner;
        
        python-viz-env = python-module.python-viz-env;
        jupyter-app = python-module.jupyter-app;
        
        build-osmosis-ufo-script = scripts-module.build-osmosis-ufo-script;
        
        # Import devShells module and pass required dependencies
        devShells-module = import ./nix/devShells.nix { 
          inherit pkgs;
          inherit python-viz-env;
          inherit jupyter-app;
          inherit build-osmosis-ufo-script;
          hermesPackage = hermesPackage;
        };
        
        # Import apps module and pass required dependencies
        apps-module = import ./nix/apps.nix {
          inherit pkgs;
          inherit self;
          inherit ufo-binaries;
          inherit jupyter-app;
          inherit all-binaries;
          inherit build-osmosis-ufo-script;
        };
        
        # Import test-apps module
        test-apps-module = import ./nix/test-apps.nix {
          inherit pkgs;
          inherit self;
        };
        
        # Add the hermesPackage definition to the flake outputs
        hermesPackage = pkgs.rustPlatform.buildRustPackage rec {
          pname = "hermes";
          version = "1.12.0";
          
          src = pkgs.fetchFromGitHub {
            owner = "informalsystems";
            repo = "hermes";
            rev = "v${version}";
            sha256 = "sha256-zZMqqrHVkBtPf74K7Etf4142CnfDcWc1JHjdjq1G9I4=";
          };
          
          # Using cargoLock instead of cargoSha256
          cargoLock = {
            lockFile = pkgs.fetchurl {
              url = "https://raw.githubusercontent.com/informalsystems/hermes/v${version}/Cargo.lock";
              sha256 = "sha256-ISPdaQvB1UuDhJtOyg8YQH9jSx9Cq3DQAbQWh61satA=";
            };
          };
          
          nativeBuildInputs = with pkgs; [ 
            pkg-config 
            rustPlatform.bindgenHook
            protobuf  # Add protobuf compiler for namada_tx
          ];
          
          buildInputs = with pkgs; [
            openssl
            libiconv
          ] ++ pkgs.lib.optionals pkgs.stdenv.isDarwin [
            pkgs.darwin.apple_sdk.frameworks.Security
            pkgs.darwin.apple_sdk.frameworks.SystemConfiguration
            pkgs.darwin.libobjc
            pkgs.darwin.apple_sdk.frameworks.CoreFoundation
            pkgs.darwin.apple_sdk.frameworks.CoreServices
          ];
          
          # Additional environment variables for macOS builds
          env = pkgs.lib.optionalAttrs pkgs.stdenv.isDarwin {
            OPENSSL_DIR = "${pkgs.openssl.dev}";
            OPENSSL_LIB_DIR = "${pkgs.openssl.out}/lib";
            OPENSSL_INCLUDE_DIR = "${pkgs.openssl.dev}/include";
          };
          
          # Skip tests during build
          doCheck = false;
        };
        
      in
      {
        packages = {
          inherit (packages-module) ufo-binaries all-binaries ufo-test-runner;
          default = ufo-binaries;
          
          # Expose Python environment with visualization packages
          python-viz = python-viz-env;
          
          # Expose Jupyter application
          jupyter = jupyter-app;
        };
        
        # Set up dev shells from module
        devShells = devShells-module;
        
        # Expose runnable commands by merging the apps and test-apps modules
        apps = apps-module // test-apps-module;
      }
    );
}
