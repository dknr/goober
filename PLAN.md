# Goober Orchestration Tool - Plan

## Overview

Goober is a dead-simple orchestration tool for managing FreeBSD hosts and vm-bhyve virtual machines. Distributed as a **single binary** with subcommands for all functionality.

It enables:
- Wake-on-LAN to start hosts on-demand
- Wait for hosts to come online before proceeding
- Start virtual machines with vm-bhyve on target hosts

## Distribution

Single binary: `gbr`

### Available Subcommands

**Daemons:**
- `gbr control-daemon` - Run the control daemon
- `gbr host-daemon` - Run the host daemon
- `gbr node-daemon` - Run the node daemon (VM)

**Utilities:**
- `gbr genkey` - Generate ed25519 key pairs

**Client:**
- `gbr wake <name>` - Wake a host via WoL
- `gbr ls` - List hosts and nodes
- `gbr stop <name>` - Stop a VM or host
- `gbr status` - Show status of VMs and hosts

## Architecture

### Components (all in one binary)

1. **Control Daemon** (`gbr control-daemon`)
   - Runs on always-up host (user's desktop)
   - HTTP REST API for control
   - Manages state across all hosts/VMs
   - Persistent storage in SQLite

2. **Host Daemon** (`gbr host-daemon`)
   - Runs on physical FreeBSD hosts
   - Receives commands from control daemon
   - Manages vm-bhyve lifecycle
   - Sends heartbeat/status updates

3. **Node Daemon** (`gbr node-daemon`)
   - Runs inside virtual machines
   - Sends heartbeat/status updates
   - Can receive control commands

4. **Client CLI** (subcommands: `wake`, `ls`, `stop`, `status`)
   - User-facing commands for control daemon
   - Reads config from `~/.config/goober/goober.toml`
   - HTTP client for control daemon

### Communication Flow

```
User (gbr CLI subcommands) → gbr control-daemon (HTTP)
                                 ↓
                          gbr host-daemon (remote host)
                                 ↓
                          gbr node-daemon (inside VM)
```

## Configuration Files

### Control Daemon Config (`gbr-control.toml`)
```toml
[server]
listen = "127.0.0.1:29530"

[[host.home-server]]
mac = "00:11:22:33:44:55"
```

### Client Config (`~/.config/goober/goober.toml`)
```toml
[server]
address = "127.0.0.1:29530"
```

## Implementation Phases

### Phase 1: Foundation
- [ ] Initialize Go module (`go mod init github.com/lore/goober`)
- [ ] Set up directory structure
- [ ] Create ed25519 key generation command (`gbr genkey`)
- [ ] Create sample TOML config files
- [ ] Logging utility

### Phase 2: Core Infrastructure
- [ ] Config parsing for control daemon
- [ ] HTTP server setup
- [ ] WoL packet sending
- [ ] Shared data structures (models)

### Phase 3: Control Daemon (`gbr control-daemon`)
- [ ] HTTP server listening on port 29530
- [ ] Host configuration parsing (table names: `[host.home-server]`)
- [ ] `POST /api/wake` endpoint
- [ ] WoL packet sending (UDP to MAC:9)
- [ ] Simple JSON responses

### Phase 4: Client CLI (`gbr wake`)
- [ ] CLI command: `gbr wake <hostname>`
- [ ] Config file: `~/.config/goober/goober.toml`
- [ ] HTTP POST to control daemon
- [ ] Display response

### Phase 5: Additional Client Commands
- [ ] `gbr ls` - List hosts/nodes
- [ ] `gbr stop <name>` - Stop VM/host
- [ ] `gbr status` - Show status

### Phase 6: Host Daemon (`gbr host-daemon`)
- [ ] HTTP server
- [ ] VM lifecycle management (vm-bhyve)
- [ ] Heartbeat/status reporting

### Phase 7: Node Daemon (`gbr node-daemon`)
- [ ] Minimal server for status updates
- [ ] Heartbeat reporting

### Phase 8: Client Utility (`gbr genkey`)
- [ ] ed25519 key generation implementation
- [ ] Key file writing
- [ ] Permissions management

## Key Design Decisions

### Configuration

- **Control daemon config** (TOML): `gbr-control.toml`
  - Listen address/port
  - SQLite database path (future)
  - Hosts as table names: `[host.name]`
  - WoL MAC addresses (no port config - always 9)

- **Client config** (TOML): `~/.config/goober/goober.toml`
  - Control daemon server address
  - User's ed25519 key pair (future)

### Host Configuration Format

```toml
[[host.home-server]]
mac = "00:11:22:33:44:55"

[[host.office-server]]
mac = "aa:bb:cc:dd:ee:ff"
```

**Benefits:**
- Table names act as host IDs (no need for separate `name` field)
- Natural grouping
- Easy to expand (future: IP, description, etc.)

### WoL Implementation

- UDP broadcast to MAC:9
- Default port 9 (no configuration needed)
- Simple magic packet format

### Transport Protocol

- HTTP REST API
- JSON request/response
- Enables future web client
- Simple and standard

### Data Storage

SQLite database for control daemon (`gbr.db`):
- `hosts` table - registered hosts
- `vms` table - registered VMs (future)
- `tasks` table - queued/control tasks (future)
- `keys` table - ed25519 public keys (future)

## Directory Structure

```
src/goober/
├── main.go              # CLI entry point - Cobra root command
├── cmd/
│   ├── control-daemon/  # gbr control-daemon subcommand
│   │   └── main.go
│   ├── host-daemon/     # gbr host-daemon subcommand
│   │   └── main.go
│   ├── node-daemon/     # gbr node-daemon subcommand
│   │   └── main.go
│   ├── genkey/          # gbr genkey subcommand
│   │   └── main.go
│   └── config/          # gbr config subcommand
│       └── main.go
├── internal/
│   ├── config/          # TOML config parsing
│   ├── crypto/          # ed25519 key operations
│   ├── database/        # SQLite helpers
│   ├── http/            # HTTP server and handlers
│   ├── models/          # Data structures
│   └── logging/         # Logging utilities
├── pkg/
│   ├── host/            # Host management logic
│   ├── vm/              # VM management logic
│   └── wol/             # Wake-on-LAN utilities
├── scripts/
│   ├── generate-config.sh # Generate sample configs
│   └── install.sh       # Install binary
├── examples/            # Sample configuration files
│   ├── gbr-control.toml
│   └── gbr.toml
├── go.mod               # Go module dependencies
├── go.sum               # Dependency checksums
└── README.md            # Documentation
```

## Dependencies

- `github.com/google/uuid` - Unique identifiers
- `github.com/spf13/cobra` - CLI framework (for gbr client)
- `github.com/tcnksm/ghotlines` - TOML config parsing
- `github.com/mattn/go-sqlite3` - SQLite driver (future)
- `github.com/gorilla/mux` - HTTP router (future)
- `github.com/go-chi/cors` - CORS middleware (future)

## Success Criteria

1. Can wake a host via WoL from command line
2. Control daemon listens on configured HTTP port
3. Hosts configured in table names: `[host.name]`
4. WoL port always 9 (no configuration needed)
5. Simple JSON responses
6. All in single `gbr` binary

## Open Questions

1. **Authentication**: Skip for now, add later if needed?
2. **Validation**: MAC address format, host existence?
3. **Error messages**: What format for errors?
4. **Web client**: When to start designing?
5. **VM management**: After host WoL is working?

## Notes

- Start with HTTP and basic WoL only
- Add SQLite, authentication, and VM support later
- Keep it "dead-simple" - one feature at a time
- Focus on `gbr wake` first, then expand