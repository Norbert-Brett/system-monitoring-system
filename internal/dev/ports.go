package dev

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// GetListeningPorts retrieves all active listening TCP ports with process and project metadata
func GetListeningPorts() ([]ListeningPort, error) {
	if runtime.GOOS == "darwin" {
		return getDarwinListeningPorts()
	}
	return getLinuxListeningPorts()
}

// KillPort terminates the process listening on the specified port
func KillPort(port int, force bool) (*ListeningPort, error) {
	ports, err := GetListeningPorts()
	if err != nil {
		return nil, fmt.Errorf("failed to list ports: %w", err)
	}

	var target *ListeningPort
	for _, p := range ports {
		if p.Port == port {
			target = &p
			break
		}
	}

	if target == nil {
		return nil, fmt.Errorf("no active process found listening on port %d", port)
	}

	proc, err := os.FindProcess(target.PID)
	if err != nil {
		return nil, fmt.Errorf("failed to find process with PID %d: %w", target.PID, err)
	}

	sig := syscall.SIGTERM
	if force {
		sig = syscall.SIGKILL
	}

	if err := proc.Signal(sig); err != nil {
		return nil, fmt.Errorf("failed to send signal to PID %d: %w", target.PID, err)
	}

	// Wait up to 2.5 seconds for the port to be released
	deadline := time.Now().Add(2500 * time.Millisecond)
	for time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
		currentPorts, _ := GetListeningPorts()
		stillOpen := false
		for _, cp := range currentPorts {
			if cp.Port == port && cp.PID == target.PID {
				stillOpen = true
				break
			}
		}
		if !stillOpen {
			return target, nil
		}
	}

	// If still open and was not forced, attempt SIGKILL
	if !force {
		_ = proc.Signal(syscall.SIGKILL)
		time.Sleep(200 * time.Millisecond)
	}

	return target, nil
}

func getDarwinListeningPorts() ([]ListeningPort, error) {
	out, err := exec.Command("lsof", "-iTCP", "-sTCP:LISTEN", "-P", "-n").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute lsof: %w", err)
	}

	var results []ListeningPort
	seen := make(map[int]bool)
	cwdCache := make(map[int]string)

	scanner := bufio.NewScanner(bytes.NewReader(out))
	if scanner.Scan() {
		// skip header
	}

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 9 {
			continue
		}

		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}

		addr := fields[8]
		idx := strings.LastIndex(addr, ":")
		if idx == -1 {
			continue
		}

		portStr := addr[idx+1:]
		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}

		if seen[port] {
			continue
		}
		seen[port] = true

		cwd, exists := cwdCache[pid]
		if !exists {
			cwd = getProcessCWD(pid)
			cwdCache[pid] = cwd
		}

		procName := fields[0]
		category, stack := ClassifyProcess(procName)

		project := ""
		if cwd != "" && cwd != "/" {
			project = filepath.Base(cwd)
		}

		results = append(results, ListeningPort{
			Port:        port,
			Protocol:    "TCP",
			PID:         pid,
			ProcessName: procName,
			User:        fields[2],
			CWD:         cwd,
			Project:     project,
			Category:    category,
			Stack:       stack,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Port < results[j].Port
	})

	return results, nil
}

func getLinuxListeningPorts() ([]ListeningPort, error) {
	// Try 'ss' command first
	out, err := exec.Command("ss", "-tlpn").Output()
	if err == nil {
		return parseSSOutput(out)
	}

	// Fallback to 'lsof'
	out, err = exec.Command("lsof", "-iTCP", "-sTCP:LISTEN", "-P", "-n").Output()
	if err == nil {
		return parseLsofLinux(out)
	}

	return nil, fmt.Errorf("neither ss nor lsof available on Linux")
}

func parseSSOutput(output []byte) ([]ListeningPort, error) {
	var results []ListeningPort
	seen := make(map[int]bool)
	scanner := bufio.NewScanner(bytes.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "State") || strings.HasPrefix(line, "Netid") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		// Local address is typically field 3
		localAddr := fields[3]
		idx := strings.LastIndex(localAddr, ":")
		if idx == -1 {
			continue
		}
		port, err := strconv.Atoi(localAddr[idx+1:])
		if err != nil || seen[port] {
			continue
		}

		pid := 0
		procName := "unknown"
		// Process info is in the last field e.g. users:(("node",pid=1234,fd=19))
		if len(fields) >= 6 {
			procInfo := fields[len(fields)-1]
			if strings.Contains(procInfo, "pid=") {
				parts := strings.Split(procInfo, "pid=")
				if len(parts) > 1 {
					pidStr := strings.Split(parts[1], ",")[0]
					pidStr = strings.TrimSuffix(pidStr, ")")
					pid, _ = strconv.Atoi(pidStr)
				}
			}
			if strings.Contains(procInfo, "\"") {
				parts := strings.Split(procInfo, "\"")
				if len(parts) > 1 {
					procName = parts[1]
				}
			}
		}

		seen[port] = true
		cwd := getProcessCWD(pid)
		category, stack := ClassifyProcess(procName)

		project := ""
		if cwd != "" && cwd != "/" {
			project = filepath.Base(cwd)
		}

		results = append(results, ListeningPort{
			Port:        port,
			Protocol:    "TCP",
			PID:         pid,
			ProcessName: procName,
			CWD:         cwd,
			Project:     project,
			Category:    category,
			Stack:       stack,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Port < results[j].Port
	})

	return results, nil
}

func parseLsofLinux(output []byte) ([]ListeningPort, error) {
	var results []ListeningPort
	seen := make(map[int]bool)
	scanner := bufio.NewScanner(bytes.NewReader(output))
	if scanner.Scan() {
		// skip header
	}

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 9 {
			continue
		}
		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		addr := fields[8]
		idx := strings.LastIndex(addr, ":")
		if idx == -1 {
			continue
		}
		port, err := strconv.Atoi(addr[idx+1:])
		if err != nil || seen[port] {
			continue
		}
		seen[port] = true

		procName := fields[0]
		cwd := getProcessCWD(pid)
		category, stack := ClassifyProcess(procName)

		project := ""
		if cwd != "" && cwd != "/" {
			project = filepath.Base(cwd)
		}

		results = append(results, ListeningPort{
			Port:        port,
			Protocol:    "TCP",
			PID:         pid,
			ProcessName: procName,
			User:        fields[2],
			CWD:         cwd,
			Project:     project,
			Category:    category,
			Stack:       stack,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Port < results[j].Port
	})

	return results, nil
}

func getProcessCWD(pid int) string {
	if pid <= 0 {
		return ""
	}

	if runtime.GOOS == "linux" {
		link, err := os.Readlink(fmt.Sprintf("/proc/%d/cwd", pid))
		if err == nil {
			return link
		}
	}

	if runtime.GOOS == "darwin" {
		out, err := exec.Command("lsof", "-a", "-p", strconv.Itoa(pid), "-d", "cwd", "-Fn").Output()
		if err == nil {
			scanner := bufio.NewScanner(bytes.NewReader(out))
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "n") && len(line) > 1 {
					return line[1:]
				}
			}
		}
	}

	return ""
}
