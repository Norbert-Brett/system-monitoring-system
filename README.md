# System Monitor CLI (`sysmon`)

A high-performance, real-time system monitoring tool and developer companion written in Go. Designed for terminal enthusiasts and software engineers, `sysmon` provides live hardware monitoring alongside specialized developer utilities like port conflict resolution, build profiling, and environment diagnostic snapshots.

---

## Features

- **Live System Metrics**: Real-time CPU usage (overall and per-core), memory usage, disk usage, and network I/O.
- **Accurate macOS & Linux Support**:
  - **macOS (Darwin)**: Native Mach kernel statistics (`host_statistics64`, `HOST_CPU_LOAD_INFO`, `PROCESSOR_CPU_LOAD_INFO`) accounting for Active, Wired, and Compressed pages, with pure Go `/usr/bin/vm_stat` fallback.
  - **Linux**: Direct `/proc/stat`, `/proc/meminfo`, and `/proc/mounts` parsing.
- **Port & Listener Inspector (`sysmon ports`)**: Inspect all open listening ports mapped to PID, process name, tech stack, and working directory / project.
- **Quick-Kill Port Conflict Resolver (`sysmon kill-port <port>`)**: Eliminate `EADDRINUSE` conflicts in a single keystroke.
- **Build & Command Resource Profiler (`sysmon run <cmd>`)**: Track exact peak RSS memory consumption, user/system CPU time, and duration for builds and tests.
- **Diagnostic Snapshot (`sysmon snapshot`)**: Generate a clean Markdown or JSON diagnostic report of host specs, developer runtimes (Go, Node, Python, Docker, Git), and active ports.
- **Colorized Terminal & JSON Modes**: ANSI terminal dashboard with threshold alerts, plus machine-readable JSON streaming.
- **Configurable**: Threshold-based color warnings (Green / Yellow / Red) and customizable refresh intervals.

---

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/sysmon/system-monitor-cli.git
cd system-monitor-cli

# Build the binary
go build -o sysmon .

# Optionally, install to your PATH
sudo mv sysmon /usr/local/bin/
```

### Requirements
- Go 1.25 or later
- macOS (Apple Silicon or Intel) or Linux

---

## Quick Start & Command Reference

```
Available Commands:
  sysmon                    Launch real-time terminal monitor
  sysmon ports              Inspect listening ports and processes
  sysmon kill-port <port>   Quickly terminate process on a port
  sysmon run <command...>   Execute and profile resource usage of a command
  sysmon snapshot           Generate developer environment & system report
  sysmon version            Print version information
```

---

### 1. Real-Time Monitor (`sysmon`)

Launch the interactive terminal dashboard with default 1-second refresh interval:

```bash
./sysmon
```

#### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--interval` | Refresh interval (e.g. `1s`, `500ms`, `2m`) | `1s` |
| `--json` | Output metrics as streaming JSON objects | `false` |
| `--log-file` | Path to log file for metrics export | (none) |
| `--config` | Path to YAML or JSON config file | (none) |
| `--cpu-threshold` | CPU alert threshold percentage (0–100) | `80.0` |
| `--mem-threshold` | Memory alert threshold percentage (0–100) | `85.0` |
| `--disk-threshold` | Disk alert threshold percentage (0–100) | `90.0` |

#### Examples

```bash
# High-frequency 500ms monitoring
sysmon --interval 500ms

# Stream JSON metrics to stdout (useful for piping into jq or files)
sysmon --json

# Set custom alert thresholds
sysmon --cpu-threshold 75 --mem-threshold 80
```

---

### 2. Port Inspector (`sysmon ports`)

Inspect all listening TCP network ports, the owning PID, process name, detected tech stack, and the associated project directory or CWD:

```bash
sysmon ports
```

**Sample Output:**
```
  PORT     PID      PROCESS              STACK            CATEGORY        PROJECT / CWD
  ------------------------------------------------------------------------------------------------
  :3000    48210    node                 Node.js/JS       Dev Server      frontend-app
  :5432    1420     postgres             PostgreSQL       Database        /var/lib/postgresql
  :6379    1530     redis-server         Redis            Database        /
  :8080    51234    uvicorn              Python           Dev Server      backend-api
  :61268   32705    language_server      LSP              Language Server /Users/alice/projects/api

  Total: 5 listening ports. (Run 'sysmon kill-port <port>' to free a port)
```

#### Flags

| Flag | Description |
|------|-------------|
| `--dev-only` | Filter out system daemons to show only dev servers, databases, containers, and LSPs |
| `-k, --kill <port>` | Kill the process listening on the specified port |
| `-f, --force` | Force kill using SIGKILL (when using `--kill`) |
| `--json` | Output ports as structured JSON |

#### Examples

```bash
# View only development-related servers and services
sysmon ports --dev-only

# Terminate process on port 3000 directly from ports command
sysmon ports --kill 3000

# Output ports as JSON
sysmon ports --json | jq .
```

---

### 3. Kill Port (`sysmon kill-port <port>`)

Instantly terminate the process holding a network port to resolve `EADDRINUSE` conflicts without having to search for PIDs manually:

```bash
sysmon kill-port <port>
```

#### Flags

| Flag | Shorthand | Description |
|------|-----------|-------------|
| `--force` | `-f` | Force kill immediately with `SIGKILL` (default: graceful `SIGTERM` followed by verification) |

#### Examples

```bash
# Free port 3000 gracefully
sysmon kill-port 3000

# Forcefully terminate a stubborn server on port 8080
sysmon kill-port 8080 --force
```

---

### 4. Command & Build Profiler (`sysmon run <command...>`)

Run any shell command, test suite, or build tool while profiling its resource usage:

```bash
sysmon run <command> [args...]
```

Tracks and reports:
- **Peak RSS (Resident Set Size)** memory allocated by the command
- **Duration** (wall-clock elapsed time)
- **User CPU Time** vs. **System CPU Time**
- **Exit Status** code

#### Examples

```bash
# Profile a test suite
sysmon run go test -v ./...

# Profile a frontend build
sysmon run -- npm run build

# Profile a Rust compile
sysmon run cargo build --release

# Output profile report as JSON (ideal for CI/CD benchmarking)
sysmon run --json go test ./...
```

**Sample Output:**
```
🚀 Profiling command: go test -v ./internal/stats

=== RUN   TestDarwinMemoryStats
--- PASS: TestDarwinMemoryStats (0.00s)
PASS

────────────────────────────────────────────────────────────
📊 Execution & Resource Summary:
  Status:       Exit Code 0
  Duration:     130ms
  Peak RSS:     31358976 bytes (29.91 MB)
  User CPU:     106ms
  System CPU:   206ms
────────────────────────────────────────────────────────────
```

---

### 5. Environment & System Snapshot (`sysmon snapshot`)

Generate a comprehensive diagnostic report detailing host hardware specs, memory pressure, installed developer toolchains, and active listening ports.

```bash
sysmon snapshot
```

#### Flags

| Flag | Shorthand | Description | Default |
|------|-----------|-------------|---------|
| `--format` | `-f` | Report format: `markdown` or `json` | `markdown` |
| `--output` | `-o` | File path to write snapshot to | stdout |

#### Examples

```bash
# Print Markdown report to terminal
sysmon snapshot

# Save snapshot to markdown file for a GitHub issue or PR
sysmon snapshot --output env-report.md

# Dump as JSON for automated audits
sysmon snapshot --format json
```

---

## Configuration File

You can persist configuration options via `config.yaml` or `config.json`.

### YAML Example (`config.yaml`)

```yaml
interval: 2s
json: false
logFile: /var/log/sysmon.log
thresholds:
  cpu: 80.0
  memory: 85.0
  disk: 90.0
```

### JSON Example (`config.json`)

```json
{
  "interval": "2s",
  "json": false,
  "logFile": "/var/log/sysmon.log",
  "thresholds": {
    "cpu": 80.0,
    "memory": 85.0,
    "disk": 90.0
  }
}
```

Load configuration with `--config`:
```bash
sysmon --config config.yaml
```

---

## Architecture & Project Structure

```
.
├── cmd/
│   ├── root.go             # Root monitor command and flags
│   ├── ports.go            # 'ports' command (inspect & kill ports)
│   ├── kill_port.go        # 'kill-port' shortcut command
│   ├── run.go              # 'run' command profiler
│   ├── snapshot.go         # 'snapshot' diagnostic report command
│   └── version.go          # 'version' command
├── internal/
│   ├── collector/          # Periodic metrics collection engine
│   ├── config/             # Viper configuration & validation
│   ├── dev/                # Developer utilities
│   │   ├── detector.go     # Tech-stack and process classifier
│   │   ├── ports.go        # Port listener inspection & process killing
│   │   ├── profiler.go     # Command resource benchmark & rusage sampler
│   │   ├── snapshot.go     # Diagnostic report generator
│   │   └── types.go        # Core developer data models
│   ├── logger/             # JSON metrics file logging
│   ├── models/             # System metrics data models
│   ├── monitor/            # Application lifecycle orchestrator
│   ├── render/             # Terminal (ANSI) & JSON renderers
│   └── stats/              # OS-specific metrics providers
│       ├── darwin.go       # Darwin stats provider interface
│       ├── darwin_cgo.go   # Mach kernel host_statistics64 & CPU load (Cgo)
│       ├── darwin_nocgo.go # /usr/bin/vm_stat fallback (!Cgo)
│       ├── linux.go        # Linux /proc provider
│       └── mock.go         # Test mock provider
├── examples/               # Example configuration files
├── main.go                 # Application entrypoint
└── README.md
```

---

## Testing

Run all unit tests:

```bash
go test -v ./...
```

Run tests specifically for Darwin stats and developer modules:

```bash
# Test developer utilities
go test -v ./internal/dev

# Test Darwin OS statistics (Cgo enabled)
CGO_ENABLED=1 go test -v ./internal/stats

# Test Darwin OS statistics fallback (Cgo disabled)
CGO_ENABLED=0 go test -v ./internal/stats
```

---

## License

MIT License. See [LICENSE](LICENSE) for details.
