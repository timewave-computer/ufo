# packages.nix - UFO package definitions
{ pkgs }:

let
  # Build the UFO binaries using buildGoModule for a more native approach
  ufo-binaries = pkgs.buildGoModule {
    pname = "ufo";
    version = "0.1.0";
    src = ../.;
    
    # Set to null to skip vendoring
    vendorHash = null;
    
    # Use Go proxy for dependencies
    proxyVendor = true;
    
    # Allow network access during build (for fetching dependencies)
    allowNetworkAccess = true;
    
    # Set environment variables
    env = {
      GOPROXY = "https://proxy.golang.org,direct";
    };
    
    # Fix the Go environment
    preBuild = ''
      export GOROOT="${pkgs.go_1_22}/share/go"
      export PATH="$GOROOT/bin:$PATH"
      export CGO_ENABLED=1
      export GOPROXY="https://proxy.golang.org,direct"
      
      # Make sure the modules are properly synced
      go mod download
      go mod tidy
    '';
    
    # Don't use subPackages, we'll build everything manually
    subPackages = [];
    
    # Build phase - build core binaries
    buildPhase = ''
      # Create bin directory
      mkdir -p bin
      
      echo "Building UFO core binaries..."
      
      # Ensure GOPROXY is set
      export GOPROXY="https://proxy.golang.org,direct"
      
      # Build the main binary
      go build -o bin/ufo ./main.go
      
      # Build the mock implementation binaries
      go build -o bin/fauxmosis-comet ./cmd/fauxmosis-comet
      go build -o bin/fauxmosis-ufo ./cmd/fauxmosis-ufo
      
      echo "UFO core binaries built successfully!"
    '';
    
    # Skip the default install phase and use our own
    dontInstall = true;
    
    # Custom install phase
    installPhase = ''
      # Create output directories
      mkdir -p $out/bin
      mkdir -p $out/tests/bin
      
      # Copy all binaries to output directories
      cp bin/* $out/bin/
      cp bin/* $out/tests/bin/
      
      echo "Installation complete. Binaries available in $out/bin and $out/tests/bin"
    '';
    
    # Disable tests during the build process
    doCheck = false;
    
    meta = with pkgs.lib; {
      description = "Universal Fast Orderer - A lightweight alternative to CometBFT for testing Cosmos applications";
      homepage = "https://github.com/timewave/ufo";
      license = licenses.asl20;
      platforms = platforms.unix;
    };
  };
  
  # Package that builds all the UFO binaries at once
  all-binaries = pkgs.buildGoModule {
    pname = "ufo-all-binaries";
    version = "0.1.0";
    src = ../.;
    
    # Set to null to skip vendoring
    vendorHash = null;
    
    # Use Go proxy for dependencies
    proxyVendor = true;
    
    # Allow network access during build
    allowNetworkAccess = true;
    
    # Set environment variables
    env = {
      GOPROXY = "https://proxy.golang.org,direct";
    };
    
    # Fix the Go environment and build all binaries
    preBuild = ''
      export GOROOT="${pkgs.go_1_22}/share/go"
      export PATH="$GOROOT/bin:$PATH"
      export CGO_ENABLED=1
      export GOPROXY="https://proxy.golang.org,direct"
      
      # Make sure the modules are properly synced
      go mod download
      go mod tidy
    '';
    
    # Don't use subPackages, we'll build everything manually
    subPackages = [];
    
    # Build phase - build all binaries
    buildPhase = ''
      # Create bin directory
      mkdir -p bin
      
      echo "Building all UFO binaries..."
      
      # Ensure GOPROXY is set
      export GOPROXY="https://proxy.golang.org,direct"
      
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
      
      echo "All binaries built successfully!"
    '';
    
    # Skip the default install phase and use our own
    dontInstall = true;
    
    # Custom install phase
    installPhase = ''
      # Create output directories
      mkdir -p $out/bin
      mkdir -p $out/tests/bin
      
      # Copy all binaries to output directories
      cp bin/* $out/bin/
      cp bin/* $out/tests/bin/
      
      echo "Installation complete. Binaries available in $out/bin and $out/tests/bin"
    '';
    
    # Skip check phase
    doCheck = false;
    
    meta = with pkgs.lib; {
      description = "All UFO binaries including integration approaches";
      homepage = "https://github.com/timewave/ufo";
      license = licenses.asl20;
      platforms = platforms.unix;
    };
  };

  # Test runner using buildGoModule
  ufo-test-runner = pkgs.buildGoModule {
    pname = "ufo-test-runner";
    version = "0.1.0";
    src = ../.;
    
    # Set to null to skip vendoring
    vendorHash = null;
    
    # Use Go proxy for dependencies
    proxyVendor = true;
    
    # Set environment variables
    env = {
      GOPROXY = "https://proxy.golang.org,direct";
    };
    
    # Fix the Go environment
    preBuild = ''
      export GOROOT="${pkgs.go_1_22}/share/go"
      export PATH="$GOROOT/bin:$PATH"
      export CGO_ENABLED=1
      export GOPROXY="https://proxy.golang.org,direct"
      
      # Make sure the modules are properly synced
      go mod download
      go mod tidy
    '';
    
    # Don't use subPackages, we'll build manually
    subPackages = [];
    
    # Build phase - build the test runner
    buildPhase = ''
      # Create bin directory
      mkdir -p bin
      
      echo "Building test runner..."
      
      # Ensure GOPROXY is set
      export GOPROXY="https://proxy.golang.org,direct"
      
      # Build the test runner
      go build -o bin/run_tests ./tests/run_tests.go
      
      echo "Test runner built successfully!"
    '';
    
    # Skip the default install phase and use our own
    dontInstall = true;
    
    # Custom install phase
    installPhase = ''
      # Create output directories
      mkdir -p $out/bin
      
      # Copy test runner
      cp bin/run_tests $out/bin/
      
      # Copy all binaries from ufo-binaries to ensure they're available
      cp ${ufo-binaries}/bin/* $out/bin/
      
      echo "Installation complete. Binaries and test runner available in $out/bin"
    '';
    
    # Disable tests during the build
    doCheck = false;
  };
in
{
  inherit ufo-binaries;
  inherit all-binaries;
  inherit ufo-test-runner;
} 