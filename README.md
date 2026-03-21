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

## Usage

### Start Control Daemon

```bash
./gbr control-daemon --config ~/.config/goober/goober-control.toml
```

The daemon listens on the address specified in the config file for HTTP requests.

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

### 1. Control Daemon (`gbr control-daemon`)
- Runs on desktop/server (typically always-on machine)
- Provides HTTP REST API on configured address/port
- Manages configured hosts and their MAC addresses
- Receives heartbeats from host daemons via WebSocket
- Tracks last_seen timestamps for each host to determine online status
- Provides `/api/hosts` endpoint with status information
- Implements Wake-on-LAN functionality via `/api/wake` endpoint
- Uses WebSocket server for real-time host communication

### 2. Host Daemon (`gbr host-daemon`)
- Runs on FreeBSD hosts that need to be managed
- Connects to control daemon via WebSocket (`ws://address:port/ws`)
- Sends periodic heartbeats with hostname to indicate availability
- Handles automatic reconnection with exponential backoff
- Supports future features: suspend detection, VM control
- Shortens hostname (removes domain part) for cleaner reporting

### 3. Client (`gbr` commands)
User-facing command-line interface that communicates with the control daemon via HTTP:
- `wake <hostname>` - Send Wake-on-LAN packet, then wait for host to come online
- `status` - Check current status of all configured hosts with thresholds
- `stop <hostname>` - Power off/suspend host (future implementation)
- `control-daemon` - Start the control daemon
- `host-daemon` - Start the host daemon
- `genkey` - Generate ed25519 key pairs (future authentication)
- `version` - Show build timestamp

### WebSocket Communication

The host daemon maintains a persistent WebSocket connection to the control daemon for real-time communication:

- **Connection**: Host daemon connects to `ws://<control-address>:<control-port>/ws`
- **Heartbeats**: Sent at configured interval (default: 60 seconds) with hostname
- **Reconnection**: Automatic reconnection with exponential backoff (max 2 minutes)
- **Message Format**: JSON payloads with `type`, `hostname`, and `timestamp` fields
- **Current Message Types**:
  - `heartbeat`: Sent from host to control daemon
  - `heartbeat:ack`: Sent from control to host (acknowledgment)
- **Future Message Types**: Planned for VM control, suspend detection, etc.

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

## Troubleshooting

### Common Issues

**Control daemon fails to start**
- Check if another process is already using the configured port
- Verify the TOML configuration file is valid and accessible
- Ensure you have permission to bind to the specified address/port

**Host daemon cannot connect to control daemon**
- Verify the control daemon is running and accessible
- Check network connectivity between host and control daemon
- Confirm WebSocket endpoint address and path are correct in host config
- Ensure no firewall is blocking the WebSocket connection

**Wake-on-LAN not working**
- Install `wakeonlan` command: `apt install wakeonlan` (Linux) or `brew install wakeonlan` (macOS)
- Verify the MAC address in the control daemon config is correct
- Ensure the target host is configured to accept WoL packets (often in BIOS)
- Check that the network allows UDP broadcast to port 9

**Host shows as offline despite being powered on**
- Verify host daemon is running on the target host
- Check WebSocket connection status in host daemon logs
- Confirm heartbeat interval is configured correctly
- Ensure the hostname in host daemon matches what's expected

### Logs
All components log to stdout/stderr. For persistent logging, redirect output:
```bash
./gbr control-daemon --config ~/.config/goober/goober-control.toml >> control.log 2>&1 &
```

### Debugging
Enable more verbose logging by modifying the logger initialization in the code (currently hardcoded to "info" level).