# apps.nix - Basic apps for running UFO
{ pkgs, self, ufo-binaries, jupyter-app, all-binaries, build-osmosis-ufo-script }:

{
  default = {
    type = "app";
    program = "${ufo-binaries}/bin/ufo";
  };
  
  # Expose all the binaries as apps
  ufo = {
    type = "app";
    program = "${ufo-binaries}/bin/ufo";
  };
  
  fauxmosis-comet = {
    type = "app";
    program = "${ufo-binaries}/bin/fauxmosis-comet";
  };
  
  fauxmosis-ufo = {
    type = "app";
    program = "${ufo-binaries}/bin/fauxmosis-ufo";
  };
  
  # Add build-all app to build all binaries directly with Go
  build-all = {
    type = "app";
    program = toString (pkgs.writeShellScript "build-all" ''
      cd ${self}
      echo "Building all UFO binaries..."
      
      # Create build directory structure
      mkdir -p build
      
      # Set up the Go environment
      export GOROOT="${pkgs.go_1_22}/share/go"
      export PATH="$GOROOT/bin:$PATH"
      export CGO_ENABLED=1
      export GO111MODULE=on
      export GOPROXY="https://proxy.golang.org,direct"
      
      # Mock implementation binaries
      echo "Building mock implementation binaries..."
      go build -o build/fauxmosis-comet ./cmd/fauxmosis-comet
      go build -o build/fauxmosis-ufo ./cmd/fauxmosis-ufo
      go build -o build/ufo ./main.go
      
      # Integration approach binaries
      echo "Building integration approach binaries..."
      go build -o build/osmosis-comet ./cmd/osmosis-comet
      go build -o build/osmosis-ufo-patched ./cmd/osmosis-ufo-patched
      go build -o build/osmosis-ufo-bridged ./cmd/osmosis-ufo-bridged
      
      echo ""
      echo "All binaries built successfully in build directory:"
      ls -la build/
      
      # Create a symlink from result to build to maintain the expected path
      rm -f result
      ln -sf build result
      
      echo ""
      echo "The binaries are accessible at: result/"
    '');
  };
  
  # Add the benchmarking script as an app
  benchmark = {
    type = "app";
    program = toString (pkgs.writeShellScript "benchmark" ''
      cd ${self}
      ${pkgs.bash}/bin/bash benchmark_assay/run_performance_tests.sh "$@"
    '');
  };
  
  jupyter = {
    type = "app";
    program = toString (pkgs.writeShellScript "jupyter" ''
      cd ${self}
      ${jupyter-app}/bin/jupyter "$@"
    '');
  };
  
  build-osmosis-ufo = {
    type = "app";
    program = toString (pkgs.writeShellScript "build-osmosis-ufo" ''
      ${build-osmosis-ufo-script}/bin/build-osmosis-ufo "$@"
    '');
  };
} 