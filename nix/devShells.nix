# devShells.nix - Development shells for UFO
{ pkgs, python-viz-env, jupyter-app, build-osmosis-ufo-script, hermesPackage ? null }:

{
  default = pkgs.mkShell {
    buildInputs = with pkgs; [
      go_1_23
      gopls
      golangci-lint
      delve
      gotools
      go-tools
    ];
    
    shellHook = ''
      # Set up the Go environment
      export GOROOT="${pkgs.go_1_23}/share/go"
      export PATH="$GOROOT/bin:$PATH"
      export CGO_ENABLED=1
      export GO111MODULE=on
      
      # Function to build the UFO binaries directly with Go
      build_binaries() {
        echo "Building UFO binaries..."
        
        # Use the central build script if it exists
        if [ -f "./scripts/build_binaries.sh" ]; then
          ./scripts/build_binaries.sh
        else
          # Create bin directory
          mkdir -p bin
          
          # Build the core binaries
          go build -o bin/fauxmosis-comet ./cmd/fauxmosis-comet
          go build -o bin/fauxmosis-ufo ./cmd/fauxmosis-ufo
          go build -o bin/ufo ./main.go
          
          # Create a symlink from result to bin to maintain the expected path
          rm -f result
          ln -sf bin result
        fi
        
        echo "UFO binaries built successfully at: ./bin/"
        ls -la ./bin/
      }
      
      # Function to build all UFO binaries including all integration approaches
      build_all_binaries() {
        echo "Building all UFO binaries and integration approaches..."
        
        # Use the central build script if it exists
        if [ -f "./scripts/build_binaries.sh" ]; then
          ./scripts/build_binaries.sh
        else
          # Create bin directory
          mkdir -p bin
          
          # Mock implementation binaries
          echo "Building mock implementation binaries..."
          go build -o bin/fauxmosis-comet ./cmd/fauxmosis-comet
          go build -o bin/fauxmosis-ufo ./cmd/fauxmosis-ufo
          go build -o bin/ufo ./main.go
          
          # Integration approach binaries
          echo "Building integration approach binaries..."
          go build -o bin/osmosis-comet ./cmd/osmosis-comet
          go build -o bin/osmosis-ufo-patched ./cmd/osmosis-ufo-patched
          go build -o bin/osmosis-ufo-bridged ./cmd/osmosis-ufo-bridged
          
          # Create a symlink from result to bin to maintain the expected path
          rm -f result
          ln -sf bin result
        fi
        
        echo "All binaries built successfully at: ./bin/"
        ls -la ./bin/
      }
      
      # Function to run tests with result binaries on PATH
      run_tests() {
        if [ ! -d "result" ]; then
          echo "The result directory with binaries does not exist yet."
          echo "Please build binaries first using: build_binaries or build_all_binaries"
          return 1
        fi
        
        # Make sure the binaries are on PATH during tests
        export PATH="$(pwd)/result:$PATH"
        
        # Run the tests
        echo "Running tests with binaries from result..."
        go test ./tests/...
      }
      
      export -f build_binaries
      export -f build_all_binaries
      export -f run_tests
      
      echo "Available commands:"
      echo "  build_binaries         - Build the UFO binaries"
      echo "  build_all_binaries     - Build all UFO binaries including integration approaches"
      echo "  run_tests              - Run tests with binaries from result"
      echo "  ufo-jupyter            - Launch Jupyter with UFO benchmark environment"
      echo "  build-osmosis-ufo      - Build Osmosis with UFO integration"
      echo "  python                 - The Python interpreter with visualization packages"
    '';
    
    # Allow network access for go
    allowNetworkAccess = true;
    
    # Include the Python environment for visualization
    packages = [ python-viz-env jupyter-app build-osmosis-ufo-script ];
  };
  
  # Test shell for running benchmarks and tests
  test = pkgs.mkShell {
    buildInputs = with pkgs; [
      go_1_23
      gotools
    ];
    
    shellHook = ''
      # Set up the Go environment
      export GOROOT="${pkgs.go_1_23}/share/go"
      export PATH="$GOROOT/bin:$PATH"
      export CGO_ENABLED=1
      export GO111MODULE=on
      export GOFLAGS="-modcacherw"
      
      # Function to build test binaries directly with Go
      build_test_binaries() {
        echo "Building test binaries..."
        
        # Create build directory
        mkdir -p build
        
        # Build core test binaries
        go build -o build/fauxmosis-comet ./cmd/fauxmosis-comet
        go build -o build/fauxmosis-ufo ./cmd/fauxmosis-ufo
        go build -o build/ufo ./main.go
        
        # Create a symlink from result to build to maintain the expected path
        rm -f result
        ln -sf build result
        
        echo "Test binaries built successfully at: ./result/"
        ls -la ./result/
      }
      
      # Function to build all UFO binaries including all integration approaches
      build_all_binaries() {
        echo "Building all UFO binaries and integration approaches..."
        
        # Create build directory
        mkdir -p build
        
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
        
        # Create a symlink from result to build to maintain the expected path
        rm -f result
        ln -sf build result
        
        echo "All binaries built successfully at: ./result/"
        ls -la ./result/
      }
      
      # Function to run tests with result binaries on PATH
      run_tests() {
        if [ ! -d "result" ]; then
          echo "The result directory with binaries does not exist yet."
          echo "Please build binaries first using: build_test_binaries or build_all_binaries"
          return 1
        fi
        
        # Make sure the binaries are on PATH during tests
        export PATH="$(pwd)/result:$PATH"
        
        # Run the tests
        echo "Running tests with binaries from result..."
        go test ./tests/...
      }
      
      export -f build_test_binaries
      export -f build_all_binaries
      export -f run_tests
      
      # Configure Hermes
      export HERMES_CONFIG=./tests/config/hermes.toml
      
      echo "Available commands:"
      echo "  build_test_binaries    - Build the test binaries"
      echo "  build_all_binaries     - Build all UFO binaries including integration approaches"
      echo "  run_tests              - Run tests with binaries from result"
      echo "  go test ./...          - Run all tests"
      echo "  cd tests && go test    - Run tests from tests directory"
    '';
    
    allowNetworkAccess = true;
  };
  
  # IBC test shell with our custom Hermes package
  ibc = pkgs.mkShell {
    buildInputs = with pkgs; [
      go_1_23
      gotools
      jq
      curl
      gnused
      coreutils
    ] ++ (if hermesPackage != null then [ hermesPackage ] else []);
    
    shellHook = ''
      # Set up the Go environment
      export GOROOT="${pkgs.go_1_23}/share/go"
      export PATH="$GOROOT/bin:$PATH"
      export CGO_ENABLED=1
      export GO111MODULE=on
      export GOFLAGS="-modcacherw"
      
      # Function to build test binaries including all integration approaches
      build_test_binaries() {
        echo "Building test binaries for IBC testing..."
        
        # Use the central build script if it exists
        if [ -f "./scripts/build_binaries.sh" ]; then
          ./scripts/build_binaries.sh
        else
          # Create bin directory
          mkdir -p bin
          
          # Build binaries for IBC testing
          go build -o bin/osmosis-ufo-patched ./cmd/osmosis-ufo-patched
          go build -o bin/osmosis-ufo-bridged ./cmd/osmosis-ufo-bridged
          
          # Create a symlink from result to bin
          rm -f result
          ln -sf bin result
        fi
        
        echo "IBC test binaries built successfully at: ./bin/"
        ls -la ./bin/
      }
      
      # Function to run IBC tests
      run_ibc_tests() {
        local binary_type=''${1:-"all"}
        shift 2>/dev/null || true
        
        if [[ "''$binary_type" != "patched" && "''$binary_type" != "bridged" && "''$binary_type" != "all" ]]; then
          echo "Usage: run_ibc_tests [patched|bridged|all] [test flags]"
          echo "  patched - Use osmosis-ufo-patched binary"
          echo "  bridged - Use osmosis-ufo-bridged binary"
          echo "  all     - Run tests with both binaries"
          return 1
        fi
        
        # Ensure the binaries are built
        if [ ! -d "result" ]; then
          echo "Building binaries first..."
          build_test_binaries
        fi
        
        # Make sure the binaries are on PATH during tests
        export PATH="''$(pwd)/result:''$PATH"
        
        # Set Hermes environment variables
        if [ -x "''$(which hermes)" ]; then
          export HERMES_BIN="''$(which hermes)"
          echo "Using Hermes at: ''$HERMES_BIN"
        else
          echo "Warning: Hermes binary not found"
        fi
        
        # Run the IBC tests
        if [[ "''$binary_type" == "all" ]]; then
          echo "Running tests with PATCHED binary..."
          UFO_BINARY_TYPE=patched go test -v ./tests/ibc "''$@"
          
          echo -e "\n=====================================\n"
          
          echo "Running tests with BRIDGED binary..."
          UFO_BINARY_TYPE=bridged go test -v ./tests/ibc "''$@"
        else
          echo "Running tests with ''$binary_type binary..."
          UFO_BINARY_TYPE=''$binary_type go test -v ./tests/ibc "''$@"
        fi
      }
      
      export -f build_test_binaries
      export -f run_ibc_tests
      
      echo "IBC Testing Environment"
      echo "Available commands:"
      echo "  build_test_binaries    - Build the IBC test binaries"
      echo "  run_ibc_tests [type]   - Run IBC tests with patched|bridged|all binary types"
      
      if [ -x "''$(which hermes)" ]; then
        echo "Hermes version: ''$(hermes version 2>/dev/null || echo 'Not available')"
        export HERMES_BIN="''$(which hermes)"
      else
        echo "Warning: Hermes binary not found"
      fi
    '';
    
    allowNetworkAccess = true;
  };
} 