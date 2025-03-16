# UFO Tests

This directory contains the test suite for the UFO (Universal Fast Orderer) project.

## Running Tests with Nix

All test functionality is now integrated with our Nix flake system. This replaces the previous Makefile-based approach with a more consistent and reproducible workflow.

### Available Test Commands

Run these commands from the project root using `nix run`:

```bash
# List all available test commands
nix run .#test-help

# Run all tests
nix run .#test-all

# Run specific test categories
nix run .#test-core
nix run .#test-consensus
nix run .#test-ibc
nix run .#test-integration
nix run .#test-stress

# Run tests with special options
nix run .#test-verbose    # Verbose output
nix run .#test-race       # Race detection
nix run .#test-cover      # Coverage report
nix run .#test-cover-html # HTML coverage report

# Other commands
nix run .#test-mockgen    # Generate mocks
nix run .#test-lint       # Run linter
nix run .#test-clean      # Clean test artifacts
```

### Building Test Binaries

Before running tests, you may need to build all the required binaries:

```bash
# Build all UFO binaries including test binaries
nix run .#build-all
```

This will build and place all binaries in the `tests/bin` directory.

## Test Directory Structure

- `core/` - Core functionality tests
- `consensus/` - Consensus algorithm tests
- `ibc/` - Inter-Blockchain Communication tests
- `integration/` - Integration tests
- `stress/` - Stress and performance tests
- `utils/` - Testing utilities
- `bin/` - Test binaries (generated during build) 