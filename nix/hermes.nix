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
          ];
          
          # Skip tests during build
          doCheck = false;
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
        };
      }
    );
} 