package licence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ResolvedNPMPackage is a single direct or transitive npm package named by a
// committed package-lock.json. The lockfile is the complete input: node_modules
// is ignored, host-specific state and is never consulted by AD-26.
type ResolvedNPMPackage struct {
	Path, Version, Licence string
	Dev                    bool
}

type npmLock struct {
	LockfileVersion int                       `json:"lockfileVersion"`
	Packages        map[string]npmLockPackage `json:"packages"`
}
type npmLockPackage struct {
	Version  string `json:"version"`
	Optional bool   `json:"optional"`
	Dev      bool   `json:"dev"`
	Licence  string `json:"license"`
}

func ResolveNPMGraph(designerDir string) ([]ResolvedNPMPackage, error) {
	lockPath := filepath.Join(designerDir, "package-lock.json")
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return nil, fmt.Errorf("read committed npm lockfile %s: %w", lockPath, err)
	}
	var lock npmLock
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("parse npm lockfile %s: %w", lockPath, err)
	}
	if lock.LockfileVersion < 2 || len(lock.Packages) == 0 {
		return nil, fmt.Errorf("npm lockfile %s has no package graph", lockPath)
	}
	var keys []string
	for key := range lock.Packages {
		if strings.Contains(key, "node_modules/") {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return nil, fmt.Errorf("npm lockfile %s resolved zero dependencies", lockPath)
	}
	packages := make([]ResolvedNPMPackage, 0, len(keys))
	for _, key := range keys {
		entry := lock.Packages[key]
		name := key[strings.LastIndex(key, "node_modules/")+len("node_modules/"):]
		if entry.Version == "" {
			return nil, fmt.Errorf("%s: lockfile record has no version", name)
		}
		if strings.TrimSpace(entry.Licence) == "" {
			return nil, fmt.Errorf("%s: lockfile record has no licence", name)
		}
		if _, err := ClassifySPDXExpression(entry.Licence); err != nil {
			return nil, fmt.Errorf("%s: lockfile licence %q is unrecognised: %w", name, entry.Licence, err)
		}
		packages = append(packages, ResolvedNPMPackage{Path: name, Version: entry.Version, Licence: entry.Licence, Dev: entry.Dev})
	}
	return packages, nil
}
