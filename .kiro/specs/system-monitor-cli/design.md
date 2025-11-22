# Design Document

## Overview

The SystemMonitor is a command-line application written in Go that provides real-time system metrics monitoring similar to the Unix `top` command. The application follows a modular architecture with clear separation between metric collection, data presentation, and CLI interface layers. It uses goroutines for concurrent metric gathering, channels for communication, and contexts for lifecycle management.

The design emphasizes testability through interface-based abstractions, idiomatic Go patterns including dependency injection, and clean separation of concerns. The application supports both interactive terminal display with ANSI formatting and JSON output mode for programmatic consumption.

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         CLI Layer                            │
│                    (Cobra Commands)                          │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                    Application Core                          │
│                  (Monitor Orchestrator)                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Context    │  │   Config     │  │   Logger     │     │
│  │  Management  │  │   Manager    │  │              │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└────────────┬───────────────────────────────┬────────────────┘
             │                               │
             ▼                               ▼
┌────────────────────────────┐  ┌──────────────────────────┐
│   Metrics Collection       │  │   Presentation Layer     │
│                            │  │                          │
│  ┌──────────────────────┐ │  │  ┌────────────────────┐ │
│  │  MetricsCollector    │ │  │  │   TerminalUI       │ │
│  │  (Goroutine)         │ │  │  │   (ANSI Renderer)  │ │
│  └──────────────────────┘ │  │  └────────────────────┘ │
│            │               │  │           │             │
│            ▼               │  │           ▼             │
│  ┌──────────────────────┐ │  │  ┌────────────────────┐ │
│  │  System Interfaces   │ │  │  │   JSON Renderer    │ │
│  │  - CPU               │ │  │  │                    │ │
│  │  - Memory            │ │  │  └────────────────────┘ │
│  │  - Disk              │ │  │                          │
│  │  - Network           │ │  └──────────────────────────┘
│  └──────────────────────┘ │
│                            │
└────────────────────────────┘
```

### Component Interaction Flow

1. **Startup**: CLI layer parses flags and configuration, initializes the Monitor Orchestrator
2. **Collection Loop**: MetricsCollector goroutine polls system interfaces at RefreshInterval
3. **Data Flow**: Metrics are sent through channels to the presentation layer
4. **Rendering**: TerminalUI or JSON renderer formats and displays metrics
5. **Shutdown**: Signal handler triggers context cancellation, goroutines clean up gracefully

## Components and Interfaces

### 1. CLI Layer (cmd package)

**Purpose**: Handle command-line interface using Cobra library

**Components**:
- `rootCmd`: Main command that starts monitoring
- `versionCmd`: Subcommand to display version information
- Flag definitions: `--interval`, `--json`, `--log-file`, `--config`, threshold flags

**Responsibilities**:
- Parse command-line arguments
- Load configuration from file if specified
- Validate flag values
- Initialize and start the Monitor Orchestrator
- Handle version display

### 2. Monitor Orchestrator (monitor package)

**Purpose**: Coordinate the lifecycle of the monitoring application

**Interface**:
```go
type Monitor interface {
    Start(ctx context.Context) error
    Stop() error
}
```

**Implementation**:
```go
type SystemMonitor struct {
    config    *Config
    collector MetricsCollector
    renderer  Renderer
    logger    Logger
    ctx       context.Context
    cancel    context.CancelFunc
}
```

**Responsibilities**:
- Create and manage context for graceful shutdown
- Initialize collector and renderer components
- Set up signal handlers (SIGINT, SIGTERM)
- Coordinate the main monitoring loop
- Handle errors and logging

### 3. Metrics Collector (collector package)

**Purpose**: Gather system metrics from OS interfaces

**Interface**:
```go
type MetricsCollector interface {
    Collect(ctx context.Context) (*Metrics, error)
    Start(ctx context.Context, interval time.Duration, out chan<- *Metrics) error
}

type SystemStatsProvider interface {
    GetCPUStats() (*CPUStats, error)
    GetMemoryStats() (*MemoryStats, error)
    GetDiskStats() (*DiskStats, error)
    GetNetworkStats() (*NetworkStats, error)
}
```

**Implementation**:
```go
type Collector struct {
    provider SystemStatsProvider
    prevNet  *NetworkStats  // For calculating rates
    prevTime time.Time
}
```

**Responsibilities**:
- Poll system statistics at configured intervals
- Calculate derived metrics (percentages, rates)
- Handle per-component errors gracefully
- Send metrics through channels
- Maintain state for rate calculations (network I/O)

### 4. System Stats Provider (stats package)

**Purpose**: Abstract OS-specific system calls

**Interface**: See `SystemStatsProvider` above

**Implementations**:
- `LinuxStatsProvider`: Uses `/proc` filesystem
- `DarwinStatsProvider`: Uses `syscall` and `sysctl`
- `MockStatsProvider`: For testing

**Responsibilities**:
- Read CPU usage from OS
- Read memory statistics
- Read disk usage for mounted filesystems
- Read network interface statistics
- Handle OS-specific differences

### 5. Renderer (render package)

**Purpose**: Format and display metrics

**Interface**:
```go
type Renderer interface {
    Render(metrics *Metrics) error
    Clear() error
    Close() error
}
```

**Implementations**:

**TerminalRenderer**:
```go
type TerminalRenderer struct {
    writer    io.Writer
    colorizer Colorizer
    thresholds *Thresholds
}
```

**JSONRenderer**:
```go
type JSONRenderer struct {
    writer  io.Writer
    encoder *json.Encoder
}
```

**Responsibilities**:
- Format metrics for display
- Apply ANSI control codes (terminal mode)
- Apply color coding based on thresholds
- Align columns and format numbers
- Handle terminal size changes
- Serialize to JSON (JSON mode)

### 6. Configuration Manager (config package)

**Purpose**: Load and merge configuration from multiple sources

**Structure**:
```go
type Config struct {
    Interval    time.Duration
    JSONMode    bool
    LogFile     string
    ConfigFile  string
    Thresholds  Thresholds
}

type Thresholds struct {
    CPU     float64
    Memory  float64
    Disk    float64
}
```

**Responsibilities**:
- Load configuration from file (YAML/JSON)
- Merge with command-line flags (flags take precedence)
- Validate configuration values
- Provide defaults

### 7. Logger (logger package)

**Purpose**: Handle logging to file and error reporting

**Interface**:
```go
type Logger interface {
    LogMetrics(metrics *Metrics) error
    LogError(err error) error
    Close() error
}
```

**Responsibilities**:
- Write timestamped metrics to log file
- Handle file I/O errors
- Flush and close on shutdown

## Data Models

### Metrics Structure

```go
type Metrics struct {
    Timestamp time.Time
    CPU       CPUStats
    Memory    MemoryStats
    Disk      []DiskStats
    Network   []NetworkStats
}

type CPUStats struct {
    Overall float64   // Overall CPU usage percentage
    PerCore []float64 // Per-core usage percentages
}

type MemoryStats struct {
    Total     uint64  // Total memory in bytes
    Used      uint64  // Used memory in bytes
    Available uint64  // Available memory in bytes
    Percent   float64 // Usage percentage
}

type DiskStats struct {
    Mountpoint string
    Total      uint64  // Total space in bytes
    Used       uint64  // Used space in bytes
    Available  uint64  // Available space in bytes
    Percent    float64 // Usage percentage
}

type NetworkStats struct {
    Interface   string
    BytesSent   uint64  // Total bytes sent
    BytesRecv   uint64  // Total bytes received
    SendRate    float64 // Bytes per second
    RecvRate    float64 // Bytes per second
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*


### Property Reflection

After analyzing all acceptance criteria, several properties can be consolidated to eliminate redundancy:

- **Timing properties**: Multiple criteria (1.4, 2.4, 3.5, 4.4) test that different metrics update at RefreshInterval. These represent the same underlying property applied to different data types.
- **Error resilience**: Multiple criteria (1.3, 2.3, 3.4, 4.3) test that errors in one subsystem don't stop others. This is a single property about partial failure handling.
- **Threshold alerting**: Criteria 12.1, 12.2, 12.3 test the same alerting behavior for different metric types.
- **Percentage calculations**: Criteria 2.2 and 3.3 test the same calculation formula.
- **Collection completeness**: Multiple criteria test that all required fields are collected, which can be verified through structural properties.

### Correctness Properties

Property 1: Metric collection periodicity
*For any* valid RefreshInterval and running SystemMonitor, metrics should be collected and delivered at intervals matching the configured RefreshInterval within acceptable timing tolerance.
**Validates: Requirements 1.4, 2.4, 3.5, 4.4, 5.4**

Property 2: Partial failure resilience
*For any* metric collection cycle where one or more subsystems fail, the SystemMonitor should continue collecting metrics from functioning subsystems and log errors for failed subsystems without terminating.
**Validates: Requirements 1.3, 2.3, 3.4, 4.3**

Property 3: Percentage calculation correctness
*For any* metric with total and used values, the calculated percentage should equal (used / total) * 100, and should be in the range [0, 100].
**Validates: Requirements 2.2, 3.3**

Property 4: CPU metrics completeness
*For any* successful CPU collection, the returned CPUStats should contain an overall usage percentage in range [0, 100] and per-core percentages for all available cores, each in range [0, 100].
**Validates: Requirements 1.1, 1.2, 1.5**

Property 5: Memory metrics completeness
*For any* successful memory collection, the returned MemoryStats should contain total, used, and available values where total = used + available, and a percentage in range [0, 100].
**Validates: Requirements 2.1, 2.2**

Property 6: Disk metrics completeness
*For any* successful disk collection, the returned DiskStats should include all mounted filesystems, each with total, used, and available values where total = used + available, and a percentage in range [0, 100].
**Validates: Requirements 3.1, 3.2, 3.3**

Property 7: Network metrics completeness
*For any* successful network collection, the returned NetworkStats should include all network interfaces, each with bytes sent, bytes received, send rate, and receive rate values.
**Validates: Requirements 4.1, 4.2**

Property 8: Network rate calculation correctness
*For any* two consecutive network measurements with time delta T, the calculated send rate should equal (bytesSent2 - bytesSent1) / T, and receive rate should equal (bytesRecv2 - bytesRecv1) / T.
**Validates: Requirements 4.2**

Property 9: Interval configuration precedence
*For any* valid duration provided via --interval flag, the SystemMonitor should use that duration as RefreshInterval, overriding any default or config file value.
**Validates: Requirements 5.1, 14.2**

Property 10: Invalid interval rejection
*For any* invalid interval value (negative, zero, or unparseable), the SystemMonitor should display an error message and terminate with non-zero exit code.
**Validates: Requirements 5.3**

Property 11: JSON serialization completeness
*For any* Metrics object in JSONMode, the serialized JSON should contain all fields (timestamp, CPU, memory, disk, network) and be valid parseable JSON.
**Validates: Requirements 6.2**

Property 12: JSON output format purity
*For any* output line in JSONMode, the output should not contain ANSI escape sequences or terminal control codes.
**Validates: Requirements 6.4**

Property 13: Terminal ANSI code presence
*For any* output in terminal mode (non-JSON), the output should contain ANSI control codes for cursor positioning, clearing, and color formatting.
**Validates: Requirements 7.1, 7.2, 7.4**

Property 14: Context cancellation propagation
*For any* running SystemMonitor, when the context is cancelled, all goroutines should receive the cancellation signal and terminate within a reasonable timeout.
**Validates: Requirements 8.2, 8.3, 11.3**

Property 15: Channel closure safety
*For any* channel operation in the SystemMonitor, closing a channel should not cause panics in sender or receiver goroutines.
**Validates: Requirements 11.4**

Property 16: Threshold alerting consistency
*For any* metric type (CPU, memory, disk) with threshold alerting enabled, when the metric value exceeds the configured threshold, a warning indicator should be displayed in the output.
**Validates: Requirements 12.1, 12.2, 12.3**

Property 17: Threshold configuration validity
*For any* threshold value provided via command-line flags, the value should be in range [0, 100] or the SystemMonitor should reject it with an error.
**Validates: Requirements 12.4**

Property 18: Log file timestamp presence
*For any* metrics entry written to the log file, the entry should contain a timestamp in ISO 8601 format.
**Validates: Requirements 13.2**

Property 19: Log file error resilience
*For any* log file operation that fails, the SystemMonitor should log the error to stderr and continue monitoring without terminating.
**Validates: Requirements 13.3**

Property 20: Configuration file format support
*For any* valid configuration file in YAML or JSON format, the SystemMonitor should successfully parse and apply the settings.
**Validates: Requirements 14.4**

Property 21: Invalid configuration rejection
*For any* configuration file with invalid syntax, the SystemMonitor should display a descriptive error message and terminate with non-zero exit code.
**Validates: Requirements 14.3**

Property 22: Mock provider behavioral equivalence
*For any* MetricsCollector using a mock SystemStatsProvider, the collection behavior should be identical to using a real provider, differing only in the data source.
**Validates: Requirements 10.4**

Property 23: CLI error handling consistency
*For any* invalid command-line input (unknown flags, invalid commands), the SystemMonitor should display usage information and exit with non-zero status code.
**Validates: Requirements 9.3**

## Error Handling

### Error Categories

1. **System Access Errors**: Permission denied, unsupported OS features
   - Strategy: Log error, continue with available metrics
   - Example: Cannot read `/proc/stat` on Linux

2. **Configuration Errors**: Invalid flags, malformed config files
   - Strategy: Display error message, terminate immediately
   - Example: Invalid duration format for --interval

3. **I/O Errors**: Cannot write to log file, terminal errors
   - Strategy: Log to stderr, continue operation
   - Example: Disk full when writing log file

4. **Runtime Errors**: Channel closed, context cancelled
   - Strategy: Graceful cleanup and shutdown
   - Example: Context cancelled during metric collection

### Error Handling Patterns

```go
// Partial failure - continue with available data
func (c *Collector) Collect(ctx context.Context) (*Metrics, error) {
    metrics := &Metrics{Timestamp: time.Now()}
    
    if cpu, err := c.provider.GetCPUStats(); err != nil {
        log.Printf("CPU collection failed: %v", err)
    } else {
        metrics.CPU = *cpu
    }
    
    // Continue with other metrics...
    return metrics, nil
}

// Fatal error - terminate immediately
func loadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("cannot read config file: %w", err)
    }
    // Parse and return
}

// Graceful shutdown
func (m *SystemMonitor) Start(ctx context.Context) error {
    metricsChan := make(chan *Metrics)
    
    go m.collector.Start(ctx, m.config.Interval, metricsChan)
    
    for {
        select {
        case <-ctx.Done():
            return m.cleanup()
        case metrics := <-metricsChan:
            m.renderer.Render(metrics)
        }
    }
}
```

## Testing Strategy

### Dual Testing Approach

The SystemMonitor employs both unit testing and property-based testing to ensure comprehensive correctness:

- **Unit tests** verify specific examples, edge cases, and integration points
- **Property-based tests** verify universal properties across all valid inputs
- Together they provide complete coverage: unit tests catch concrete bugs, property tests verify general correctness

### Unit Testing

**Test Coverage**:
- Configuration parsing with various flag combinations
- Signal handler registration and invocation
- Terminal restoration on shutdown
- Specific error conditions (file not found, permission denied)
- Edge cases (empty metrics, zero values, single core CPU)

**Example Unit Tests**:
```go
func TestDefaultInterval(t *testing.T) {
    config := NewConfig()
    assert.Equal(t, 1*time.Second, config.Interval)
}

func TestInvalidIntervalReturnsError(t *testing.T) {
    _, err := ParseInterval("-1s")
    assert.Error(t, err)
}

func TestVersionCommand(t *testing.T) {
    output := executeCommand("version")
    assert.Contains(t, output, "v1.0.0")
}
```

### Property-Based Testing

**Framework**: Use `gopter` (Go property testing library) for property-based tests

**Configuration**: Each property-based test should run a minimum of 100 iterations

**Tagging Convention**: Each property-based test MUST be tagged with a comment explicitly referencing the correctness property from this design document using the format:
```go
// Feature: system-monitor-cli, Property 1: Metric collection periodicity
```

**Property Test Implementation**:

Each correctness property listed above MUST be implemented by a SINGLE property-based test. The test should:
1. Generate random valid inputs using gopter generators
2. Execute the system behavior
3. Assert the property holds

**Example Property Tests**:

```go
// Feature: system-monitor-cli, Property 3: Percentage calculation correctness
func TestPercentageCalculation(t *testing.T) {
    properties := gopter.NewProperties(nil)
    properties.Property("percentage is (used/total)*100 and in [0,100]", 
        prop.ForAll(
            func(total, used uint64) bool {
                if total == 0 || used > total {
                    return true // Skip invalid inputs
                }
                pct := calculatePercentage(used, total)
                expected := float64(used) / float64(total) * 100
                return math.Abs(pct-expected) < 0.01 && pct >= 0 && pct <= 100
            },
            gen.UInt64(),
            gen.UInt64(),
        ))
    properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: system-monitor-cli, Property 2: Partial failure resilience
func TestPartialFailureResilience(t *testing.T) {
    properties := gopter.NewProperties(nil)
    properties.Property("collector continues on partial failures",
        prop.ForAll(
            func(cpuFails, memFails, diskFails, netFails bool) bool {
                mock := &MockProvider{
                    CPUError:  cpuFails,
                    MemError:  memFails,
                    DiskError: diskFails,
                    NetError:  netFails,
                }
                collector := NewCollector(mock)
                metrics, err := collector.Collect(context.Background())
                
                // Should never return error, always return metrics
                if err != nil {
                    return false
                }
                
                // Verify non-failing subsystems have data
                if !cpuFails && metrics.CPU.Overall == 0 {
                    return false
                }
                return true
            },
            gen.Bool(),
            gen.Bool(),
            gen.Bool(),
            gen.Bool(),
        ))
    properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: system-monitor-cli, Property 16: Threshold alerting consistency
func TestThresholdAlerting(t *testing.T) {
    properties := gopter.NewProperties(nil)
    properties.Property("warning shown when threshold exceeded",
        prop.ForAll(
            func(value, threshold float64) bool {
                if threshold < 0 || threshold > 100 {
                    return true // Skip invalid thresholds
                }
                
                renderer := NewTerminalRenderer(&Thresholds{CPU: threshold})
                metrics := &Metrics{CPU: CPUStats{Overall: value}}
                output := renderer.Render(metrics)
                
                hasWarning := strings.Contains(output, "WARNING")
                shouldWarn := value > threshold
                
                return hasWarning == shouldWarn
            },
            gen.Float64Range(0, 100),
            gen.Float64Range(0, 100),
        ))
    properties.TestingRun(t, gopter.ConsoleReporter(false))
}
```

### Integration Testing

**Scope**: Test component interactions with real system calls on CI/CD platforms

**Tests**:
- End-to-end monitoring cycle with real OS data
- Signal handling with actual process signals
- File I/O with temporary files
- Terminal rendering with pseudo-terminals

### Mock Implementations

**MockStatsProvider**: Provides controllable test data
```go
type MockStatsProvider struct {
    CPUStats    *CPUStats
    MemStats    *MemoryStats
    DiskStats   *DiskStats
    NetStats    *NetworkStats
    CPUError    error
    MemError    error
    DiskError   error
    NetError    error
}
```

## Implementation Considerations

### OS-Specific Implementations

**Linux**:
- CPU: Parse `/proc/stat` for overall and per-core usage
- Memory: Parse `/proc/meminfo` for memory statistics
- Disk: Use `syscall.Statfs` for filesystem statistics
- Network: Parse `/proc/net/dev` for interface statistics

**macOS**:
- CPU: Use `syscall.Sysctl` with `kern.cp_time` and `kern.cp_times`
- Memory: Use `syscall.Sysctl` with `hw.memsize` and `vm.vmstat`
- Disk: Use `syscall.Statfs` for filesystem statistics
- Network: Use `syscall.Sysctl` with `net.interface` statistics

### Concurrency Patterns

**Collector Goroutine**:
```go
func (c *Collector) Start(ctx context.Context, interval time.Duration, out chan<- *Metrics) error {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    defer close(out)
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            metrics, _ := c.Collect(ctx)
            select {
            case out <- metrics:
            case <-ctx.Done():
                return ctx.Err()
            }
        }
    }
}
```

**Main Loop**:
```go
func (m *SystemMonitor) Start(ctx context.Context) error {
    metricsChan := make(chan *Metrics, 1)
    
    var wg sync.WaitGroup
    wg.Add(1)
    
    go func() {
        defer wg.Done()
        m.collector.Start(ctx, m.config.Interval, metricsChan)
    }()
    
    for {
        select {
        case <-ctx.Done():
            wg.Wait()
            return m.cleanup()
        case metrics, ok := <-metricsChan:
            if !ok {
                return nil
            }
            m.renderer.Render(metrics)
            if m.logger != nil {
                m.logger.LogMetrics(metrics)
            }
        }
    }
}
```

### Terminal Handling

**ANSI Control Codes**:
- Clear screen: `\033[2J`
- Move cursor to home: `\033[H`
- Clear line: `\033[2K`
- Color codes: `\033[31m` (red), `\033[32m` (green), `\033[33m` (yellow), `\033[0m` (reset)

**Terminal Detection**:
```go
func isTerminal(w io.Writer) bool {
    if f, ok := w.(*os.File); ok {
        return term.IsTerminal(int(f.Fd()))
    }
    return false
}
```

### Configuration File Format

**YAML Example**:
```yaml
interval: 2s
json: false
logFile: /var/log/sysmon.log
thresholds:
  cpu: 80.0
  memory: 85.0
  disk: 90.0
```

**JSON Example**:
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

## Dependencies

### External Libraries

1. **github.com/spf13/cobra** - CLI framework
   - Purpose: Command and flag parsing
   - License: Apache 2.0

2. **github.com/spf13/viper** - Configuration management
   - Purpose: Config file parsing and merging
   - License: MIT

3. **github.com/fatih/color** - Terminal colors
   - Purpose: ANSI color output
   - License: MIT

4. **github.com/leaanthony/go-ansi-parser** - ANSI parsing
   - Purpose: Terminal control codes
   - License: MIT

5. **gopkg.in/yaml.v3** - YAML parsing
   - Purpose: Configuration file support
   - License: Apache 2.0

6. **github.com/leanovate/gopter** - Property-based testing
   - Purpose: Property test framework
   - License: MIT

### Standard Library

- `context` - Cancellation and timeouts
- `os/signal` - Signal handling
- `time` - Timing and intervals
- `encoding/json` - JSON serialization
- `syscall` - OS-specific system calls
- `sync` - Synchronization primitives

## Performance Considerations

### Resource Usage

- **CPU**: Minimal overhead, collection should use <1% CPU
- **Memory**: Bounded memory usage, no unbounded growth
- **I/O**: Efficient parsing of `/proc` files, minimal disk I/O

### Optimization Strategies

1. **Reuse buffers** for parsing `/proc` files
2. **Cache file descriptors** for frequently read files
3. **Batch terminal updates** to reduce flicker
4. **Use buffered channels** to prevent blocking
5. **Limit string allocations** in hot paths

## Security Considerations

### Permissions

- Read access to `/proc` filesystem (Linux)
- Read access to system statistics (macOS)
- Write access to log file location
- No elevated privileges required for basic functionality

### Input Validation

- Validate all user inputs (flags, config files)
- Sanitize file paths to prevent directory traversal
- Limit interval values to reasonable ranges (1ms - 1h)
- Validate threshold values are in [0, 100]

### Error Information Disclosure

- Avoid exposing sensitive system paths in error messages
- Log detailed errors to file, show generic errors to user
- Don't include stack traces in user-facing output

## Future Enhancements

1. **Process monitoring**: Show top processes by CPU/memory
2. **Historical graphs**: Display sparklines for metric trends
3. **Remote monitoring**: Collect metrics from remote hosts
4. **Plugin system**: Allow custom metric collectors
5. **Web dashboard**: HTTP server for browser-based monitoring
6. **Alerting**: Send notifications via email/webhook
7. **Metric export**: Prometheus/StatsD integration
