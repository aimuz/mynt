# Mynt - Walkthrough

Mynt is a simplified NAS system built with Go, following the RSC philosophy of simplicity.

## Components

- **myntd**: The main daemon that manages ZFS, metadata, and serves the Web UI.
- **mynt**: The CLI tool to interact with the daemon.

## Building

```bash
go build ./cmd/myntd
go build ./cmd/mynt
```

## Running

1. Start the daemon (requires root for ZFS operations):
   ```bash
   sudo ./myntd
   ```
   
2. Access the Web UI:
   Open [http://localhost:8080](http://localhost:8080) in your browser.

3. Use the CLI (Optional):
   ```bash
   ./mynt pool list
   ./mynt dataset list
   ```

## Architecture Notes

- **ZFS Integration**: Uses `os/exec` to call `zpool` and `zfs` binaries.
- **Metadata**: Stores system configuration in `mynt.db` (SQLite).
- **Web UI**: Embedded Single Page Application (Vanilla JS) served by `myntd`.
- **Background Tasks**: Periodic health checks run every minute.
- **Notifications**: Real-time updates via Server-Sent Events (SSE).
