//go:build matrix

// Package folio's matrix_test.go is Story 1.2's cross-target byte-identity
// harness. It is gated behind the "matrix" build tag (AC12) so the routine
// `go test ./...` path — which runs on every story under D-000.4 — never
// spawns Docker or Node. `go vet -tags=matrix ./...` still compiles it
// (RP-8), which is Story 1.2's per-story obligation to prove the tagged
// code compiles even though its full run is deferred to this story only.
//
// It lives in package folio, not a separate package or module, so it can
// call assertWellFormedPDF, firstDivergence and hexWindow directly — all
// three already exist in render_test.go for the same purpose (byte-level
// PDF validation and divergence diagnostics) and this story reuses them
// rather than re-deriving a second copy (AD-21, vacuity guard 3).
//
// AD-21 governs: the same render must produce identical bytes on
// darwin/arm64, linux/amd64, linux/arm64 and js/wasm. Nothing here changes
// how anything is rendered (folio.Render's signature is untouched, D-1.1.c)
// — this file only captures, hashes and compares what already exists.
package folio

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// matrixToolchain is the exact Go toolchain every leg of this harness
// pins to (AD-22, D-1.2.4) — the same "go1.26.0" go.mod already carries
// as its literal toolchain directive.
const matrixToolchain = "go1.26.0"

// init exports CGO_ENABLED=0 and GOTOOLCHAIN=go1.26.0 from inside the
// harness itself (AC2, D-1.2.4). A local run must be exactly as pinned as
// a CI run — D-000.4's epic-boundary gate runs this harness on a laptop,
// and a harness that only trusted the caller's shell to have set these
// would let that gate silently measure an unpinned build. Setting them in
// init(), before any test runs, means every build/run subprocess this
// file spawns inherits them via os.Environ(), and TestHarnessExportsPins
// below asserts they actually took effect.
func init() {
	if err := os.Setenv("CGO_ENABLED", "0"); err != nil {
		panic("matrix harness: set CGO_ENABLED: " + err.Error())
	}
	if err := os.Setenv("GOTOOLCHAIN", matrixToolchain); err != nil {
		panic("matrix harness: set GOTOOLCHAIN: " + err.Error())
	}
}

// matrixTarget is one of the four build targets AD-21 requires.
type matrixTarget struct {
	name   string // "darwin/arm64" etc — used verbatim in every message
	goos   string
	goarch string
}

// matrixTargets is the target list AC1/AC11 require: exactly these four,
// as data, so every place that needs "all four targets" iterates the same
// list rather than re-enumerating it.
var matrixTargets = []matrixTarget{
	{"darwin/arm64", "darwin", "arm64"},
	{"linux/amd64", "linux", "amd64"},
	{"linux/arm64", "linux", "arm64"},
	{"js/wasm", "js", "wasm"},
}

// wantMatrixLegs is AC11's counted-four: the literal constant every leg
// count in this file must equal. It is deliberately NOT len(matrixTargets)
// — Story 1.2's own QA review (Blocker 1) measured that comparing a loop's
// output length against len(matrixTargets), the loop's own input, can
// never fail for any input: deleting a target from matrixTargets shrinks
// both sides together and the check stays green on three legs. Comparing
// against this literal constant is what makes a shortened matrixTargets
// list, or a shortened results/legs count, an actual failure.
const wantMatrixLegs = 4

// assertExactlyFourMatrixTargets is the first line of defence: it fails
// immediately if matrixTargets itself has been narrowed, before any leg
// even starts building. This is what makes the red-proof for Blocker 1
// (deleting the "js/wasm" entry) fail loudly instead of silently running
// three legs.
func assertExactlyFourMatrixTargets(t *testing.T) {
	t.Helper()
	if len(matrixTargets) != wantMatrixLegs {
		t.Fatalf(
			"matrixTargets must list exactly %d targets, got %d: %s (AC11, vacuity guard 4)",
			wantMatrixLegs, len(matrixTargets), matrixTargetNames(),
		)
	}
}

func (m matrixTarget) isNative() bool {
	return m.goos == runtime.GOOS && m.goarch == runtime.GOARCH
}

func (m matrixTarget) isWasm() bool {
	return m.goos == "js" && m.goarch == "wasm"
}

func findMatrixTarget(name string) (matrixTarget, bool) {
	for _, m := range matrixTargets {
		if m.name == name {
			return m, true
		}
	}
	return matrixTarget{}, false
}

func matrixTargetNames() string {
	names := make([]string, len(matrixTargets))
	for i, m := range matrixTargets {
		names[i] = m.name
	}
	return strings.Join(names, ", ")
}

// TestHarnessExportsPins is AC2: the harness's own build environment must
// carry CGO_ENABLED=0 and GOTOOLCHAIN=go1.26.0, asserted rather than
// trusted, because every build step in this file (buildEnv) relies on
// them having actually been set by init() above.
func TestHarnessExportsPins(t *testing.T) {
	if got := os.Getenv("CGO_ENABLED"); got != "0" {
		t.Fatalf("the harness itself must export CGO_ENABLED=0, not merely rely on the caller's shell (AC2, D-1.2.4) — got %q", got)
	}
	if got := os.Getenv("GOTOOLCHAIN"); got != matrixToolchain {
		t.Fatalf("the harness itself must export GOTOOLCHAIN=%s, not merely rely on the caller's shell (AC2, D-1.2.4) — got %q", matrixToolchain, got)
	}
}

// matrixBuildDir returns folio-go/.matrix-build, creating it if needed.
// Build outputs go inside the repo (not t.TempDir()) so the Docker bind
// mount used by the two Linux legs lands inside a path Docker Desktop
// shares by default, and so a failure leaves inspectable artifacts behind
// (F-7). It is git-ignored (AC12).
func matrixBuildDir(t *testing.T, root string) string {
	t.Helper()
	dir := filepath.Join(root, "folio-go", ".matrix-build")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create matrix build dir %s: %v", dir, err)
	}
	return dir
}

// buildEnv is the subprocess environment for every build step: the
// current process environment (which already carries init()'s pins) plus
// an explicit, redundant CGO_ENABLED/GOTOOLCHAIN pin and the target's
// GOOS/GOARCH. Go's exec de-duplicates repeated keys in favour of the
// last occurrence, so these explicit entries win even if something
// upstream set them differently.
//
// Every leg — including the two Linux ones — is built here, on the host,
// with the one pinned host toolchain (F-6, D-1.2.1). AD-22 already pins
// the toolchain, so toolchain variance is a deliberately excluded
// variable; AD-21's matrix exists to isolate target codegen, which is
// where FMA contraction lives. Building all four with one compiler tests
// exactly that variable and removes container-Go-version drift as a
// confound — a container that built its own binary would vary two things
// at once and attribute the result to one. This is why the two Linux
// binaries are cross-compiled here and merely executed inside their
// containers (see runDocker), and why this harness never falls back to
// `go run`, which would put a toolchain inside the container.
func buildEnv(target matrixTarget) []string {
	env := os.Environ()
	env = append(env,
		"CGO_ENABLED=0",
		"GOTOOLCHAIN="+matrixToolchain,
		"GOOS="+target.goos,
		"GOARCH="+target.goarch,
	)
	return env
}

// buildRenderTestBinary builds folio-go's own package test binary for
// target with `go test -c` (F-3, D-1.2.1): no new main package, no second
// rendering path. The binary is later run with FOLIO_SUBPROCESS_RENDER=1
// or FOLIO_SUBPROCESS_TOOLCHAIN=1, both of which TestMain (render_test.go)
// handles before m.Run() is ever reached, so command-line arguments are
// irrelevant and none are passed.
func buildRenderTestBinary(t *testing.T, root string, target matrixTarget, buildDir string) string {
	t.Helper()
	binName := "folio." + target.goos + "-" + target.goarch + ".test"
	binPath := filepath.Join(buildDir, binName)
	cmd := exec.Command("go", "test", "-c", "-o", binPath, ".")
	cmd.Dir = filepath.Join(root, "folio-go")
	cmd.Env = buildEnv(target)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build of %s render/toolchain test binary failed: %v\nstderr: %s", target.name, err, stderr.String())
	}
	return binPath
}

// buildProbeBinary builds hashmatrix/probe for target with a plain
// `go build`, run from inside the hashmatrix module (F-5): hashmatrix has
// no dependency on folio-go, so this is an entirely separate build.
func buildProbeBinary(t *testing.T, root string, target matrixTarget, buildDir string) string {
	t.Helper()
	binName := "probe." + target.goos + "-" + target.goarch
	binPath := filepath.Join(buildDir, binName)
	cmd := exec.Command("go", "build", "-o", binPath, "./probe")
	cmd.Dir = filepath.Join(root, "hashmatrix")
	cmd.Env = buildEnv(target)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build of probe binary for %s failed: %v\nstderr: %s", target.name, err, stderr.String())
	}
	return binPath
}

// envWith returns the current process environment plus extra key=value
// pairs appended (and so, by Go's de-duplication-favours-last rule,
// effectively overriding any existing value for the same key).
func envWith(extra map[string]string) []string {
	env := os.Environ()
	for k, v := range extra {
		env = append(env, k+"="+v)
	}
	return env
}

// runOnTarget is the runner selector (Task 5): native when the target
// matches the host, Docker for a non-matching Linux target, and Node's
// go_js_wasm_exec for js/wasm. Every branch fails loudly and actionably
// on a missing prerequisite — it never skips (AC11, RP-7): a missing
// Docker daemon, a missing linux/arm64 platform, or a missing Node
// installation is a defect in the run, not a reason to quietly report
// fewer legs.
func runOnTarget(t *testing.T, target matrixTarget, binPath string, env map[string]string, args ...string) []byte {
	t.Helper()
	switch {
	case target.isNative():
		return runNative(t, binPath, env, args...)
	case target.isWasm():
		return runWasm(t, binPath, env, args...)
	case target.goos == "linux":
		return runDocker(t, target, binPath, env, args...)
	default:
		t.Fatalf("no runner defined for target %s", target.name)
		return nil
	}
}

func runNative(t *testing.T, binPath string, env map[string]string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command(binPath, args...)
	cmd.Env = envWith(env)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("native run of %s failed: %v\nstderr: %s", binPath, err, stderr.String())
	}
	return stdout.Bytes()
}

// runDocker runs a pre-built Linux binary inside a container that carries
// no Go toolchain and no repo mount (F-6, D-1.2.1): the bind mount below
// carries only the prebuilt binaries directory, read-only. A CGO_ENABLED=0
// pure-Go static binary makes the base image a process host only — it
// cannot contribute to the output bytes — so alpine:3.20 is a convenience
// default, not part of the contract, and is overridable via
// FOLIO_MATRIX_DOCKER_IMAGE. Any failure here (missing daemon, missing
// --platform support) is fatal with Docker's own actionable stderr, never
// a skip (AC11, RP-7).
func runDocker(t *testing.T, target matrixTarget, binPath string, env map[string]string, args ...string) []byte {
	t.Helper()

	buildDir := filepath.Dir(binPath)
	binName := filepath.Base(binPath)
	image := os.Getenv("FOLIO_MATRIX_DOCKER_IMAGE")
	if image == "" {
		image = "alpine:3.20"
	}

	dockerArgs := []string{"run", "--rm", "--platform", target.name}
	for k, v := range env {
		dockerArgs = append(dockerArgs, "-e", k+"="+v)
	}
	dockerArgs = append(dockerArgs, "-v", buildDir+":/b:ro", image, "/b/"+binName)
	dockerArgs = append(dockerArgs, args...)

	cmd := exec.Command("docker", dockerArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf(
			"docker run for target %s FAILED — this is a failure, not a skip (AC11). "+
				"Is the Docker daemon running, and is --platform %s available (buildx)? "+
				"The container carries no Go toolchain and no repo mount by design (F-6); "+
				"this harness never falls back to `go run`.\nerror: %v\nstderr: %s",
			target.name, target.name, err, stderr.String(),
		)
	}
	return stdout.Bytes()
}

// runWasm runs a js/wasm binary under Node via the Go toolchain's own
// wrapper, resolved from `go env GOROOT` at runtime (AC4) — never a
// hard-coded Homebrew path, never a vendored copy of wasm_exec.js. Node's
// presence is asserted, not assumed.
func runWasm(t *testing.T, binPath string, env map[string]string, args ...string) []byte {
	t.Helper()

	wrapper := wasmExecPath(t)
	cmdArgs := append([]string{binPath}, args...)
	cmd := exec.Command(wrapper, cmdArgs...)
	cmd.Env = envWith(env)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf(
			"wasm run of %s FAILED — this is a failure, not a skip (AC11). Is Node.js on PATH? "+
				"wrapper=%s\nerror: %v\nstderr: %s",
			binPath, wrapper, err, stderr.String(),
		)
	}
	return stdout.Bytes()
}

func wasmExecPath(t *testing.T) string {
	t.Helper()

	if _, err := exec.LookPath("node"); err != nil {
		t.Fatalf("node not found on PATH; the js/wasm leg requires Node.js (AC4): %v", err)
	}

	cmd := exec.Command("go", "env", "GOROOT")
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("go env GOROOT failed: %v\nstderr: %s", err, stderr.String())
	}
	goroot := strings.TrimSpace(out.String())
	wrapper := filepath.Join(goroot, "lib", "wasm", "go_js_wasm_exec")
	if _, err := os.Stat(wrapper); err != nil {
		t.Fatalf("wasm exec wrapper not found at %s (AC4: resolved from `go env GOROOT`, never a vendored wasm_exec.js): %v", wrapper, err)
	}
	return wrapper
}

// toolchainWitness runs binPath with FOLIO_SUBPROCESS_TOOLCHAIN=1 and
// returns what it reports. runtime.Version() compiled into the binary is
// the only honest witness of what built it (F-9): the host's `go version`
// describes the machine, not the artifact, and in CI the four legs are
// built on three different runners.
func toolchainWitness(t *testing.T, target matrixTarget, binPath string) string {
	t.Helper()
	out := runOnTarget(t, target, binPath, map[string]string{toolchainEnvVar: "1"})
	return strings.TrimSpace(string(out))
}

// assertToolchainWitness is AC3's per-target toolchain gate: it requires
// binPath to report exactly matrixToolchain AND to equal the fixture's
// recorded goToolchain, before any hash is compared, with no skip path.
//
// This is deliberately NOT harmonised with TestRenderMatchesGoldenFixture
// (fixture_test.go)'s toolchain gate, which t.Fatal()s only AFTER
// structural validation — a contributor-local affordance the Story 1.1
// finisher deliberately introduced (Finding 5) so a contributor on a
// different toolchain still observes the document's self-consistency
// instead of nothing at all. This harness has no such affordance: a
// matrix comparing hashes across mismatched toolchains is not a weaker
// test, it is a meaningless one whose green result actively misleads
// (AD-22), so off-toolchain means fail, full stop, before anything else
// runs. The asymmetry is deliberate — do not unify the two gates.
func assertToolchainWitness(t *testing.T, target matrixTarget, binPath string, fixtureToolchain string) {
	t.Helper()
	got := toolchainWitness(t, target, binPath)
	if got != matrixToolchain {
		t.Fatalf(
			"target %s: toolchain witness is %q, want the pinned %q (AD-22, counter-metric C6). "+
				"This check runs before any hash is compared and has no skip path.",
			target.name, got, matrixToolchain,
		)
	}
	if got != fixtureToolchain {
		t.Fatalf(
			"target %s: toolchain witness %q does not match fixtures/minimal-rect/expected.json's "+
				"recorded goToolchain %q (AD-22, counter-metric C6). This check runs before any hash "+
				"is compared and has no skip path.",
			target.name, got, fixtureToolchain,
		)
	}
}

// captureRender runs binPath with FOLIO_SUBPROCESS_RENDER=1 (the seam
// Story 1.1 already put in TestMain) and returns the rendered bytes.
func captureRender(t *testing.T, target matrixTarget, binPath string) []byte {
	t.Helper()
	return runOnTarget(t, target, binPath, map[string]string{subprocessEnvVar: "1"})
}

// checkFixtureShape, loadExpectedFixture and TestFixtureShapeCheckRedProof
// (DW-1's AC6/RP-4 closure) moved to the untagged fixture_test.go (Blocker
// 3, this story's QA review): this file carries the "matrix" build tag, so
// they previously executed in zero gates. loadExpectedFixture is still
// called below, from the same package.

// renderLegResult is one target's captured, hashed render output.
type renderLegResult struct {
	target matrixTarget
	hash   string
	bytes  []byte
}

func renderTableRows(results []renderLegResult) []string {
	rows := make([]string, len(results))
	for i, r := range results {
		rows[i] = fmt.Sprintf("%-14s %s  %d bytes", r.target.name, r.hash, len(r.bytes))
	}
	return rows
}

// TestCrossTargetByteIdentity is AC1's single local entry point and AC7's
// pairwise-plus-golden check, run all in one process. For every target it
// builds the render/toolchain test binary, runs AC3's toolchain gate
// before anything else, captures the render, validates it structurally
// (AC6, on every leg's output — not one "reference" render), and hashes
// it. It then implements AC17's two-case divergence report: all four
// agreeing with each other but not the golden is a legitimate versioned
// change under AD-22 to be investigated and re-recorded by a human; any
// pairwise disagreement is NFR1 falsified. It never re-records the
// fixture, majority-votes, or auto-selects the host's value (D-1.2.2).
func TestCrossTargetByteIdentity(t *testing.T) {
	assertExactlyFourMatrixTargets(t)

	root := repoRootFromTest(t)
	buildDir := matrixBuildDir(t, root)
	fixture := loadExpectedFixture(t, filepath.Join(root, "fixtures", "minimal-rect", "expected.json"))

	var results []renderLegResult
	toolchainLegs := 0
	for _, target := range matrixTargets {
		binPath := buildRenderTestBinary(t, root, target, buildDir)

		// AC3: toolchain gate runs before any hash is compared, no skip.
		assertToolchainWitness(t, target, binPath, fixture.GoToolchain)
		toolchainLegs++

		raw := captureRender(t, target, binPath)
		assertWellFormedPDF(t, target.name, raw) // AC6, every leg's own output

		sum := sha256.Sum256(raw)
		hash := hex.EncodeToString(sum[:])
		results = append(results, renderLegResult{target: target, hash: hash, bytes: raw})
	}

	// AC11: assert exactly four render legs AND exactly four toolchain
	// legs, each against the literal wantMatrixLegs constant, never
	// against len(matrixTargets) — that comparison is the loop's own
	// output against its own input and cannot fail (Blocker 1). The
	// toolchain-leg count is tracked independently rather than inferred
	// from the render loop, per the same finding: there was previously
	// no toolchain-leg count at all.
	if toolchainLegs != wantMatrixLegs {
		t.Fatalf("expected exactly %d toolchain legs, got %d (AC11)", wantMatrixLegs, toolchainLegs)
	}
	if len(results) != wantMatrixLegs {
		t.Fatalf("expected exactly %d render legs, got %d (AC11)", wantMatrixLegs, len(results))
	}

	table := strings.Join(renderTableRows(results), "\n")
	t.Logf("cross-target render matrix:\n%s", table)

	allAgreeWithEachOther := true
	for _, r := range results[1:] {
		if r.hash != results[0].hash {
			allAgreeWithEachOther = false
			break
		}
	}
	allMatchGolden := true
	for _, r := range results {
		if r.hash != fixture.SHA256 {
			allMatchGolden = false
			break
		}
	}

	if allAgreeWithEachOther && allMatchGolden {
		return
	}

	diagDir := filepath.Join(buildDir, "diverged")
	if err := os.MkdirAll(diagDir, 0o755); err != nil {
		t.Fatalf("create diagnostic dir: %v", err)
	}
	for _, r := range results {
		outPath := filepath.Join(diagDir, "diverged."+r.target.goos+"-"+r.target.goarch+".pdf")
		if err := os.WriteFile(outPath, r.bytes, 0o644); err != nil {
			t.Fatalf("write diagnostic output for %s: %v", r.target.name, err)
		}
	}

	if allAgreeWithEachOther && !allMatchGolden {
		// Case 1: a legitimate versioned change under AD-22 — never
		// auto-fixed by this harness.
		t.Fatalf(
			"all four targets agree with EACH OTHER but NOT with the recorded golden "+
				"(fixtures/minimal-rect/expected.json sha256=%s):\n%s\n\n"+
				"This is a legitimate versioned change under AD-22 (e.g. a Go toolchain bump, or a "+
				"deliberate, documented rendering-format change) to be investigated and "+
				"deliberately re-recorded by a human — never auto-fixed by this harness. The "+
				"harness never re-records the fixture, majority-votes, or auto-selects the host's "+
				"value (D-1.2.2). Diverging bytes written under %s.",
			fixture.SHA256, table, diagDir,
		)
	}

	// Case 2: targets disagree with each other — NFR1 falsified, the
	// product's core claim broken. Name every diverging pair.
	ref := results[0]
	var diagLines []string
	for _, r := range results[1:] {
		if r.hash != ref.hash {
			offset, window := firstDivergence(ref.bytes, r.bytes)
			diagLines = append(diagLines, fmt.Sprintf(
				"%s (%s) vs %s (%s): first divergence at byte offset %d; %s",
				ref.target.name, ref.hash, r.target.name, r.hash, offset, window,
			))
		}
	}
	t.Fatalf(
		"cross-target byte identity FAILED — targets disagree with EACH OTHER, not merely with "+
			"the golden. This falsifies NFR1, the product's core claim:\n%s\n\n%s\n\n"+
			"Diverging bytes written under %s. This harness never re-records the fixture, "+
			"majority-votes three targets against one, or auto-selects the host's value (D-1.2.2). "+
			"A real cross-target divergence is an owner decision, not one this harness or a "+
			"developer resolves by choosing a value.",
		table, strings.Join(diagLines, "\n"), diagDir,
	)
}

// TestTargetRenderHash is the single-target mode CI's per-target jobs
// call (AC1, AC13): it exercises the same code path as
// TestCrossTargetByteIdentity's per-leg body — toolchain gate,
// structural validation, golden comparison — for exactly the one target
// named by FOLIO_MATRIX_TARGET, and writes its hash to a file the
// workflow can publish as a build artifact. It contains no hash logic of
// its own beyond what TestCrossTargetByteIdentity already has; the
// workflow's compare job re-derives the pairwise property from the
// published hash files rather than reimplementing hashing in YAML.
//
// When FOLIO_MATRIX_TARGET is unset this test intentionally does nothing
// (it is CI's opt-in entry point, not one of the four counted legs
// AC11 governs) and passes trivially — TestCrossTargetByteIdentity above
// is what the D-000.4 local gate runs, and it is not gated behind this
// variable.
func TestTargetRenderHash(t *testing.T) {
	targetName := os.Getenv("FOLIO_MATRIX_TARGET")
	if targetName == "" {
		t.Log("FOLIO_MATRIX_TARGET not set: this test asserts NOTHING and is a deliberate no-op — " +
			"it is CI's single-target entry point (AC1, AC13), not one of the four counted legs AC11 " +
			"governs. Set e.g. FOLIO_MATRIX_TARGET=darwin/arm64 to exercise it locally. " +
			"TestCrossTargetByteIdentity exercises all four targets from one process and is what " +
			"the D-000.4 local gate runs. CI itself never reaches this no-op path, because every " +
			"render-*/fma-probe-* job in matrix.yml sets FOLIO_MATRIX_TARGET explicitly, and the " +
			"compare jobs additionally fail on a missing or malformed published artifact rather than " +
			"trusting that every per-target job ran (Blocker 2).")
		return
	}
	target, ok := findMatrixTarget(targetName)
	if !ok {
		t.Fatalf("FOLIO_MATRIX_TARGET=%q is not one of the four supported targets: %s", targetName, matrixTargetNames())
	}

	root := repoRootFromTest(t)
	buildDir := matrixBuildDir(t, root)
	fixture := loadExpectedFixture(t, filepath.Join(root, "fixtures", "minimal-rect", "expected.json"))

	binPath := buildRenderTestBinary(t, root, target, buildDir)
	assertToolchainWitness(t, target, binPath, fixture.GoToolchain)

	raw := captureRender(t, target, binPath)
	assertWellFormedPDF(t, target.name, raw)

	sum := sha256.Sum256(raw)
	hash := hex.EncodeToString(sum[:])

	hashOutPath := filepath.Join(buildDir, "hash."+target.goos+"-"+target.goarch+".txt")
	if err := os.WriteFile(hashOutPath, []byte(hash+"\n"), 0o644); err != nil {
		t.Fatalf("write hash output for %s: %v", target.name, err)
	}
	t.Logf("target %s: sha256=%s (%d bytes; written to %s)", target.name, hash, len(raw), hashOutPath)

	if hash != fixture.SHA256 {
		t.Fatalf("target %s: sha256 %s does not match recorded golden %s (fixtures/minimal-rect/expected.json)",
			target.name, hash, fixture.SHA256)
	}
}

// TestTargetProbeHex is CI's per-target entry point for the FMA probe,
// required independently by AC13 ("a job … asserting AC9's partition" runs
// on native per-target runners, same as the render legs) — unlike this
// file's local run, which uses one machine's Docker/Node to reach all four
// targets, TestFMAProbeDiverges's all-four-in-one-process design cannot run
// as a single CI job: there is no runner that is simultaneously native
// darwin/arm64 and has a Linux Docker daemon. (F-6 is widely documented to
// say GitHub-hosted macOS runners provide no Docker daemon, but that claim
// was explicitly NOT measured in this run and is NOT load-bearing here —
// D-1.2.5 — so it is cited only as background, never as the reason this
// test exists; AC13 alone is that reason.) Each of AC13's four native
// per-target runners instead calls this test for its own target, twice
// (AC10), and writes the hex output to a file a separate CI job reads to
// assert AC9's relation — mirroring TestTargetRenderHash's
// per-target/compare-job split. It contains no hash or relation logic
// beyond what TestFMAProbeDiverges already has.
//
// When FOLIO_MATRIX_TARGET is unset this test intentionally does nothing
// and passes trivially, exactly like TestTargetRenderHash.
func TestTargetProbeHex(t *testing.T) {
	targetName := os.Getenv("FOLIO_MATRIX_TARGET")
	if targetName == "" {
		t.Log("FOLIO_MATRIX_TARGET not set: this test asserts NOTHING and is a deliberate no-op — " +
			"it is CI's single-target FMA-probe entry point (AC13), not one of the four counted " +
			"probe legs AC11 governs. Set e.g. FOLIO_MATRIX_TARGET=darwin/arm64 to exercise it " +
			"locally. TestFMAProbeDiverges exercises all four targets from one process and is what " +
			"the D-000.4 local gate runs. CI itself never reaches this no-op path, because every " +
			"fma-probe-* job in matrix.yml sets FOLIO_MATRIX_TARGET explicitly.")
		return
	}
	target, ok := findMatrixTarget(targetName)
	if !ok {
		t.Fatalf("FOLIO_MATRIX_TARGET=%q is not one of the four supported targets: %s", targetName, matrixTargetNames())
	}

	root := repoRootFromTest(t)
	buildDir := matrixBuildDir(t, root)

	args := []string{"0.1", "0.2", "-0.02"}
	binPath := buildProbeBinary(t, root, target, buildDir)

	run1 := runOnTarget(t, target, binPath, nil, args...)
	run2 := runOnTarget(t, target, binPath, nil, args...)
	if !bytes.Equal(run1, run2) {
		t.Fatalf("target %s: probe is nondeterministic WITHIN the target — two runs with identical "+
			"operands %v produced %x and %x (AC10)", target.name, args, run1, run2)
	}
	if len(run1) != 8 {
		t.Fatalf("target %s: probe output is %d bytes, want 8 raw big-endian bytes from "+
			"math.Float64bits — never a formatted decimal (AC8, vacuity guard 10)",
			target.name, len(run1))
	}

	hexOut := hex.EncodeToString(run1)
	outPath := filepath.Join(buildDir, "probe."+target.goos+"-"+target.goarch+".txt")
	if err := os.WriteFile(outPath, []byte(hexOut+"\n"), 0o644); err != nil {
		t.Fatalf("write probe output for %s: %v", target.name, err)
	}
	t.Logf("target %s: probe output=%s (written to %s)", target.name, hexOut, outPath)
}

// TestFMAProbeDiverges is AC8–AC10, the negative test. It builds and runs
// hashmatrix/probe on all four targets, twice each with identical
// operands (AC10: the divergence in AC9 cannot be nondeterminism wearing
// a disguise), and asserts the divergence RELATION of AC9 — never a
// recorded probe value (D-1.2.3): the four outputs must not all be
// equal, and specifically darwin/arm64 and linux/arm64 must both differ
// from js/wasm. If the probe ever stops diverging, the build fails: a
// silently converged probe is a dead detector behind a green dashboard,
// which is the failure mode this test exists to prevent.
func TestFMAProbeDiverges(t *testing.T) {
	assertExactlyFourMatrixTargets(t)

	root := repoRootFromTest(t)
	buildDir := matrixBuildDir(t, root)

	// The operands are the probe's own os.Args inputs, not compile-time
	// constants (F-8, vacuity guard 9): literal operands are
	// constant-folded by the compiler and every target agrees on the
	// unfused value, making the probe silently vacuous.
	args := []string{"0.1", "0.2", "-0.02"}

	type probeResult struct {
		target matrixTarget
		hexOut string
	}
	var results []probeResult

	for _, target := range matrixTargets {
		binPath := buildProbeBinary(t, root, target, buildDir)

		run1 := runOnTarget(t, target, binPath, nil, args...)
		run2 := runOnTarget(t, target, binPath, nil, args...)
		if !bytes.Equal(run1, run2) {
			t.Fatalf(
				"target %s: probe is nondeterministic WITHIN the target — two runs with identical "+
					"operands %v produced %x and %x (AC10: the divergence across targets cannot be "+
					"trusted if a target cannot even agree with itself)",
				target.name, args, run1, run2,
			)
		}
		if len(run1) != 8 {
			t.Fatalf(
				"target %s: probe output is %d bytes, want 8 raw big-endian bytes from "+
					"math.Float64bits — never a formatted decimal (AC8, vacuity guard 10)",
				target.name, len(run1),
			)
		}
		results = append(results, probeResult{target: target, hexOut: hex.EncodeToString(run1)})
	}

	// AC11: assert exactly four probe legs, against the literal
	// wantMatrixLegs constant — see wantMatrixLegs's comment (Blocker 1).
	if len(results) != wantMatrixLegs {
		t.Fatalf("expected exactly %d probe legs, got %d (AC11)", wantMatrixLegs, len(results))
	}

	rows := make([]string, len(results))
	outByName := map[string]string{}
	for i, r := range results {
		rows[i] = fmt.Sprintf("%-14s %s", r.target.name, r.hexOut)
		outByName[r.target.name] = r.hexOut
	}
	table := strings.Join(rows, "\n")
	t.Logf("FMA contraction probe outputs (operands %v):\n%s", args, table)

	allEqual := true
	for _, r := range results[1:] {
		if r.hexOut != results[0].hexOut {
			allEqual = false
			break
		}
	}

	// The relation AC9 requires is stronger than "not all equal": both
	// arm64 targets must differ from wasm specifically. A GOOS-keyed or
	// GOARCH-keyed probe (runtime.GOOS, a darwin/linux split) could
	// satisfy "not all equal" while putting linux/arm64 on the same side
	// as linux/amd64 — this is exactly what RP-6 exercises, and exactly
	// what guard 5 forbids: a probe detecting its platform rather than
	// contraction.
	//
	// The three lookups below use the comma-ok form deliberately (Finding
	// 3): outByName["js/wasm"] on a Go map returns the zero value "" for
	// a MISSING key, and "" happens to differ from every real probe
	// output — so a target that never ran would read as "diverges from
	// wasm" and silently satisfy this relation. This is a distinct
	// mechanism from the counted-four checks above (Blocker 1): even with
	// exactly four results in hand, a renamed or mismatched target name
	// would still hit this gap, which is why it needs its own guard
	// rather than relying on the leg count.
	wasmOut, wasmOK := outByName["js/wasm"]
	darwinOut, darwinOK := outByName["darwin/arm64"]
	linuxArmOut, linuxArmOK := outByName["linux/arm64"]
	if !wasmOK || !darwinOK || !linuxArmOK {
		t.Fatalf(
			"expected probe output for darwin/arm64, linux/arm64 and js/wasm, but at least one is "+
				"missing (darwin/arm64 present=%v, linux/arm64 present=%v, js/wasm present=%v) — a "+
				"missing leg must not be read as a satisfied divergence (AC9, Finding 3):\n%s",
			darwinOK, linuxArmOK, wasmOK, table,
		)
	}
	bothArm64DifferFromWasm := darwinOut != wasmOut && linuxArmOut != wasmOut

	// This relation is mirrored in .github/workflows/matrix.yml's
	// assert-fma-probe-diverges job (both_arm64_differ_from_wasm), which
	// re-derives the same check in shell over the four published probe
	// files rather than calling back into this test binary (AC1: the
	// render-side compare job contains no hash logic of its own; the
	// probe-side job is real logic per AC13's own mandate). The two
	// implementations are NOT pinned together mechanically — if this
	// relation is ever strengthened or changed, update matrix.yml's
	// mirror in the same change, or CI would keep asserting a weaker
	// relation than this file does (Finding 8).
	if allEqual || !bothArm64DifferFromWasm {
		t.Fatalf(
			"the contraction probe no longer diverges across targets; the matrix's ability to "+
				"detect FMA contraction is unproven — investigate before trusting a green render "+
				"matrix (AC9):\n%s",
			table,
		)
	}
}
