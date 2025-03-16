#!/bin/bash
# Script to clean up benchmark results directory, keeping only consolidated runs

set -e

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BENCHMARK_DIR="${PROJECT_DIR}/benchmark_results"

# Function to print usage
function print_usage {
  echo "Usage: $0 [OPTIONS]"
  echo "Options:"
  echo "  --keep-run NAME       Name of the benchmark run to keep (required)"
  echo "  --backup-dir DIR      Directory for backup (default: ${BENCHMARK_DIR}_backup)"
  echo "  --no-backup           Don't create a backup before cleanup"
  echo "  --help                Display this help message"
}

# Default values
KEEP_RUN=""
BACKUP_DIR="${BENCHMARK_DIR}_backup"
NO_BACKUP=false

# Parse arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    --keep-run)
      KEEP_RUN="$2"
      shift 2
      ;;
    --backup-dir)
      BACKUP_DIR="$2"
      shift 2
      ;;
    --no-backup)
      NO_BACKUP=true
      shift
      ;;
    --help)
      print_usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      print_usage
      exit 1
      ;;
  esac
done

# Check if keep-run is specified
if [ -z "$KEEP_RUN" ]; then
  echo "Error: You must specify which run to keep with --keep-run"
  print_usage
  exit 1
fi

# Check if the specified run exists
if [ ! -d "${BENCHMARK_DIR}/${KEEP_RUN}" ]; then
  echo "Error: Run directory '${BENCHMARK_DIR}/${KEEP_RUN}' does not exist."
  echo "Available runs:"
  ls -d ${BENCHMARK_DIR}/benchmark_run_* 2>/dev/null || echo "  No benchmark runs found."
  exit 1
fi

# Backup the benchmark directory if requested
if [ "$NO_BACKUP" = false ]; then
  echo "Creating backup of benchmark_results directory to $BACKUP_DIR..."
  rm -rf "$BACKUP_DIR"
  mkdir -p $(dirname "$BACKUP_DIR")
  cp -R "$BENCHMARK_DIR" "$BACKUP_DIR"
  echo "Backup created successfully."
fi

echo "Cleaning up benchmark_results directory..."
echo "Keeping only the run: $KEEP_RUN"

# Create a temporary directory to store the run we want to keep
TEMP_DIR=$(mktemp -d)
cp -R "${BENCHMARK_DIR}/${KEEP_RUN}" "$TEMP_DIR"

# Remove all files in the benchmark directory except .gitignore
find "$BENCHMARK_DIR" -mindepth 1 -not -name ".gitignore" -exec rm -rf {} \; 2>/dev/null || true

# Move the kept run back to the benchmark directory
mkdir -p "$BENCHMARK_DIR"
cp -R "${TEMP_DIR}/$(basename ${KEEP_RUN})"/* "$BENCHMARK_DIR"
rm -rf "$TEMP_DIR"

echo "Clean up completed. Only the consolidated results from run '$KEEP_RUN' remain."
echo "All benchmark results are now directly in the benchmark_results directory."
echo "Results can be found at: $BENCHMARK_DIR"
echo "Jupyter notebook: $BENCHMARK_DIR/benchmark_analysis.ipynb"
echo
echo "To open the notebook using Nix:"
echo "nix run .#notebook \"$BENCHMARK_DIR/benchmark_analysis.ipynb\""
echo
echo "Or for a more interactive experience:"
echo "nix develop .#jupyter"
echo
echo "To use the notebook in Cursor IDE:"
echo "1. Open the notebook in Cursor IDE"
echo "2. Ensure you're using a Nix environment with Jupyter"
echo "3. Select the 'UFO Benchmark' kernel from the kernel selector" 