package embedding

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type PythonFindOptions struct {
	RequireAIVectorMemory bool
	PreferredPath         string
}

// FindPython locates a runnable python interpreter.
// On macOS packaged apps, PATH is often minimal; we therefore search common dirs explicitly.
func FindPython(opts PythonFindOptions) string {
	candidates := pythonCandidates(opts.PreferredPath)
	for _, py := range candidates {
		path := resolvePythonPath(py)
		if path == "" {
			continue
		}
		if !isRunnablePython(path) {
			continue
		}
		if opts.RequireAIVectorMemory && !hasAIVectorMemory(path) {
			continue
		}
		return path
	}
	return ""
}

func pythonCandidates(preferred string) []string {
	home, _ := os.UserHomeDir()

	candidates := make([]string, 0, 32)
	if preferred != "" {
		candidates = append(candidates, preferred)
	}

	// Check project venv first (dev convenience)
	candidates = append(candidates,
		filepath.Join(home, "item", "run-memory-mcp-server", ".venv", "bin", "python3"),
		filepath.Join(home, "item", "run-memory-mcp-server", ".venv", "bin", "python"),
	)

	// Names resolved via PATH (we'll resolve against an expanded PATH set)
	candidates = append(candidates, "python3", "python")

	// Common installation paths
	candidates = append(candidates,
		filepath.Join(home, ".pyenv", "shims", "python3"),
		filepath.Join(home, ".pyenv", "shims", "python"),
		filepath.Join(home, "miniconda3", "bin", "python"),
		filepath.Join(home, "anaconda3", "bin", "python"),
		"/usr/local/bin/python3",
		"/usr/bin/python3",
		"/opt/homebrew/bin/python3",
	)

	if runtime.GOOS == "darwin" {
		if p := detectPythonOrgFramework(); p != "" {
			candidates = append([]string{p}, candidates...)
		}
	}

	return candidates
}

func resolvePythonPath(candidate string) string {
	c := strings.TrimSpace(candidate)
	if c == "" {
		return ""
	}
	// Expand "~"
	if strings.HasPrefix(c, "~") {
		home, _ := os.UserHomeDir()
		c = filepath.Join(home, strings.TrimPrefix(c, "~"))
	}

	if filepath.IsAbs(c) {
		if _, err := os.Stat(c); err == nil {
			return c
		}
		return ""
	}

	// For names like "python3", resolve against an expanded PATH so packaged apps can still find Homebrew/Conda.
	pathEnv := expandedPATH()
	for _, dir := range strings.Split(pathEnv, string(os.PathListSeparator)) {
		if dir == "" {
			continue
		}
		p := filepath.Join(dir, c)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// Fallback to LookPath with current PATH (dev shells usually work)
	if found, err := exec.LookPath(c); err == nil {
		return found
	}
	return ""
}

func expandedPATH() string {
	base := os.Getenv("PATH")
	home, _ := os.UserHomeDir()

	// macOS GUI apps often miss Homebrew/pyenv/conda entries.
	extras := []string{
		"/opt/homebrew/bin",
		"/usr/local/bin",
		filepath.Join(home, ".pyenv", "shims"),
		filepath.Join(home, "miniconda3", "bin"),
		filepath.Join(home, "anaconda3", "bin"),
	}

	seen := map[string]bool{}
	out := make([]string, 0, 16)
	for _, part := range strings.Split(base, string(os.PathListSeparator)) {
		part = strings.TrimSpace(part)
		if part == "" || seen[part] {
			continue
		}
		seen[part] = true
		out = append(out, part)
	}
	for _, part := range extras {
		part = strings.TrimSpace(part)
		if part == "" || seen[part] {
			continue
		}
		seen[part] = true
		out = append(out, part)
	}
	return strings.Join(out, string(os.PathListSeparator))
}

func isRunnablePython(path string) bool {
	// Filters out architecture-mismatched interpreters (e.g. arm64 python from an x64 app),
	// as well as broken installations.
	out, err := exec.Command(path, "-c", "import sys; print(sys.version_info[0])").CombinedOutput()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "3"
}

func hasAIVectorMemory(path string) bool {
	out, err := exec.Command(path, "-c", "import aivectormemory; print('ok')").CombinedOutput()
	return err == nil && strings.TrimSpace(string(out)) == "ok"
}

func detectPythonOrgFramework() string {
	versDir := "/Library/Frameworks/Python.framework/Versions"
	entries, err := os.ReadDir(versDir)
	if err != nil {
		return ""
	}
	best := ""
	bestKey := ""
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if name == "Current" {
			continue
		}
		py := filepath.Join(versDir, name, "bin", "python3")
		if _, err := os.Stat(py); err != nil {
			continue
		}
		key := padVersionKey(name)
		if best == "" || key > bestKey {
			best = py
			bestKey = key
		}
	}
	return best
}

func padVersionKey(v string) string {
	parts := strings.Split(v, ".")
	if len(parts) == 0 {
		return v
	}
	for len(parts) < 3 {
		parts = append(parts, "0")
	}
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			out = append(out, p)
			continue
		}
		out = append(out, leftPad3(n))
	}
	return strings.Join(out, ".")
}

func leftPad3(n int) string {
	s := strconv.Itoa(n)
	if len(s) == 1 {
		return "00" + s
	}
	if len(s) == 2 {
		return "0" + s
	}
	return s
}

