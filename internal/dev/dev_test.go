package dev

import (
	"context"
	"runtime"
	"testing"
)

func TestClassifyProcess(t *testing.T) {
	tests := []struct {
		name         string
		wantCategory string
		wantStack    string
	}{
		{"node", "Dev Server", "Node.js/JS"},
		{"/usr/local/bin/vite", "Dev Server", "Node.js/JS"},
		{"python3", "Dev Server", "Python"},
		{"uvicorn", "Dev Server", "Python"},
		{"gopls", "Language Server", "Go"},
		{"rust-analyzer", "Language Server", "Rust"},
		{"docker", "Container", "Docker"},
		{"postgres", "Database", "PostgreSQL"},
		{"redis-server", "Database", "Redis"},
		{"cursor", "Editor / IDE", "IDE"},
		{"launchd", "System", "Other"},
	}

	for _, tc := range tests {
		cat, stack := ClassifyProcess(tc.name)
		if cat != tc.wantCategory || stack != tc.wantStack {
			t.Errorf("ClassifyProcess(%q) = (%q, %q); want (%q, %q)",
				tc.name, cat, stack, tc.wantCategory, tc.wantStack)
		}
	}
}

func TestGetListeningPorts(t *testing.T) {
	ports, err := GetListeningPorts()
	if err != nil {
		t.Fatalf("GetListeningPorts() failed: %v", err)
	}

	t.Logf("Found %d listening ports", len(ports))
	for _, p := range ports {
		if p.Port <= 0 || p.Port > 65535 {
			t.Errorf("invalid port number %d", p.Port)
		}
		if p.PID <= 0 {
			t.Errorf("invalid PID %d for port %d", p.PID, p.Port)
		}
		if p.ProcessName == "" {
			t.Errorf("empty process name for port %d", p.Port)
		}
	}
}

func TestProfileCommand(t *testing.T) {
	profile, err := ProfileCommand(context.Background(), "go", []string{"version"}, nil)
	if err != nil {
		t.Fatalf("ProfileCommand failed: %v", err)
	}

	if profile.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", profile.ExitCode)
	}
	if profile.Duration <= 0 {
		t.Errorf("expected duration > 0, got %v", profile.Duration)
	}
	if profile.PeakRSS == 0 {
		t.Errorf("expected PeakRSS > 0, got %d", profile.PeakRSS)
	}

	t.Logf("Profile: Duration=%v, PeakRSS=%.2f MB",
		profile.Duration, float64(profile.PeakRSS)/(1024*1024))
}

func TestGenerateSnapshot(t *testing.T) {
	snap, err := GenerateSnapshot()
	if err != nil {
		t.Fatalf("GenerateSnapshot failed: %v", err)
	}

	if snap.OS != runtime.GOOS {
		t.Errorf("expected OS %s, got %s", runtime.GOOS, snap.OS)
	}
	if snap.CPUCores <= 0 {
		t.Errorf("expected CPUCores > 0, got %d", snap.CPUCores)
	}
	if snap.MemoryTotal == 0 {
		t.Errorf("expected MemoryTotal > 0")
	}

	md := snap.FormatMarkdown()
	if len(md) == 0 {
		t.Errorf("expected non-empty markdown output")
	}

	t.Logf("Generated snapshot with %d runtimes and %d listening ports",
		len(snap.Runtimes), len(snap.ListeningPorts))
}
