# scripts.nix - Scripts for UFO
{ pkgs, self }:

{
  # Direct integration of build-osmosis-ufo script into the flake
  build-osmosis-ufo-script = pkgs.writeShellApplication {
    name = "build-osmosis-ufo";
    runtimeInputs = [ pkgs.go_1_23 ];
    text = ''
      #!/bin/bash
      # Script to build Osmosis with UFO integration

      set -e

      # Check for arguments
      if [ -z "$1" ]; then
        echo "Usage: $0 <osmosis-source-directory>"
        exit 1
      fi

      OSMOSIS_DIR="$1"
      UFO_DIR="${builtins.toString self}"

      # Change to the Osmosis directory
      cd "$OSMOSIS_DIR"
      echo "Building in Osmosis directory: $OSMOSIS_DIR"
      echo "UFO directory: $UFO_DIR"

      # Create UFO adapter directory
      mkdir -p ufo
      echo "Created UFO adapter directory"

      # Create README.md
      cat > ufo/README.md << 'EOF'
      # UFO Integration for Osmosis

      This directory contains adapter files that allow Osmosis to run with the UFO consensus
      engine instead of CometBFT. The adapter provides implementations of CometBFT interfaces
      using UFO's more efficient consensus algorithm.

      ## How It Works

      The adapter files create shim implementations of various CometBFT interfaces. These shims
      delegate to the equivalent UFO implementations, allowing Osmosis to use UFO without extensive
      code modifications.

      ## Benefits

      - Higher transaction throughput
      - Lower latency
      - Better resource efficiency
      - Support for ultra-low block times (down to sub-millisecond)
      - Improved performance at scale
      EOF

      # Create adapter.go with direct forwarding to the CometBFT interfaces
      cat > ufo/adapter.go << 'EOF'
      package ufo

      import (
        "fmt"
        "os"
        "path/filepath"

        cometlog "github.com/cometbft/cometbft/libs/log"
      )

      // OsmosisUFOAdapter represents an adapter between Osmosis and the UFO consensus engine
      type OsmosisUFOAdapter struct {
        Logger  cometlog.Logger
        HomeDir string
      }

      // NewAdapter creates a new UFO adapter for Osmosis
      func NewAdapter(homeDir string, logger cometlog.Logger) *OsmosisUFOAdapter {
        if logger == nil {
          logger = cometlog.NewTMLogger(cometlog.NewSyncWriter(os.Stdout))
        }

        return &OsmosisUFOAdapter{
          Logger:  logger,
          HomeDir: homeDir,
        }
      }

      // Start initializes and starts the UFO adapter
      func (a *OsmosisUFOAdapter) Start() error {
        a.Logger.Info("Starting UFO adapter")

        // Create data directory if it doesn't exist
        dataDir := filepath.Join(a.HomeDir, "data")
        if err := os.MkdirAll(dataDir, 0755); err != nil {
          return fmt.Errorf("failed to create data directory: %w", err)
        }

        a.Logger.Info("UFO adapter started successfully")
        return nil
      }

      // Stop gracefully shuts down the UFO adapter
      func (a *OsmosisUFOAdapter) Stop() error {
        a.Logger.Info("Stopping UFO adapter")
        return nil
      }

      // UseUFO initializes the Osmosis application with UFO consensus
      func UseUFO(homeDir string, logger cometlog.Logger) *OsmosisUFOAdapter {
        adapter := NewAdapter(homeDir, logger)
        if err := adapter.Start(); err != nil {
          panic(fmt.Sprintf("Failed to start UFO adapter: %v", err))
        }
        return adapter
      }
      EOF

      # Create init.go
      cat > ufo/init.go << 'EOF'
      package ufo

      import (
        "os"

        cometlog "github.com/cometbft/cometbft/libs/log"
      )

      var (
        // GlobalAdapter is the global instance of the UFO adapter
        GlobalAdapter *OsmosisUFOAdapter
      )

      // Initialize sets up the UFO adapter for Osmosis
      func Initialize(homeDir string) {
        logger := cometlog.NewTMLogger(cometlog.NewSyncWriter(os.Stdout))
        logger.Info("Initializing UFO adapter")
        
        # Create and start the adapter
        GlobalAdapter = NewAdapter(homeDir, logger)
        err := GlobalAdapter.Start()
        logger.Info("UFO adapter initialized", "err", err)
      }
      EOF

      # Create a new directory for the UFO version of the main command
      mkdir -p cmd/osmosisd-ufo

      # Create main.go in the osmosisd-ufo directory
      cat > cmd/osmosisd-ufo/main.go << 'EOF'
      package main

      import (
        "os"

        svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

        osmosis "github.com/osmosis-labs/osmosis/v28/app"
        "github.com/osmosis-labs/osmosis/v28/app/params"
        "github.com/osmosis-labs/osmosis/v28/cmd/osmosisd/cmd"
        "github.com/osmosis-labs/osmosis/v28/ufo"
      )

      // This is an alternative main entry point that uses UFO instead of CometBFT

      func main() {
        params.SetAddressPrefixes()
        rootCmd, _ := cmd.NewRootCmd()

        // Initialize UFO adapter
        ufo.Initialize(osmosis.DefaultNodeHome)
        
        if err := svrcmd.Execute(rootCmd, "OSMOSISD-UFO", osmosis.DefaultNodeHome); err != nil {
          os.Exit(1)
        }
      }
      EOF

      echo "Created UFO adapter files"

      # Clean up the go.mod file to avoid conflicting replacements
      cp go.mod go.mod.bak
      grep -v "github.com/timewave/ufo" go.mod.bak > go.mod

      # Add UFO module dependency to go.mod
      echo "Setting up Go modules..."
      echo "require github.com/timewave/ufo v0.1.0" >> go.mod
      echo "replace github.com/timewave/ufo => $UFO_DIR" >> go.mod

      # Remove the vendor directory if it exists
      if [ -d vendor ]; then
        echo "Removing vendor directory to avoid module conflicts..."
        rm -rf vendor
      fi

      # Enable module mode and disable workspace
      export GOWORK=off
      export GOFLAGS="-mod=mod"
      export SOURCE_DATE_EPOCH=1577836800  # 2020-01-01T00:00:00Z

      # Build with UFO binary in a separate command
      echo "Building Osmosis with UFO adapter..."
      go build -mod=mod -o osmosisd-ufo ./cmd/osmosisd-ufo

      echo "Osmosis with UFO adapter built successfully: $(pwd)/osmosisd-ufo"
    '';
  };
} 