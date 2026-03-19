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

**Phase 1-3 Complete**: Basic HTTP control daemon and WoL functionality (uses `wakeonlan` binary), WebSocket host daemon for heartbeats

## Dependencies

- `wakeonlan` - Wake-on-LAN packet sender (required for WoL functionality)
  - Install: `apt install wakeonlan` or `brew install wakeonlan`
- `go get github.com/gorilla/websocket` - WebSocket support

## Usage

### Start Control Daemon

```bash
./gbr control-daemon --config ~/.config/goober/goober-control.toml
```

The daemon listens on `127.0.0.1:29530` for HTTP requests. Requires `wakeonlan` to be installed.

### Wake a Host

```bash
./gbr wake chungus
```

This sends a Wake-on-LAN packet to the configured MAC address for the specified host. Requires `wakeonlan` to be installed.

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
```

### Client Config (`~/.config/goober/goober-client.toml`)

```toml
[server]
address = "127.0.0.1:29530"
```