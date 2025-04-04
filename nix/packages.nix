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
      export GOROOT="${pkgs.go_1_23}/share/go"
      export PATH="$GOROOT/bin:$PATH"
      export CGO_ENABLED=1
      export GOPROXY="https://proxy.golang.org,direct"
      
      # Make sure the modules are properly synced
      go mod download
      go mod tidy
      
      # Platform-specific setup
      ${if pkgs.stdenv.isDarwin then ''
        # macOS specific setup
        echo "Building on Darwin (macOS)"
      '' else if pkgs.stdenv.isLinux then ''
        # Linux specific setup
        echo "Building on Linux"
        # Set specific flags for x86 Linux
        export CGO_CFLAGS="-I${pkgs.glibc.dev}/include"
        export CGO_LDFLAGS="-L${pkgs.glibc.out}/lib"
        # Add library paths to LD_LIBRARY_PATH
        export LD_LIBRARY_PATH="${pkgs.stdenv.cc.cc.lib}/lib:${pkgs.zlib}/lib:$LD_LIBRARY_PATH"
      '' else ''
        echo "Building on unsupported platform"
      ''}
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
      platforms = with platforms; linux ++ darwin;
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
      export GOROOT="${pkgs.go_1_23}/share/go"
      export PATH="$GOROOT/bin:$PATH"
      export CGO_ENABLED=1
      export GOPROXY="https://proxy.golang.org,direct"
      
      # Make sure the modules are properly synced
      go mod download
      go mod tidy
      
      # Platform-specific setup
      ${if pkgs.stdenv.isDarwin then ''
        # macOS specific setup
        echo "Building on Darwin (macOS)"
      '' else if pkgs.stdenv.isLinux then ''
        # Linux specific setup
        echo "Building on Linux"
        # Set specific flags for x86 Linux
        export CGO_CFLAGS="-I${pkgs.glibc.dev}/include"
        export CGO_LDFLAGS="-L${pkgs.glibc.out}/lib"
        # Add library paths to LD_LIBRARY_PATH
        export LD_LIBRARY_PATH="${pkgs.stdenv.cc.cc.lib}/lib:${pkgs.zlib}/lib:$LD_LIBRARY_PATH"
      '' else ''
        echo "Building on unsupported platform"
      ''}
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
      platforms = with platforms; linux ++ darwin;
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
      export GOROOT="${pkgs.go_1_23}/share/go"
      export PATH="$GOROOT/bin:$PATH"
      export CGO_ENABLED=1
      export GOPROXY="https://proxy.golang.org,direct"
      
      # Make sure the modules are properly synced
      go mod download
      go mod tidy
      
      # Platform-specific setup
      ${if pkgs.stdenv.isDarwin then ''
        # macOS specific setup
        echo "Building test runner on Darwin (macOS)"
      '' else if pkgs.stdenv.isLinux then ''
        # Linux specific setup
        echo "Building test runner on Linux"
        # Set specific flags for x86 Linux
        export CGO_CFLAGS="-I${pkgs.glibc.dev}/include"
        export CGO_LDFLAGS="-L${pkgs.glibc.out}/lib"
        # Add library paths to LD_LIBRARY_PATH
        export LD_LIBRARY_PATH="${pkgs.stdenv.cc.cc.lib}/lib:${pkgs.zlib}/lib:$LD_LIBRARY_PATH"
      '' else ''
        echo "Building test runner on unsupported platform"
      ''}
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
    '';
    
    # Skip the default install phase and use our own
    dontInstall = true;
    
    # Custom install phase
    installPhase = ''
      # Create output directories
      mkdir -p $out/bin
      
      # Copy test runner binary to output directory
      cp bin/run_tests $out/bin/
      
      echo "Installation complete. Test runner available at $out/bin/run_tests"
    '';
    
    # Disable tests during the build process
    doCheck = false;
    
    meta = with pkgs.lib; {
      description = "Test runner for UFO";
      homepage = "https://github.com/timewave/ufo";
      license = licenses.asl20;
      platforms = with platforms; linux ++ darwin;
    };
  };
in
{
  ufo-binaries = ufo-binaries;
  all-binaries = all-binaries;
  ufo-test-runner = ufo-test-runner;
} 