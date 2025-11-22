# Implementation Plan

- [x] 1. Set up project structure and core data models
  - Initialize Go module and directory structure (cmd, internal/collector, internal/render, internal/config, internal/stats)
  - Define core data structures: Metrics, CPUStats, MemoryStats, DiskStats, NetworkStats
  - Define interfaces: SystemStatsProvider, MetricsCollector, Renderer, Logger
  - _Requirements: 15.1, 15.2, 15.3, 15.4_

- [ ]* 1.1 Write property test for percentage calculation
  - **Property 3: Percentage calculation correctness**
  - **Validates: Requirements 2.2, 3.3**

- [x] 2. Implement system stats providers for Linux and macOS
  - Create LinuxStatsProvider that reads from /proc filesystem
  - Implement CPU stats collection from /proc/stat (overall and per-core)
  - Implement memory stats collection from /proc/meminfo
  - Implement disk stats collection using syscall.Statfs
  - Implement network stats collection from /proc/net/dev
  - Create DarwinStatsProvider using syscall.Sysctl
  - Implement CPU, memory, disk, and network collection for macOS
  - Add OS detection and provider selection logic
  - _Requirements: 1.1, 1.2, 2.1, 3.1, 3.2, 4.1_

- [ ]* 2.1 Write property test for CPU metrics completeness
  - **Property 4: CPU metrics completeness**
  - **Validates: Requirements 1.1, 1.2, 1.5**

- [ ]* 2.2 Write property test for memory metrics completeness
  - **Property 5: Memory metrics completeness**
  - **Validates: Requirements 2.1, 2.2**

- [ ]* 2.3 Write property test for disk metrics completeness
  - **Property 6: Disk metrics completeness**
  - **Validates: Requirements 3.1, 3.2, 3.3**

- [ ]* 2.4 Write property test for network metrics completeness
  - **Property 7: Network metrics completeness**
  - **Validates: Requirements 4.1, 4.2**

- [x] 3. Implement metrics collector with concurrency
  - Create Collector struct that wraps SystemStatsProvider
  - Implement Collect() method that gathers all metrics and calculates percentages
  - Implement partial failure handling (continue on individual subsystem errors)
  - Implement Start() method with goroutine that polls at RefreshInterval
  - Add network rate calculation by tracking previous measurements
  - Use channels to send collected metrics
  - Handle context cancellation for graceful shutdown
  - _Requirements: 1.3, 1.4, 2.2, 2.3, 2.4, 3.3, 3.4, 3.5, 4.2, 4.3, 4.4, 11.1, 11.2, 11.3_

- [ ]* 3.1 Write property test for partial failure resilience
  - **Property 2: Partial failure resilience**
  - **Validates: Requirements 1.3, 2.3, 3.4, 4.3**

- [ ]* 3.2 Write property test for network rate calculation
  - **Property 8: Network rate calculation correctness**
  - **Validates: Requirements 4.2**

- [ ]* 3.3 Write property test for metric collection periodicity
  - **Property 1: Metric collection periodicity**
  - **Validates: Requirements 1.4, 2.4, 3.5, 4.4, 5.4**

- [ ]* 3.4 Write property test for channel closure safety
  - **Property 15: Channel closure safety**
  - **Validates: Requirements 11.4**

- [x] 4. Implement terminal renderer with ANSI formatting
  - Create TerminalRenderer struct with writer and colorizer
  - Implement Render() method that formats metrics as aligned table
  - Add ANSI control codes for cursor positioning and clearing
  - Add color coding for different metric types
  - Implement column alignment logic
  - Add threshold-based warning indicators
  - Implement terminal capability detection and fallback
  - Implement Clear() and Close() methods for cleanup
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 12.1, 12.2, 12.3_

- [ ]* 4.1 Write property test for ANSI code presence
  - **Property 13: Terminal ANSI code presence**
  - **Validates: Requirements 7.1, 7.2, 7.4**

- [ ]* 4.2 Write property test for threshold alerting
  - **Property 16: Threshold alerting consistency**
  - **Validates: Requirements 12.1, 12.2, 12.3**

- [x] 5. Implement JSON renderer
  - Create JSONRenderer struct with JSON encoder
  - Implement Render() method that serializes metrics to JSON
  - Ensure output contains no ANSI codes
  - Output one line per refresh
  - _Requirements: 6.2, 6.3, 6.4_

- [ ]* 5.1 Write property test for JSON serialization completeness
  - **Property 11: JSON serialization completeness**
  - **Validates: Requirements 6.2**

- [ ]* 5.2 Write property test for JSON output purity
  - **Property 12: JSON output format purity**
  - **Validates: Requirements 6.4**

- [x] 6. Implement configuration management
  - Create Config struct with all configuration fields
  - Implement configuration file loading (YAML and JSON support using viper)
  - Implement flag-to-config merging with proper precedence
  - Add validation for interval, threshold, and file path values
  - Implement default values (1 second interval)
  - _Requirements: 5.2, 12.4, 14.1, 14.2, 14.4_

- [ ]* 6.1 Write property test for interval configuration precedence
  - **Property 9: Interval configuration precedence**
  - **Validates: Requirements 5.1, 14.2**

- [ ]* 6.2 Write property test for invalid interval rejection
  - **Property 10: Invalid interval rejection**
  - **Validates: Requirements 5.3**

- [ ]* 6.3 Write property test for threshold configuration validity
  - **Property 17: Threshold configuration validity**
  - **Validates: Requirements 12.4**

- [ ]* 6.4 Write property test for configuration file format support
  - **Property 20: Configuration file format support**
  - **Validates: Requirements 14.4**

- [ ]* 6.5 Write property test for invalid configuration rejection
  - **Property 21: Invalid configuration rejection**
  - **Validates: Requirements 14.3**

- [x] 7. Implement file logger
  - Create FileLogger struct with file handle
  - Implement LogMetrics() method that writes timestamped entries
  - Implement error handling for file I/O failures
  - Implement Close() method to flush and close file
  - _Requirements: 13.1, 13.2, 13.3, 13.4_

- [ ]* 7.1 Write property test for log timestamp presence
  - **Property 18: Log file timestamp presence**
  - **Validates: Requirements 13.2**

- [ ]* 7.2 Write property test for log file error resilience
  - **Property 19: Log file error resilience**
  - **Validates: Requirements 13.3**

- [x] 8. Implement CLI with Cobra
  - Set up Cobra root command with global flags (--interval, --json, --log-file, --config, threshold flags)
  - Implement version subcommand
  - Add flag validation and error handling
  - Implement help text and usage documentation
  - Wire flags to configuration
  - _Requirements: 5.1, 5.3, 6.1, 9.1, 9.2, 9.3, 9.4, 12.4, 13.1, 14.1_

- [ ]* 8.1 Write unit test for version command
  - Test that version subcommand displays version and exits
  - _Requirements: 9.2_

- [ ]* 8.2 Write property test for CLI error handling
  - **Property 23: CLI error handling consistency**
  - **Validates: Requirements 9.3**

- [x] 9. Implement monitor orchestrator
  - Create SystemMonitor struct with collector, renderer, logger, and config
  - Implement Start() method that sets up context and signal handlers
  - Launch collector goroutine and handle metrics channel
  - Implement main loop that receives metrics and renders them
  - Add optional logging to file
  - Implement graceful shutdown on SIGINT/SIGTERM
  - Implement Stop() method for cleanup (close logger, restore terminal)
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 11.1, 11.2, 11.3_

- [ ]* 9.1 Write property test for context cancellation propagation
  - **Property 14: Context cancellation propagation**
  - **Validates: Requirements 8.2, 8.3, 11.3**

- [ ]* 9.2 Write unit tests for signal handling and shutdown
  - Test SIGINT triggers graceful shutdown
  - Test terminal restoration on exit
  - Test exit code is 0 on clean shutdown
  - _Requirements: 8.1, 8.4, 8.5_

- [x] 10. Wire everything together in main
  - Create main.go that initializes and executes root command
  - Set application version
  - Handle top-level errors
  - _Requirements: 15.1, 15.2_

- [x] 11. Create mock implementations for testing
  - Implement MockStatsProvider with controllable data and errors
  - Ensure mock can simulate all success and failure scenarios
  - _Requirements: 10.1, 10.2, 10.3, 10.4_

- [ ]* 11.1 Write property test for mock provider equivalence
  - **Property 22: Mock provider behavioral equivalence**
  - **Validates: Requirements 10.4**

- [x] 12. Add example configuration files
  - Create example YAML configuration file
  - Create example JSON configuration file
  - Add documentation comments
  - _Requirements: 14.4_

- [x] 13. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 14. Create README with usage documentation
  - Document installation instructions
  - Document all CLI flags and commands
  - Provide usage examples
  - Document configuration file format
  - Add troubleshooting section
  - _Requirements: 9.4_

- [x] 15. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.
