package licence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// moduleRecord mirrors the subset of `go list -m -json`'s per-module
// object this package needs.
type moduleRecord struct {
	Path    string
	Version string
	Main    bool
	Dir     string
	Replace *moduleRecord
}

// ResolvedModule is one non-main module in a graph, with the on-disk
// directory Go resolved it to (empty if unresolved) — Dir empty is
// AC19's "unresolvable" case.
type ResolvedModule struct {
	Path    string
	Version string
	Dir     string
}

// ResolveGraph runs `go list -m -json all` inside moduleDir, hermetically
// — GOPROXY=off, so it can only ever resolve what a local `replace` or an
// already-populated module cache provides (AC30, RP-12: proving the
// checker resolves without fetching is both CI hermeticity and what
// makes retaining a fake graph safe) — and returns every non-main module
// in the graph.
func ResolveGraph(moduleDir string) ([]ResolvedModule, error) {
	cmd := exec.Command("go", "list", "-m", "-json", "all")
	cmd.Dir = moduleDir
	cmd.Env = append(os.Environ(), "GOPROXY=off", "GOFLAGS=-mod=mod")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("go list -m -json all in %s: %w (stderr: %s)", moduleDir, err, stderr.String())
	}

	var modules []ResolvedModule
	dec := json.NewDecoder(&stdout)
	for dec.More() {
		var rec moduleRecord
		if err := dec.Decode(&rec); err != nil {
			return nil, fmt.Errorf("decode module list in %s: %w", moduleDir, err)
		}
		if rec.Main {
			continue
		}
		dir := rec.Dir
		if rec.Replace != nil {
			dir = rec.Replace.Dir
			if dir == "" {
				dir = filepath.Join(moduleDir, rec.Replace.Path)
			}
		}
		modules = append(modules, ResolvedModule{Path: rec.Path, Version: rec.Version, Dir: dir})
	}
	return modules, nil
}

// licenceFileNames is the closed set of filenames this package looks
// for at a resolved module's root.
var licenceFileNames = []string{"LICENSE", "LICENSE.txt", "LICENSE.md", "COPYING"}

// ReadLicenceText returns the content of the first licence file found
// at the module's resolved directory, or ("", false) if none exists —
// AC19's "unresolvable" case, or a module with no licence file at all
// (AC29's unknown/ fixture).
func ReadLicenceText(moduleDir string) (string, bool) {
	if moduleDir == "" {
		return "", false
	}
	for _, name := range licenceFileNames {
		data, err := os.ReadFile(filepath.Join(moduleDir, name))
		if err == nil {
			return string(data), true
		}
	}
	return "", false
}
