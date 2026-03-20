# Goober - FreeBSD Orchestration Tool

Dead-simple orchestration tool for managing FreeBSD hosts and vm-bhyve virtual machines.

## Build

### Build for all platforms
```bash
make all
```

Builds for 4 platforms into `dist/`:
- `dist/linux-amd64/gbr`
- `dist/linux-arm64/gbr`
- `dist/freebsd-amd64/gbr`
- `dist/freebsd-arm64/gbr`

### Build for specific platform
```bash
make build-linux-amd64    # Linux amd64
make build-linux-arm64    # Linux arm64
make build-freebsd-amd64  # FreeBSD amd64
make build-freebsd-arm64  # FreeBSD arm64
```

### Other commands
```bash
make clean    # Remove dist/ folder and local binary
make test     # Run tests
make deps     # Download dependencies
```

## Current Status

**Phase 1-4 Complete**: WebSocket-based orchestration with status tracking and thresholds, wake validation

## Dependencies

- `wakeonlan` - Wake-on-LAN packet sender (required for WoL functionality)
  - Install: `apt install wakeonlan` or `brew install wakeonlan`
- `go get github.com/gorilla/websocket` - WebSocket support

## Usage

### Start Control Daemon

```bash
./gbr control-daemon --config ~/.config/goober/goober-control.toml
```

The daemon listens on `127.0.0.1:29530` for HTTP requests.

### Wake a Host

```bash
./gbr wake chungus
```

This sends a Wake-on-LAN packet to the configured MAC address for the specified host, then waits up to 30 seconds for the host to connect via WebSocket before declaring it online.

### Check Host Status

```bash
./gbr status
```

Shows current status of all configured hosts:
- **online**: last heartbeat received within 2 minutes
- **unknown**: last heartbeat between 2-5 minutes ago
- **offline**: last heartbeat older than 5 minutes (or never connected)

Output example:
```
NAME    STATUS    LAST SEEN
---------------------------
chungus offline   2026-03-19 21:28:00
office-server online  2026-03-19 22:15:30
```

### Start Host Daemon

```bash
./gbr host-daemon --config ~/.config/goober/goober-host.toml
```

The host daemon connects to the control daemon via WebSocket and sends periodic heartbeats. Configure heartbeat interval in your config file.

## Configuration

### Control Daemon Config (`goober-control.toml`)

```toml
[server]
listen = "127.0.0.1:29530"

[host.home-server]
mac = "00:11:22:33:44:55"

[host.office-server]
mac = "aa:bb:cc:dd:ee:ff"
```

### Host Daemon Config (`goober-host.toml`)

```toml
[server]
listen = "127.0.0.1:29532"

[control]
address = "127.0.0.1:29530"
path = "/ws"

[heartbeat]
interval = 60

### Client Config (`~/.config/goober/goober-client.toml`)

```toml
[server]
address = "127.0.0.1:29530"
```

## Architecture

Goober consists of three main components:

1. **Control Daemon** (`gbr control-daemon`)
   - Runs on desktop/server
   - Listens on configured address for HTTP requests
   - Manages configured hosts and their MAC addresses
   - Receives heartbeats from host daemons via WebSocket
   - Tracks last_seen timestamps for each host
   - Provides `/api/hosts` endpoint with status information

2. **Host Daemon** (`gbr host-daemon`)
   - Runs on FreeBSD hosts
   - Connects to control daemon via WebSocket
   - Sends periodic heartbeats with hostname
   - Handles reconnection on connection loss
   - Supports future: suspend detection, VM control

3. **Client** (`gbr` commands)
   - `wake <hostname>` - Send WoL, wait for connection
   - `status` - Check host status with thresholds
   - `control-daemon` - Start control daemon
   - `host-daemon` - Start host daemon
   - `genkey` - Generate ed25519 keypairs (future)
   - `stop <hostname>` - Power off/suspend host (future)

### Status Thresholds

Host status is determined by last_seen timestamp:
- **2 minutes**: online
- **2-5 minutes**: unknown
- **5+ minutes**: offline

### Future Features

- **Suspend Detection**: Host daemon sends message before S3 suspend
- **Authentication**: ed25519 keypairs for host identification
- **Power Control**: `gbr stop` command for power-off/suspend
- **VM Lifecycle**: Start/stop VMs via `vm:bhyve` commands
- **Web UI**: HTTP API with real-time WebSocket updates
```

### Client Config (`~/.config/goober/goober-client.toml`)

```toml
[server]
address = "127.0.0.1:29530"
```