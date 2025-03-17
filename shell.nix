{ pkgs ? import <nixpkgs> {} }:

let
  # Import our custom hermes flake
  hermesFlake = import ./nix/hermes.nix;
  # Get the hermes package for the current system
  hermes = (pkgs.callPackage hermesFlake {}).packages.${builtins.currentSystem}.hermes;
in

pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    hermes  # Use our custom hermes from the flake
    jq
    curl
    gnused
    coreutils
  ];

  shellHook = ''
    export UFO_PATH=${builtins.toString ./.}
    export GO111MODULE=on
    export GOPATH=$HOME/go
    export PATH=$GOPATH/bin:$PATH
    
    echo "UFO Development Environment"
    echo "Go version: $(go version)"
    echo "Hermes binary: $(which hermes)"
    echo "Hermes version: $(hermes --version 2>/dev/null || echo 'Not available')"
    
    # Make sure our Hermes can be found easily in tests
    export HERMES_BIN="$(which hermes)"
  '';
} 