{
  description = "Flake for Hermes IBC relayer";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        
        hermes = pkgs.rustPlatform.buildRustPackage rec {
          pname = "hermes";
          version = "1.7.4";
          
          src = pkgs.fetchFromGitHub {
            owner = "informalsystems";
            repo = "hermes";
            rev = "v${version}";
            hash = "sha256-JTZMp4By/pGsMdKzfi4H1LQS1RKYQHBq5NEju5ADX/s=";
          };
          
          cargoHash = "sha256-/9U5c3h4R1GYuaUcY3o5s1frfF+LB5XTR4CmIfkDT+4=";
          
          nativeBuildInputs = [ pkgs.pkg-config ];
          
          buildInputs = with pkgs; [
            openssl
            libiconv
          ] ++ pkgs.lib.optionals pkgs.stdenv.isDarwin [
            pkgs.darwin.apple_sdk.frameworks.Security
            pkgs.darwin.apple_sdk.frameworks.SystemConfiguration
          ] ++ pkgs.lib.optionals pkgs.stdenv.isLinux [
            # Linux specific dependencies
            pkgs.glibc
          ];
          
          # Environment variables for platform-specific builds
          env = pkgs.lib.optionalAttrs pkgs.stdenv.isDarwin {
            OPENSSL_DIR = "${pkgs.openssl.dev}";
            OPENSSL_LIB_DIR = "${pkgs.openssl.out}/lib";
            OPENSSL_INCLUDE_DIR = "${pkgs.openssl.dev}/include";
          } // pkgs.lib.optionalAttrs pkgs.stdenv.isLinux {
            OPENSSL_DIR = "${pkgs.openssl.dev}";
            OPENSSL_LIB_DIR = "${pkgs.openssl.out}/lib";
            OPENSSL_INCLUDE_DIR = "${pkgs.openssl.dev}/include";
          };
          
          # Skip tests during build
          doCheck = false;
          
          meta = with pkgs.lib; {
            description = "Implementation of an IBC relayer for the Cosmos ecosystem";
            homepage = "https://hermes.informal.systems";
            license = licenses.asl20;
            platforms = with platforms; linux ++ darwin;
          };
        };
        
      in {
        packages = {
          inherit hermes;
          default = hermes;
        };
        
        devShell = pkgs.mkShell {
          buildInputs = [
            hermes
            pkgs.go
          ];
          
          shellHook = ''
            echo "Hermes IBC Relayer Development Environment"
            echo "Hermes version: $(hermes --version 2>/dev/null || echo 'Not available')"
            
            # Platform-specific setup
            ${if pkgs.stdenv.isDarwin then ''
              echo "Running on Darwin (macOS)"
            '' else if pkgs.stdenv.isLinux then ''
              echo "Running on Linux"
            '' else ''
              echo "Running on unsupported platform"
            ''}
          '';
        };
      }
    );
} 