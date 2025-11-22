# System Monitor CLI

A real-time system monitoring tool written in Go that displays CPU usage, memory usage, disk usage, and network I/O statistics in your terminal, similar to the Unix `top` command.

## Features

- **Real-time Metrics**: Monitor CPU, memory, disk, and network statistics
- **Cross-Platform**: Supports Linux and macOS
- **Colorized Output**: Terminal display with ANSI colors and threshold-based warnings
- **JSON Mode**: Output metrics as JSON for integration with other tools
- **Configurable**: Set refresh intervals and alert thresholds
- **Logging**: Export metrics to a file for historical analysis
- **Graceful Shutdown**: Clean termination with Ctrl+C
- **Per-Core CPU**: View CPU usage for each individual core

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/system-monitor-cli.git
cd system-monitor-cli

# Build the binary
go build -o sysmon .

# Optionally, install to your PATH
sudo mv sysmon /usr/local/bin/
```

### Requirements

- Go 1.24 or later
- Linux or macOS operating system

## Usage

### Basic Usage

Start monitoring with default settings (1-second refresh interval):

```bash
./sysmon
```

### Command-Line Flags

```bash
# Set custom refresh interval
./sysmon --interval 2s

# Output as JSON
./sysmon --json

# Log metrics to a file
./sysmon --log-file /var/log/sysmon.log

# Set custom alert thresholds
./sysmon --cpu-threshold 90 --mem-threshold 80 --disk-threshold 95

# Use a configuration file
./sysmon --config config.yaml

# Combine multiple options
./sysmon --interval 500ms --cpu-threshold 75 --log-file metrics.log
```

### Available Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--interval` | Refresh interval (e.g., 1s, 500ms, 2m) | 1s |
| `--json` | Output metrics as JSON | false |
| `--log-file` | Path to log file for metrics export | (none) |
| `--config` | Path to configuration file (YAML or JSON) | (none) |
| `--cpu-threshold` | CPU usage alert threshold (0-100) | 80 |
| `--mem-threshold` | Memory usage alert threshold (0-100) | 85 |
| `--disk-threshold` | Disk usage alert threshold (0-100) | 90 |

### Commands

```bash
# Display version information
./sysmon version

# Show help
./sysmon --help
```

## Configuration File

You can use a configuration file to persist your preferences. Both YAML and JSON formats are supported.

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

### Using Configuration Files

```bash
# Load configuration from file
./sysmon --config config.yaml

# Command-line flags override config file settings
./sysmon --config config.yaml --interval 1s
```

Example configuration files are provided in the `examples/` directory.

## Output Modes

### Terminal Mode (Default)

Displays colorized, formatted output that updates in place:

```
System Monitor - 2024-01-15 14:30:45

CPU Usage:
  Overall:  45.23%
  Per Core:
    Core  0:  42.10%
    Core  1:  48.50%
    Core  2:  44.20%
    Core  3:  46.10%

Memory Usage:
  Usage:     62.50%
  Total:     16.00 GB
  Used:      10.00 GB
  Available:  6.00 GB

Disk Usage:
  /
    Usage:     75.20%
    Total:    500.00 GB
    Used:     376.00 GB
    Available: 124.00 GB

Network I/O:
  eth0
    Sent:     1.50 GB (125.00 KB/s)
    Received: 3.20 GB (250.00 KB/s)
```

### JSON Mode

Outputs one JSON object per refresh interval:

```bash
./sysmon --json
```

```json
{
  "timestamp": "2024-01-15T14:30:45Z",
  "cpu": {
    "overall": 45.23,
    "perCore": [42.1, 48.5, 44.2, 46.1]
  },
  "memory": {
    "total": 17179869184,
    "used": 10737418240,
    "available": 6442450944,
    "percent": 62.5
  },
  "disk": [
    {
      "mountpoint": "/",
      "total": 536870912000,
      "used": 403726073856,
      "available": 133144838144,
      "percent": 75.2
    }
  ],
  "network": [
    {
      "interface": "eth0",
      "bytesSent": 1610612736,
      "bytesRecv": 3435973836,
      "sendRate": 128000,
      "recvRate": 256000
    }
  ]
}
```

## Alert Thresholds

When metrics exceed configured thresholds, warnings are displayed:

- **Green**: Normal (below 80% of threshold)
- **Yellow**: Warning (80-100% of threshold)
- **Red**: Critical (above threshold) with ⚠ WARNING indicator

## Logging

Enable logging to export metrics to a file:

```bash
./sysmon --log-file /var/log/sysmon.log
```

Log entries are written in JSON format with ISO 8601 timestamps:

```json
{
  "timestamp": "2024-01-15T14:30:45Z",
  "metrics": { ... }
}
```

## Graceful Shutdown

Press `Ctrl+C` to stop monitoring. The application will:
- Cancel all running goroutines
- Flush and close log files
- Restore terminal to original state
- Exit with status code 0

## Architecture

The application follows a modular design:

- **CLI Layer**: Cobra-based command-line interface
- **Monitor Orchestrator**: Coordinates lifecycle and components
- **Metrics Collector**: Gathers system statistics using goroutines
- **Stats Providers**: OS-specific implementations (Linux/macOS)
- **Renderers**: Terminal (ANSI) and JSON output formatters
- **Logger**: File-based metrics export

## Platform Support

### Linux

- CPU: Reads from `/proc/stat`
- Memory: Reads from `/proc/meminfo`
- Disk: Uses `syscall.Statfs`
- Network: Reads from `/proc/net/dev`

### macOS

- CPU: Uses `syscall.Sysctl` with `kern.cp_time`
- Memory: Uses `syscall.Sysctl` with `hw.memsize`
- Disk: Uses `syscall.Statfs`
- Network: Limited support (requires IOKit integration)

## Troubleshooting

### Permission Denied Errors

Some system files may require elevated permissions:

```bash
# Run with sudo if needed
sudo ./sysmon
```

### High CPU Usage

If the monitor itself uses too much CPU, increase the refresh interval:

```bash
./sysmon --interval 5s
```

### No Network Statistics (macOS)

Network statistics on macOS require additional system integration. This is a known limitation.

### Configuration File Not Found

Ensure the path to your configuration file is correct:

```bash
./sysmon --config /path/to/config.yaml
```

## Development

### Building

```bash
go build -o sysmon .
```

### Running Tests

```bash
go test ./...
```

### Project Structure

```
.
├── cmd/                    # CLI commands
│   ├── root.go            # Root command
│   └── version.go         # Version command
├── internal/
│   ├── collector/         # Metrics collection
│   ├── config/            # Configuration management
│   ├── logger/            # File logging
│   ├── models/            # Data structures
│   ├── monitor/           # Orchestrator
│   ├── render/            # Output renderers
│   └── stats/             # OS-specific providers
├── examples/              # Example configurations
├── main.go               # Application entry point
└── README.md             # This file
```

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## Acknowledgments

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [Color](https://github.com/fatih/color) - Terminal colors
