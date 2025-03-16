{ pkgs ? import <nixpkgs> {} }:

with pkgs;

buildGoModule rec {
  pname = "ufo";
  version = "0.1.0";
  src = ./.;

  vendorHash = null;

  nativeBuildInputs = [ makeWrapper ];

  buildInputs = [ 
    # Add any additional dependencies here
  ];

  buildPhase = ''
    runHook preBuild
    go build -o ufo ./src/cmd/ufo
    go build -o osmosis-ufo ./src/cmd/osmosis-ufo
    runHook postBuild
  '';

  installPhase = ''
    runHook preInstall
    mkdir -p $out/bin
    cp ufo $out/bin/
    cp osmosis-ufo $out/bin/
    runHook postInstall
  '';

  meta = {
    description = "Universal Fast Orderer - A lightweight alternative to CometBFT for Cosmos applications";
    homepage = "https://github.com/timewave/ufo";
    license = licenses.mit;
    platforms = platforms.unix;
  };
} 