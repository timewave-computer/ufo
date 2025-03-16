# test-apps.nix - Test-related apps for UFO
{ pkgs, self }:

{
  test-all = {
    type = "app";
    program = toString (pkgs.writeShellScript "test-all" ''
      cd ${self}
      echo "Running all tests..."
      
      # Check if result directory exists with binaries
      if [ ! -d "result/bin" ]; then
        echo "The result/bin directory with binaries does not exist yet."
        echo "Please build binaries first using: nix run .#build-all"
        exit 1
      fi
      
      # Check if tests directory exists
      if [ ! -d "tests" ]; then
        echo "The tests directory does not exist yet. Please create it first."
        exit 1
      fi
      
      # Make sure the binaries are on PATH during tests
      export PATH="$(pwd)/result/bin:$PATH"
      
      # Run all tests
      ${pkgs.go_1_22}/bin/go test ./tests/...
    '');
  };
  
  test-core = {
    type = "app";
    program = toString (pkgs.writeShellScript "test-core" ''
      cd ${self}
      echo "Running core interface tests..."
      
      # Check if result directory exists with binaries
      if [ ! -d "result/bin" ]; then
        echo "The result/bin directory with binaries does not exist yet."
        echo "Please build binaries first using: nix run .#build-all"
        exit 1
      fi
      
      # Check if tests/core directory exists
      if [ ! -d "tests/core" ]; then
        echo "The tests/core directory does not exist yet. Please create it first."
        exit 1
      fi
      
      # Make sure the binaries are on PATH during tests
      export PATH="$(pwd)/result/bin:$PATH"
      
      # Run only tests in the tests/core directory, not subdirectories
      cd tests/core
      ${pkgs.go_1_22}/bin/go test -v
    '');
  };
  
  test-consensus = {
    type = "app";
    program = toString (pkgs.writeShellScript "test-consensus" ''
      cd ${self}
      echo "Running consensus tests..."
      
      # Check if result directory exists with binaries
      if [ ! -d "result/bin" ]; then
        echo "The result/bin directory with binaries does not exist yet."
        echo "Please build binaries first using: nix run .#build-all"
        exit 1
      fi
      
      # Check if tests/consensus directory exists
      if [ ! -d "tests/consensus" ]; then
        echo "The tests/consensus directory does not exist yet. Please create it first."
        exit 1
      fi
      
      # Make sure the binaries are on PATH during tests
      export PATH="$(pwd)/result/bin:$PATH"
      
      ${pkgs.go_1_22}/bin/go test ./tests/consensus/...
    '');
  };
  
  test-ibc = {
    type = "app";
    program = toString (pkgs.writeShellScript "test-ibc" ''
      cd ${self}
      echo "Running IBC tests..."
      
      # Check if result directory exists with binaries
      if [ ! -d "result/bin" ]; then
        echo "The result/bin directory with binaries does not exist yet."
        echo "Please build binaries first using: nix run .#build-all"
        exit 1
      fi
      
      # Check if tests/ibc directory exists
      if [ ! -d "tests/ibc" ]; then
        echo "The tests/ibc directory does not exist yet. Please create it first."
        exit 1
      fi
      
      # Make sure the binaries are on PATH during tests
      export PATH="$(pwd)/result/bin:$PATH"
      
      ${pkgs.go_1_22}/bin/go test ./tests/ibc/...
    '');
  };
  
  test-integration = {
    type = "app";
    program = toString (pkgs.writeShellScript "test-integration" ''
      cd ${self}
      echo "Running integration tests..."
      
      # Check if result directory exists with binaries
      if [ ! -d "result/bin" ]; then
        echo "The result/bin directory with binaries does not exist yet."
        echo "Please build binaries first using: nix run .#build-all"
        exit 1
      fi
      
      # Check if tests/integration directory exists
      if [ ! -d "tests/integration" ]; then
        echo "The tests/integration directory does not exist yet. Please create it first."
        exit 1
      fi
      
      # Make sure the binaries are on PATH during tests
      export PATH="$(pwd)/result/bin:$PATH"
      
      ${pkgs.go_1_22}/bin/go test ./tests/integration/...
    '');
  };
  
  test-stress = {
    type = "app";
    program = toString (pkgs.writeShellScript "test-stress" ''
      cd ${self}
      echo "Running stress tests..."
      
      # Check if result directory exists with binaries
      if [ ! -d "result/bin" ]; then
        echo "The result/bin directory with binaries does not exist yet."
        echo "Please build binaries first using: nix run .#build-all"
        exit 1
      fi
      
      # Check if tests/stress directory exists
      if [ ! -d "tests/stress" ]; then
        echo "The tests/stress directory does not exist yet. Please create it first."
        exit 1
      fi
      
      # Make sure the binaries are on PATH during tests
      export PATH="$(pwd)/result/bin:$PATH"
      
      ${pkgs.go_1_22}/bin/go test ./tests/stress/...
    '');
  };
  
  test-mockgen = {
    type = "app";
    program = toString (pkgs.writeShellScript "test-mockgen" ''
      cd ${self}
      echo "Generating mocks for testing..."
      
      # Check if tests directory exists
      if [ ! -d "tests" ]; then
        echo "The tests directory does not exist yet. Please create it first."
        exit 1
      fi
      
      ${pkgs.go_1_22}/bin/go generate ./tests/...
    '');
  };
  
  test-verbose = {
    type = "app";
    program = toString (pkgs.writeShellScript "test-verbose" ''
      cd ${self}
      echo "Running tests with verbose output..."
      
      # Check if result directory exists with binaries
      if [ ! -d "result/bin" ]; then
        echo "The result/bin directory with binaries does not exist yet."
        echo "Please build binaries first using: nix run .#build-all"
        exit 1
      fi
      
      # Check if tests directory exists
      if [ ! -d "tests" ]; then
        echo "The tests directory does not exist yet. Please create it first."
        exit 1
      fi
      
      # Make sure the binaries are on PATH during tests
      export PATH="$(pwd)/result/bin:$PATH"
      
      ${pkgs.go_1_22}/bin/go test -v ./tests/...
    '');
  };
  
  test-race = {
    type = "app";
    program = toString (pkgs.writeShellScript "test-race" ''
      cd ${self}
      echo "Running tests with race detection..."
      
      # Check if result directory exists with binaries
      if [ ! -d "result/bin" ]; then
        echo "The result/bin directory with binaries does not exist yet."
        echo "Please build binaries first using: nix run .#build-all"
        exit 1
      fi
      
      # Check if tests directory exists
      if [ ! -d "tests" ]; then
        echo "The tests directory does not exist yet. Please create it first."
        exit 1
      fi
      
      # Make sure the binaries are on PATH during tests
      export PATH="$(pwd)/result/bin:$PATH"
      
      ${pkgs.go_1_22}/bin/go test -race ./tests/...
    '');
  };
  
  test-cover = {
    type = "app";
    program = toString (pkgs.writeShellScript "test-cover" ''
      cd ${self}
      echo "Running tests with coverage..."
      
      # Check if result directory exists with binaries
      if [ ! -d "result/bin" ]; then
        echo "The result/bin directory with binaries does not exist yet."
        echo "Please build binaries first using: nix run .#build-all"
        exit 1
      fi
      
      # Check if tests directory exists
      if [ ! -d "tests" ]; then
        echo "The tests directory does not exist yet. Please create it first."
        exit 1
      fi
      
      # Make sure the binaries are on PATH during tests
      export PATH="$(pwd)/result/bin:$PATH"
      
      ${pkgs.go_1_22}/bin/go test -cover ./tests/...
    '');
  };
  
  test-cover-html = {
    type = "app";
    program = toString (pkgs.writeShellScript "test-cover-html" ''
      cd ${self}
      echo "Running tests with coverage and generating HTML report..."
      
      # Check if result directory exists with binaries
      if [ ! -d "result/bin" ]; then
        echo "The result/bin directory with binaries does not exist yet."
        echo "Please build binaries first using: nix run .#build-all"
        exit 1
      fi
      
      # Check if tests directory exists
      if [ ! -d "tests" ]; then
        echo "The tests directory does not exist yet. Please create it first."
        exit 1
      fi
      
      # Make sure the binaries are on PATH during tests
      export PATH="$(pwd)/result/bin:$PATH"
      
      ${pkgs.go_1_22}/bin/go test -coverprofile=coverage.out ./tests/...
      ${pkgs.go_1_22}/bin/go tool cover -html=coverage.out -o coverage.html
      echo "Coverage report generated at: $(pwd)/coverage.html"
    '');
  };
  
  test-lint = {
    type = "app";
    program = toString (pkgs.writeShellScript "test-lint" ''
      cd ${self}
      echo "Running linter on tests..."
      
      # Check if tests directory exists
      if [ ! -d "tests" ]; then
        echo "The tests directory does not exist yet. Please create it first."
        exit 1
      fi
      
      ${pkgs.golangci-lint}/bin/golangci-lint run ./tests/...
    '');
  };
  
  test-clean = {
    type = "app";
    program = toString (pkgs.writeShellScript "test-clean" ''
      cd ${self}
      echo "Cleaning test artifacts..."
      
      if [ -d "tests/tmp" ]; then
        rm -rf tests/tmp/
        echo "Test artifacts cleaned."
      else
        echo "No test artifacts to clean."
      fi
    '');
  };
  
  test-help = {
    type = "app";
    program = toString (pkgs.writeShellScript "test-help" ''
      echo "Available UFO test commands (use with 'nix run .#command'):"
      echo ""
      echo "  build-all:        Build all binaries to the result directory (REQUIRED FIRST STEP)"
      echo "  test-all:         Run all tests"
      echo "  test-core:        Run core interface tests"
      echo "  test-consensus:   Run consensus tests"
      echo "  test-ibc:         Run IBC tests"
      echo "  test-integration: Run integration tests"
      echo "  test-stress:      Run stress tests"
      echo "  test-mockgen:     Generate mocks for testing"
      echo "  test-verbose:     Run tests with verbose output"
      echo "  test-race:        Run tests with race detection"
      echo "  test-cover:       Run tests with coverage"
      echo "  test-cover-html:  Run tests with coverage and generate HTML report"
      echo "  test-lint:        Run linter on tests"
      echo "  test-clean:       Clean test artifacts"
      echo ""
      echo "NOTE: You must build binaries first with 'nix run .#build-all' before running tests"
      echo ""
      echo "Example workflow:"
      echo "  1. nix run .#build-all      # Build all binaries"
      echo "  2. nix run .#test-verbose   # Run tests with verbose output"
    '');
  };
} 