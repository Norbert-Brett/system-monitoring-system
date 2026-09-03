package dev

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ProfileOptions configures a command profiling run
type ProfileOptions struct {
	Stdout io.Writer
	Stderr io.Writer
}

// ProfileCommand runs the command and samples resource usage until completion
func ProfileCommand(ctx context.Context, command string, args []string, opts *ProfileOptions) (*CommandProfile, error) {
	if opts == nil {
		opts = &ProfileOptions{
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}
	}

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdout = opts.Stdout
	cmd.Stderr = opts.Stderr

	startTime := time.Now()
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	pid := cmd.Process.Pid
	var peakRSS uint64
	var mu sync.Mutex

	// Goroutine to sample memory every 25ms
	ticker := time.NewTicker(25 * time.Millisecond)
	stopSampling := make(chan struct{})

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-stopSampling:
				return
			case <-ticker.C:
				rss := sampleProcessRSS(pid)
				mu.Lock()
				if rss > peakRSS {
					peakRSS = rss
				}
				mu.Unlock()
			}
		}
	}()

	waitErr := cmd.Wait()
	close(stopSampling)
	duration := time.Since(startTime)

	exitCode := 0
	if waitErr != nil {
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	var userCPU, sysCPU time.Duration
	var rusageRSS uint64
	if cmd.ProcessState != nil {
		userCPU = cmd.ProcessState.UserTime()
		sysCPU = cmd.ProcessState.SystemTime()
		if ru, ok := cmd.ProcessState.SysUsage().(*syscall.Rusage); ok {
			if runtime.GOOS == "darwin" {
				rusageRSS = uint64(ru.Maxrss)
			} else {
				// Linux reports in KB
				rusageRSS = uint64(ru.Maxrss) * 1024
			}
		}
	}

	mu.Lock()
	finalPeak := peakRSS
	mu.Unlock()

	if rusageRSS > finalPeak {
		finalPeak = rusageRSS
	}

	return &CommandProfile{
		Command:   command,
		Args:      args,
		ExitCode:  exitCode,
		Duration:  duration,
		PeakRSS:   finalPeak,
		UserCPU:   userCPU,
		SysCPU:    sysCPU,
		Timestamp: startTime,
	}, nil
}

func sampleProcessRSS(pid int) uint64 {
	if pid <= 0 {
		return 0
	}

	if runtime.GOOS == "linux" {
		// Read /proc/<pid>/statm
		data, err := os.ReadFile(fmt.Sprintf("/proc/%d/statm", pid))
		if err == nil {
			fields := strings.Fields(string(data))
			if len(fields) >= 2 {
				pages, _ := strconv.ParseUint(fields[1], 10, 64)
				return pages * uint64(os.Getpagesize())
			}
		}
	}

	// Fallback using ps command
	out, err := exec.Command("ps", "-o", "rss=", "-p", strconv.Itoa(pid)).Output()
	if err == nil {
		rssKB, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
		if err == nil {
			return rssKB * 1024
		}
	}

	return 0
}
