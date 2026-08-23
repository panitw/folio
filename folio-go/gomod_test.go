package folio

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestGoModPinsToolchainExactly is AC1: go.mod carries a literal
// "toolchain go1.26.0" line (exact version, no range) and a literal
// "go 1.25.0" line. Both lines are asserted, not just the toolchain one:
// Go strips a toolchain directive that is not strictly greater than the
// go directive (F-1, RP-9), so a go.mod with "go 1.26.0" and
// "toolchain go1.26.0" would look pinned right up until `go mod tidy`
// silently drops the pin — the low `go` line is what makes the pin
// survive, and this test must not be satisfiable by that broken shape.
func TestGoModPinsToolchainExactly(t *testing.T) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}

	lines := strings.Split(string(data), "\n")

	hasGoLine := false
	hasToolchainLine := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "go 1.25.0" {
			hasGoLine = true
		}
		if trimmed == "toolchain go1.26.0" {
			hasToolchainLine = true
		}
	}

	if !hasGoLine {
		t.Error(`go.mod does not contain a literal "go 1.25.0" line`)
	}
	if !hasToolchainLine {
		t.Error(`go.mod does not contain a literal "toolchain go1.26.0" line`)
	}
}

// TestModuleGraphHasNoThirdPartyDependencies is AC2: `go list -m all`
// returns exactly the main module — no third-party PDF writer (or any
// other dependency) anywhere in the resolved module graph.
func TestModuleGraphHasNoThirdPartyDependencies(t *testing.T) {
	cmd := exec.Command("go", "list", "-m", "all")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("go list -m all failed: %v\nstderr: %s", err, stderr.String())
	}

	lines := []string{}
	for _, l := range strings.Split(strings.TrimSpace(stdout.String()), "\n") {
		if strings.TrimSpace(l) != "" {
			lines = append(lines, strings.TrimSpace(l))
		}
	}

	if len(lines) != 1 {
		t.Fatalf("expected exactly one module in the graph (the main module), got %d: %v", len(lines), lines)
	}
	if !strings.HasPrefix(lines[0], "github.com/panitw/folio/folio-go") {
		t.Fatalf("unexpected main module entry: %q", lines[0])
	}
}
