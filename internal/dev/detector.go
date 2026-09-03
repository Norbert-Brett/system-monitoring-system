package dev

import (
	"path/filepath"
	"strings"
)

// ClassifyProcess inspects a process name and returns its category and tech stack
func ClassifyProcess(name string) (category string, stack string) {
	lower := strings.ToLower(name)
	base := filepath.Base(lower)

	switch {
	// Language Servers & Tooling
	case strings.Contains(base, "gopls"):
		return "Language Server", "Go"
	case strings.Contains(base, "rust-analyzer"):
		return "Language Server", "Rust"
	case strings.Contains(base, "tsserver") || strings.Contains(base, "vtsls"):
		return "Language Server", "TypeScript"
	case strings.Contains(base, "pyright") || strings.Contains(base, "pylsp"):
		return "Language Server", "Python"
	case strings.Contains(base, "clangd"):
		return "Language Server", "C/C++"
	case strings.Contains(base, "language_") || strings.Contains(base, "languageserver") || strings.Contains(base, "language-server"):
		return "Language Server", "LSP"

	// Containers & Virtualization
	case strings.Contains(base, "docker") || strings.Contains(base, "containerd"):
		return "Container", "Docker"
	case strings.Contains(base, "orbstack"):
		return "Container", "OrbStack"
	case strings.Contains(base, "podman"):
		return "Container", "Podman"
	case strings.Contains(base, "colima"):
		return "Container", "Colima"

	// Databases & Caches
	case strings.Contains(base, "postgres"):
		return "Database", "PostgreSQL"
	case strings.Contains(base, "mysql") || strings.Contains(base, "mariadb"):
		return "Database", "MySQL"
	case strings.Contains(base, "redis"):
		return "Database", "Redis"
	case strings.Contains(base, "mongo"):
		return "Database", "MongoDB"
	case strings.Contains(base, "elastic"):
		return "Database", "Elasticsearch"

	// Dev Servers & Runtimes
	case strings.Contains(base, "node") || strings.Contains(base, "npm") || strings.Contains(base, "pnpm") ||
		strings.Contains(base, "yarn") || strings.Contains(base, "bun") || strings.Contains(base, "deno") ||
		strings.Contains(base, "next") || strings.Contains(base, "vite") || strings.Contains(base, "webpack"):
		return "Dev Server", "Node.js/JS"

	case strings.Contains(base, "python") || strings.Contains(base, "uvicorn") || strings.Contains(base, "gunicorn") ||
		strings.Contains(base, "fastapi") || strings.Contains(base, "flask") || strings.Contains(base, "django"):
		return "Dev Server", "Python"

	case strings.Contains(base, "cargo") || strings.Contains(base, "rustc"):
		return "Build Tool", "Rust"

	case strings.Contains(base, "go") || strings.Contains(base, "air"):
		return "Dev Server", "Go"

	case strings.Contains(base, "java") || strings.Contains(base, "gradle") || strings.Contains(base, "mvn"):
		return "Dev Server", "Java/JVM"

	case strings.Contains(base, "ruby") || strings.Contains(base, "rails"):
		return "Dev Server", "Ruby"

	// Editors / IDEs
	case strings.Contains(base, "code") || strings.Contains(base, "cursor") ||
		strings.Contains(base, "antigravi") ||
		strings.Contains(base, "electron") ||
		strings.Contains(base, "nvim") || strings.Contains(base, "vim") ||
		strings.Contains(base, "idea") || strings.Contains(base, "goland"):
		return "Editor / IDE", "IDE"

	default:
		return "System", "Other"
	}
}
