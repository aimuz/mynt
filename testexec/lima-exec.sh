#!/bin/bash
# lima-exec.sh - Execute test binary on Lima VM
#
# Usage:
#   go test -exec ./testexec/lima-exec.sh ./...
#   LIMA_INSTANCE=custom go test -exec ./testexec/lima-exec.sh ./...
#
# This script is designed for use with `go test -exec`.
# It copies the compiled test binary to the Lima VM and executes it there.

set -e

LIMA_INSTANCE="${LIMA_INSTANCE:-mynt}"
TEST_BINARY="$1"
shift

if [ -z "$TEST_BINARY" ]; then
    echo "Usage: $0 <test-binary> [args...]" >&2
    exit 1
fi

# Verify VM is running
if ! limactl list -q | grep -q "^${LIMA_INSTANCE}$"; then
    echo "Error: Lima instance '$LIMA_INSTANCE' not found." >&2
    echo "Create it with: limactl create --name=$LIMA_INSTANCE testexec/lima-mynt.yaml" >&2
    exit 1
fi

# Copy binary to VM
REMOTE_PATH="/tmp/$(basename "$TEST_BINARY")-$$"
limactl copy "$TEST_BINARY" "${LIMA_INSTANCE}:${REMOTE_PATH}"

# Cleanup on exit (including Ctrl+C)
trap 'limactl shell "$LIMA_INSTANCE" rm -f "$REMOTE_PATH"' EXIT

# Execute on VM
limactl shell "$LIMA_INSTANCE" chmod +x "$REMOTE_PATH"
limactl shell "$LIMA_INSTANCE" sudo "$REMOTE_PATH" "$@"
