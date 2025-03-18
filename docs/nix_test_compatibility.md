# IBC Testing in Nix Environment

## Overview

When using the nix environment for IBC testing, there are some important differences in how binaries behave compared to standard development environments. This document explains these differences and provides guidance on how to work with them.

## The Challenge

The nix-built binaries (e.g., `osmosis-ufo-patched`) automatically start a node when executed, rather than accepting commands like `init` or `start`. This behavior differs from the standard binaries used in the regular test setup, where we rely on command sequences like:

1. Initialize a node: `osmosis-ufo-patched init ...`
2. Start the node: `osmosis-ufo-patched start ...`

In the nix environment, the binary immediately starts a node when executed, which breaks the standard test workflow.

## Solutions

### Option 1: Use Nix-Compatible Tests

We've created a set of nix-compatible tests that demonstrate how to work with the auto-starting behavior:

```go
// Run these tests to see how nix binaries behave
go test -v ./tests/ibc/... -run TestNix
```

Available nix-compatible tests:
- `TestNixBasicIBCTransfer` - Basic test for starting chains
- `TestNixLightClientUpdates` - Light client functionality in nix
- `TestNixIBCSecurityBasics` - Security testing in nix
- `TestNixIBCTransfer` - IBC transfers in nix
- `TestNixIBCFork` - IBC fork testing in nix
- `TestNixIBCByzantine` - Byzantine validator testing in nix
- `TestNixIBCEnvironmentSetup` - Environment setup verification
- `TestNixHermesConfig` - Hermes relayer configuration testing
- `TestNixMultiChainSetup` - Multi-chain IBC setup testing

These tests:
- Detect the nix environment
- Start the binaries directly
- Show the auto-starting behavior
- Provide simple diagnostic information
- Use shared utilities for nix testing (from `nix_utils.go`)

### Option 2: Use the Nix Utilities in Your Tests

We've created reusable utilities for nix-compatible testing in `nix_utils.go`. Key components:

1. **Chain Configuration and Management**:
   ```go
   // Configure chains
   chain1Config := NixChainConfig{
       ChainID:    "chain-1",
       HomeDir:    chain1Dir,
       RPCPort:    26657,
       BinaryPath: binaryPath,
   }
   
   // Start chains
   chains, cleanup := StartNixChains(t, ctx, []NixChainConfig{chain1Config})
   defer cleanup()
   ```

2. **Test Setup and Environment**:
   ```go
   // Prepare directories
   chain1Dir, chain2Dir, relayerDir, cleanup := PrepareNixTestDirs(t, "TestName")
   defer cleanup()
   
   // Get binary path
   binaryPath := GetNixBinaryPath(t)
   ```

3. **Environment Detection**:
   ```go
   if !isInNixShell() {
       t.Skip("This test requires a nix environment")
   }
   
   // Configure environment
   configureEnvironment(t)
   ```

### Option 3: Run Tests Outside Nix

For full IBC testing without modifications, the most reliable approach is to run the tests outside the nix environment:

1. Exit your nix shell
2. Build the binaries manually
3. Run the tests with the manually built binaries

### Option 4: Modify the Nix Binary Behavior

If you need to run the full tests within nix, you would need to modify the nix flake to:

1. Build binaries that accept the standard command structure
2. Ensure the binaries don't auto-start when executed
3. Update the test infrastructure to use these modified binaries

## Adapting Existing Tests

To adapt existing tests to work in the nix environment:

1. **Replace direct command execution** with API/RPC calls:
   - Instead of `osmosis-ufo-patched init`, use RPC/API endpoints
   - Query status via HTTP instead of CLI commands

2. **Update chain initialization**:
   - Use `StartNixChains()` instead of running init and start commands
   - Pass configuration via environment variables

3. **Modify relayer setup**:
   - Use file-based Hermes configuration
   - Interact with Hermes via its JSON-RPC interface

4. **Update validation logic**:
   - Use HTTP/RPC queries instead of CLI commands
   - Implement proper error handling for RPC failures

## Testing Tips

When running tests in the nix environment:

1. **Use the -run flag to target specific tests**:
   ```
   nix develop .#ibc --command bash -c "go test -v ./tests/ibc/... -run TestNix"
   ```

2. **Check logs carefully**:
   - The logs will show details about port configurations
   - Watch for specific errors related to chain startup
   - Look for connectivity issues between chains

3. **Use longer timeouts**:
   - Nix environment may be slower to start
   - Configure with `defaultTestTimeout()`
   - Increase wait times after chain startup

4. **Debugging failing tests**:
   - Temporary directories are preserved on test failure
   - Check the logs for specific RPC or API errors
   - Verify binary permissions and paths

## Future Improvements

Planned improvements include:

1. Enhancing the nix flake to build test-friendly binaries
2. Providing wrappers or proxy scripts to adapt the auto-starting behavior
3. Developing a hybrid approach that works in both nix and non-nix environments
4. Adding more comprehensive HTTP/RPC interaction utilities

## Questions and Feedback

If you encounter issues or have suggestions for improving the nix IBC testing experience, please open an issue or discuss in the development channels. 