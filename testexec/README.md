# testexec

Remote test execution using Lima VMs.

## Quick Start

```bash
# Install Lima (macOS)
brew install lima

# Create the test VM
cd testexec
make vm-create

# Run integration tests on VM
make test-integration
```

## Commands

| Command | Description |
|---------|-------------|
| `make vm-create` | Create and start VM with Debian + ZFS |
| `make vm-start` | Start existing VM |
| `make vm-stop` | Stop VM |
| `make vm-delete` | Delete VM |
| `make vm-shell` | Open shell in VM |
| `make test-integration` | Run integration tests on VM |

## Manual Usage

```bash
# Cross-compile for Linux and run on VM
GOOS=linux go test -tags=integration -exec ./testexec/lima-exec.sh ./zfs/...

# Use custom VM instance
LIMA_INSTANCE=myvm go test -exec ./testexec/lima-exec.sh ./...
```

## How It Works

1. `go test` compiles test binary for Linux (`GOOS=linux`)
2. `lima-exec.sh` copies binary to VM via `limactl copy`
3. Executes binary on VM via `limactl shell`
4. Returns exit code to `go test`

## VM Configuration

The VM uses Debian 12 (Bookworm) with:
- 2 CPUs, 4GB RAM, 20GB disk
- ZFS (installed via provision script)
- Home directory mounted read-only
