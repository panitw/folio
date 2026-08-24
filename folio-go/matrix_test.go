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
	"regexp"
	"runtime"
	"strconv"
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

// fatalReporter is the minimal *testing.T surface assertFixturesShareToolchain
// needs. *testing.T satisfies it directly; TestAssertFixturesShareToolchainRedProof's
// red-proof below uses a non-testing.T fake instead of a real subtest,
// because a genuinely-failing t.Run subtest always marks its PARENT test
// failed too (a stdlib testing property, not something a subtest's own
// assertions can suppress) — which would make "prove this fails under a
// mismatch" indistinguishable from "this test itself is broken".
type fatalReporter interface {
	Helper()
	Fatalf(format string, args ...any)
}

// assertFixturesShareToolchain is Finding 15's fix (QA review): before
// this fix, assertToolchainWitness was called ONLY against
// matrixDocuments[0]'s fixture (minimal-rect) — fixtures/font-text/'s
// own recorded goToolchain was loaded but never compared against
// anything. Both currently record go1.26.0, so nothing diverged in
// practice, but the font fixture's toolchain field was decorative: if
// font-text were ever re-recorded under a different toolchain, the
// matrix would silently accept a mismatched provenance claim for one of
// the two documents — the exact AD-22/C6 property the toolchain gate
// exists to enforce, unenforced for font-text. This asserts every
// document's fixture records the SAME goToolchain, so a divergence
// between the two fixtures' recorded toolchains fails loudly, before
// assertToolchainWitness runs against a single one of them.
func assertFixturesShareToolchain(t fatalReporter, fixtures map[string]expectedFixture) {
	t.Helper()
	if len(matrixDocuments) == 0 {
		return
	}
	ref := fixtures[matrixDocuments[0].label]
	for _, doc := range matrixDocuments[1:] {
		f := fixtures[doc.label]
		if f.GoToolchain != ref.GoToolchain {
			t.Fatalf(
				"fixture toolchain mismatch: %q records goToolchain=%q but %q records goToolchain=%q "+
					"(AD-22, counter-metric C6, Finding 15) — every document's fixture must record the "+
					"same toolchain, or the matrix's single-fixture toolchain gate silently stops "+
					"covering the others",
				matrixDocuments[0].label, ref.GoToolchain, doc.label, f.GoToolchain,
			)
		}
	}
}

// fakeFatalReporter is a fatalReporter that RECORDS a Fatalf call
// instead of failing the current test — see fatalReporter's doc comment
// for why a real t.Run subtest cannot be used here.
type fakeFatalReporter struct {
	called bool
	msg    string
}

func (f *fakeFatalReporter) Helper() {}
func (f *fakeFatalReporter) Fatalf(format string, args ...any) {
	f.called = true
	f.msg = fmt.Sprintf(format, args...)
}

// TestAssertFixturesShareToolchainRedProof red-proofs Finding 15's fix
// directly, without needing a full Docker/Node matrix run: it swaps
// matrixDocuments for a two-entry scratch list (restored via defer) and
// confirms assertFixturesShareToolchain does NOT report a failure for
// matching toolchains and DOES for mismatched ones.
func TestAssertFixturesShareToolchainRedProof(t *testing.T) {
	orig := matrixDocuments
	matrixDocuments = []matrixDocument{{label: "a"}, {label: "b"}}
	defer func() { matrixDocuments = orig }()

	matching := map[string]expectedFixture{
		"a": {GoToolchain: "go1.26.0"},
		"b": {GoToolchain: "go1.26.0"},
	}
	rep := &fakeFatalReporter{}
	assertFixturesShareToolchain(rep, matching)
	if rep.called {
		t.Fatalf("matching-toolchain fixtures must not fail assertFixturesShareToolchain, got: %s", rep.msg)
	}

	mismatched := map[string]expectedFixture{
		"a": {GoToolchain: "go1.26.0"},
		"b": {GoToolchain: "go1.25.0"},
	}
	rep = &fakeFatalReporter{}
	assertFixturesShareToolchain(rep, mismatched)
	if !rep.called {
		t.Fatal("mismatched-toolchain fixtures should have failed assertFixturesShareToolchain (Finding 15 red-proof)")
	}
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
// Story 1.1 already put in TestMain) and returns the rendered bytes —
// the FONTLESS document (AC10a: this selector keeps rendering
// fixtures/minimal-rect/'s document, unchanged).
func captureRender(t *testing.T, target matrixTarget, binPath string) []byte {
	t.Helper()
	return runOnTarget(t, target, binPath, map[string]string{subprocessEnvVar: "1"})
}

// captureFontRender runs binPath with FOLIO_SUBPROCESS_RENDER_FONT=1 —
// AC10a's SECOND selector, rendering the template+font document
// (fixtures/font-text/) through the public Render path.
func captureFontRender(t *testing.T, target matrixTarget, binPath string) []byte {
	t.Helper()
	return runOnTarget(t, target, binPath, map[string]string{subprocessFontEnvVar: "1"})
}

// captureImageRender runs binPath with FOLIO_SUBPROCESS_RENDER_IMAGE=1 —
// AC22's THIRD selector (D-1.8.6, joining the two above, replacing
// neither): rendering the image-embedding document
// (fixtures/image-embed/) through the public Render path.
func captureImageRender(t *testing.T, target matrixTarget, binPath string) []byte {
	t.Helper()
	return runOnTarget(t, target, binPath, map[string]string{subprocessImageEnvVar: "1"})
}

// captureMultiScriptRender runs binPath with
// FOLIO_SUBPROCESS_RENDER_MULTISCRIPT=1 — Story 2.2's AC8 FOURTH
// selector (D-1.8.6, joining the three above, replacing none): rendering
// the multi-script fallback-chain document (fixtures/multi-script-
// fallback/) through the public Render path, against the real shipped
// face set.
func captureMultiScriptRender(t *testing.T, target matrixTarget, binPath string) []byte {
	t.Helper()
	return runOnTarget(t, target, binPath, map[string]string{subprocessMultiScriptEnvVar: "1"})
}

// captureShapedTextRender runs binPath with
// FOLIO_SUBPROCESS_RENDER_SHAPEDTEXT=1 — Story 2.3's AC10 FIFTH selector
// (joining the four above, replacing none): rendering the shaped-text
// document (fixtures/shaped-text/) through the public Render path,
// against the real shipped face set.
func captureShapedTextRender(t *testing.T, target matrixTarget, binPath string) []byte {
	t.Helper()
	return runOnTarget(t, target, binPath, map[string]string{subprocessShapedTextEnvVar: "1"})
}

// captureWrappedTextRender runs binPath with
// FOLIO_SUBPROCESS_RENDER_WRAPPEDTEXT=1 — Story 2.4's SIXTH selector,
// rendering fixtures/wrapped-text/ through the public Render path.
func captureWrappedTextRender(t *testing.T, target matrixTarget, binPath string) []byte {
	t.Helper()
	return runOnTarget(t, target, binPath, map[string]string{subprocessWrappedTextEnvVar: "1"})
}

// captureThreeBandRender runs binPath with
// FOLIO_SUBPROCESS_RENDER_THREEBAND=1 — Story 2.5's SEVENTH selector,
// rendering fixtures/three-band-page/ through the public Render path.
func captureThreeBandRender(t *testing.T, target matrixTarget, binPath string) []byte {
	t.Helper()
	return runOnTarget(t, target, binPath, map[string]string{subprocessThreeBandEnvVar: "1"})
}

// captureMultiPageRender runs binPath with
// FOLIO_SUBPROCESS_RENDER_MULTIPAGE=1 — Story 2.6's EIGHTH selector,
// rendering fixtures/multi-page/ through the public Render path.
func captureMultiPageRender(t *testing.T, target matrixTarget, binPath string) []byte {
	t.Helper()
	return runOnTarget(t, target, binPath, map[string]string{subprocessMultiPageEnvVar: "1"})
}

// matrixKidsArrayRE and matrixWellFormedRefRE are this file's OWN copies of
// golden_structural_validity_test.go's kidsArrayRE / wellFormedRefRE, not a
// shared import: that file is package folio_test (external), this one is
// package folio (internal, so it can call readEmittedRuns et al. directly,
// AD-21) — the two packages cannot share unexported identifiers. The
// duplication is a regex LITERAL, not a stateful walk (contrast Finding 8's
// nonStreamRegions/assertStreamLengthsAreExact coupling, which this is NOT
// an instance of), and the two callers check genuinely different subjects:
// golden_structural_validity_test.go resolves the COMMITTED golden on disk;
// requirePageTreeResolves below resolves the FRESH per-target render, before
// any byte comparison — D-2.6.8's layer/subject distinction applied to this
// guard (Story 2.6 finisher, Blocker 1 / Finding 2).
var matrixKidsArrayRE = regexp.MustCompile(`/Type /Pages /Kids \[([^\]]*)\] /Count (\d+)`)
var matrixWellFormedRefRE = regexp.MustCompile(`^(\d+) 0 R$`)

// requirePageTreeResolves is Story 2.6's finisher fix for Blocker 1 / Major
// 2: the ORIGINAL two checks it replaces — bytes.Count("<< /Type /Page
// /Parent ") == wantPages, and bytes.Contains("] /Count %d ") — both measure
// PRESENCE, and both were satisfied by the exact broken bytes this story
// exists to prevent: "<< /Type /Pages /Kids [8 0 R10 0 R] /Count 2 >>" has
// two correctly-emitted /Type /Page objects and the literal substring
// "] /Count 2 ", while its page tree resolves to ZERO pages (D-000.53). A
// fix that merely counted to 2 would leave a guard that cannot tell a
// working document from the artifact that started this story.
//
// So this resolves the /Kids array into the SET of object numbers it
// actually references — not merely their COUNT, which
// golden_structural_validity_test.go's finisher fix (Finding 4) shows is not
// enough either: a duplicated reference plus an orphaned page has the right
// count and the wrong set — and requires that set to equal wantPages AND to
// equal the set of objects actually defined as "/Type /Page". Every
// well-formedness precondition below fires with a diagnostic BEFORE the
// weaker checks would have (D-000.21 sharpened).
func requirePageTreeResolves(t *testing.T, target matrixTarget, raw []byte, wantCount int) {
	t.Helper()

	m := matrixKidsArrayRE.FindSubmatch(raw)
	if m == nil {
		t.Fatalf("%s: no page tree of the form \"/Type /Pages /Kids [...] /Count N\" found", target.name)
	}
	declaredCount, err := strconv.Atoi(string(m[2]))
	if err != nil || declaredCount != wantCount {
		t.Fatalf("%s: page tree declares /Count %q, want %d", target.name, m[2], wantCount)
	}

	fields := bytes.Fields(m[1])
	if len(fields)%3 != 0 {
		t.Fatalf("%s: /Kids array %q does not tokenize into whole \"N 0 R\" triples — a run-together "+
			"reference (e.g. \"8 0 R10 0 R\") collapses two tokens into one unparseable token, which is "+
			"exactly how a two-page document shipped with an EMPTY page tree behind a correct-looking "+
			"/Count (D-000.53)", target.name, m[1])
	}
	refSet := map[int]bool{}
	for i := 0; i+2 < len(fields); i += 3 {
		ref := string(fields[i]) + " " + string(fields[i+1]) + " " + string(fields[i+2])
		rm := matrixWellFormedRefRE.FindStringSubmatch(ref)
		if rm == nil {
			t.Fatalf("%s: /Kids array contains %q, which is not a well-formed indirect reference (\"N 0 R\")", target.name, ref)
		}
		num, _ := strconv.Atoi(rm[1])
		refSet[num] = true
	}
	if len(refSet) != declaredCount {
		t.Fatalf("%s: /Kids array references %d DISTINCT object(s) but /Count says %d — a duplicated "+
			"reference plus an orphaned page has the right count and the wrong set, and a count comparison "+
			"cannot see it (Finding 4)", target.name, len(refSet), declaredCount)
	}

	objs := pdfObjects(t, raw)
	definedSet := map[int]bool{}
	for num, body := range objs {
		if bytes.HasPrefix(body, []byte("<< /Type /Page /Parent ")) {
			definedSet[num] = true
		}
	}
	for num := range refSet {
		if !definedSet[num] {
			t.Fatalf("%s: /Kids references object %d, which is never defined as a /Type /Page object — a "+
				"syntactically valid reference to a non-existent or wrong-typed object is still an empty "+
				"page tree", target.name, num)
		}
	}
	if len(definedSet) != len(refSet) {
		t.Fatalf("%s: the file defines %d /Type /Page object(s) but the page tree references %d distinct "+
			"one(s) — a page object that exists and is unreachable is invisible to a reader and to a hash "+
			"alike", target.name, len(definedSet), len(refSet))
	}
}

// requireMultiPageIsGenuinelyMultiPage is Story 2.6's OWN feature guard, and
// it exists for the same reason requireShapedTextIsShaped,
// requireWrappedTextIsWrapped and requireThreeBandPageUsesAllThreeBands do:
// "contains a FontFile2" is satisfied by any embedding, and a document that
// FAILED TO PAGINATE would still embed a face and still agree across four
// targets — four legs agreeing on the WRONG answer is the failure mode a
// cross-target hash comparison cannot see by itself.
//
// It asserts, on EVERY leg and before any byte comparison, the four things
// that make this document the pagination fixture rather than just another
// golden:
//
//  1. Its page tree RESOLVES — every /Kids reference is a distinct, defined
//     /Type /Page object, matching the declared /Count exactly
//     (requirePageTreeResolves, added by the finisher: the ORIGINAL two
//     checks here — a page-object count and a "] /Count N " substring —
//     both measured PRESENCE and both passed on the exact broken bytes this
//     story exists to prevent; see Blocker 1 / Finding 2).
//  2. NO text is placed at a negative PDF Y. That is the defect itself:
//     before this story the same document drew its last seven lines below
//     the bottom edge of the sheet, silently.
//  3. The running header and footer appear on BOTH pages, so a leg that
//     paginated the content but dropped the running bands from page 2 —
//     which is what "page 34 is as complete as page 1" forbids — fails.
//     This reads EVERY text-bearing content stream via readEmittedRuns
//     (shaped_fixture_test.go), which Story 2.6a repaired to walk every
//     text-bearing stream in deterministic ascending-object-number order.
//     Before that repair this call used a narrow sibling,
//     readEmittedRunsAllPages, added here rather than by changing
//     readEmittedRuns' contract for its OTHER call sites — 11 of them at
//     Story 2.6a's own start-of-work baseline (Story 2.6 finisher,
//     Blocker 1) — Story 2.6a's charter was to sweep those 11 sites and,
//     having found none of them depended on single-stream reading (Story
//     2.6a, AC6), merged the sibling's behaviour directly into
//     readEmittedRuns and retired the sibling. THIS call site (line 692)
//     is a TWELFTH, multi-stream site the merge itself created; Story
//     2.6a's review (Finding 5) caught that AC6's classification table
//     never covered it, and the finisher's correction adds it explicitly:
//     it is one of the 3 call sites in the finishing-commit tree that
//     genuinely depend on every-stream reading (the other 2 are new tests
//     shaped_fixture_test.go added closing Finding 1 / Blocker 1), against
//     11 single-stream sites elsewhere — see readEmittedRuns' own docblock
//     for the corrected count.
func requireMultiPageIsGenuinelyMultiPage(t *testing.T, target matrixTarget, raw []byte) {
	t.Helper()

	if len(raw) == 0 {
		t.Fatalf("%s: the multi-page leg produced no bytes", target.name)
	}
	requirePageTreeResolves(t, target, raw, 2)

	runs := readEmittedRuns(t, raw)
	if len(runs) == 0 {
		t.Fatalf("%s: the multi-page leg emitted no text runs", target.name)
	}
	for _, r := range runs {
		if r.OriginYMilli < 0 {
			t.Fatalf("%s: the multi-page leg places text at a NEGATIVE PDF Y (%d) — text drawn below the bottom edge of the sheet, which is precisely the defect Story 2.6 fixed", target.name, r.OriginYMilli)
		}
	}

	// The running bands, one occurrence per page. Their PDF Y is identical on
	// every page by construction (AC3), so counting occurrences of each Y is
	// exactly "does it appear on both pages".
	headerRuns, footerRuns := 0, 0
	for _, r := range runs {
		switch r.OriginYMilli {
		case 798269:
			headerRuns++
		case 51448:
			footerRuns++
		}
	}
	if headerRuns != 2 || footerRuns != 2 {
		t.Fatalf("%s: the multi-page leg carries %d page-header run(s) and %d page-footer run(s) at their declared Y; a two-page document must carry exactly 2 of each — one per page",
			target.name, headerRuns, footerRuns)
	}
}

// NOTE (Story 2.6a, tombstone — deliberately not a doc comment, per the
// story's own review Finding 10: a future reader must not mistake this
// for documenting a live declaration below): readEmittedRunsAllPages used
// to be this file's multi-page-aware sibling of readEmittedRuns
// (shaped_fixture_test.go). Story 2.6's finisher added it narrowly for
// this one caller rather than changing readEmittedRuns' contract for its
// other call sites — 11 of them at Story 2.6a's own start-of-work
// baseline (Blocker 1 — scope guard), so that Story 2.6a — chartered to
// sweep those 11 sites — could measure what it fixed rather than
// inheriting an already-changed function.
//
// Story 2.6a's sweep (AC4, AC5, AC6) found none of the 11 baseline sites
// depended on single-stream reading, merged this function's deterministic
// ascending-object-number, every-text-stream behaviour directly into
// readEmittedRuns, and retired this sibling — its one caller,
// requireMultiPageIsGenuinelyMultiPage above, now calls readEmittedRuns
// directly. That also satisfies AC5: a multi-page-aware reader is reachable
// from the ordinary, untagged suite, because readEmittedRuns always was.
// This function is deliberately deleted rather than kept as an unused
// alias — a retired helper visible to future callers reads as an option.
//
// Correction (Story 2.6a review, Finding 5): that merge made
// requireMultiPageIsGenuinelyMultiPage itself a TWELFTH call site, and
// AC6's classification table, written against the 11-site baseline,
// never covered it. It is corrected above, at this function's own former
// call site.

// requireThreeBandPageUsesAllThreeBands is Story 2.5's OWN feature
// guard, and it exists for the same reason requireShapedTextIsShaped and
// requireWrappedTextIsWrapped do: "contains a FontFile2" is satisfied by
// any embedding, and a document whose bands were composed WRONGLY would
// still embed a face and still agree across four targets.
//
// Measured at Story 2.5's creation, by injecting a wrong origin into
// each band in turn: a wrong PAGE-HEADER origin was caught by ZERO tests
// in the repository, because all six pre-2.5 fixtures have an EMPTY
// pageHeader band. This guard asserts, on EVERY leg and before any byte
// comparison, that the captured stream carries FOUR runs at FOUR
// DISTINCT baselines spanning all three bands — read off the PRODUCED
// bytes, never off the renderer's intermediate state.
//
// The three band boundaries are stated as literals derived by hand from
// the fixture's page setup (see three_band_page_fixture_test.go for the
// arithmetic): the printable column runs from PDF-user-space y 811890
// down to y 42000, the header band ends 18000 mp below the top, and the
// footer band begins 24000 mp above the bottom.
func requireThreeBandPageUsesAllThreeBands(t *testing.T, target matrixTarget, raw []byte) {
	t.Helper()

	runs := readEmittedRuns(t, raw)
	if len(runs) == 0 {
		t.Fatalf("%s: the three-band-page leg emitted no text runs", target.name)
	}
	ys := linesByOrigin(runs)
	if len(ys) != 4 {
		t.Fatalf("%s: the three-band-page leg occupies %d distinct baselines %v, want exactly 4 (one per text element: one header, two content, one footer) — a leg that collapsed two bands would compare byte-identical to itself and certify nothing",
			target.name, len(ys), ys)
	}

	// PDF user space, bottom-up: printable top = 841890 - 30000 = 811890,
	// printable bottom = 42000. Header band: [811890-18000, 811890] =
	// [793890, 811890]. Footer band: [42000, 42000+24000] = [42000, 66000].
	var header, content, footer int
	for _, y := range ys {
		switch {
		case y >= 793890:
			header++
		case y <= 66000:
			footer++
		default:
			content++
		}
	}
	if header == 0 || content == 0 || footer == 0 {
		t.Fatalf("%s: the three-band-page leg's baselines %v put %d run(s) in the page-header band, %d in the content band and %d in the page-footer band — every band must carry at least one, or a band-origin defect is invisible on this target",
			target.name, ys, header, content, footer)
	}
}

// requireWrappedTextIsWrapped is Story 2.4's OWN feature guard, and it
// exists for the same reason requireShapedTextIsShaped does.
//
// "Contains a FontFile2" is satisfied by any embedding, and every
// pre-2.4 fixture's text FITS its box — so a matrix comparing four legs
// of a document that never wrapped would agree byte for byte and certify
// nothing about line breaking. This guard asserts on EVERY leg, before
// any byte comparison, that the captured stream actually carries wrapped
// text: MORE DISTINCT BASELINES THAN TEXT ELEMENTS. The fixture has four
// text elements and is measured to occupy nine baselines; a build that
// stopped wrapping would emit four, and this fires.
//
// It is a property of the PRODUCED bytes, read back off them.
func requireWrappedTextIsWrapped(t *testing.T, target matrixTarget, raw []byte) {
	t.Helper()

	const wantElements = 4
	runs := readEmittedRuns(t, raw)
	if len(runs) == 0 {
		t.Fatalf("%s: the wrapped-text leg emitted no text runs", target.name)
	}
	ys := linesByOrigin(runs)
	if len(ys) <= wantElements {
		t.Fatalf(
			"%s: the wrapped-text leg occupies %d distinct baselines %v for %d text elements — nothing "+
				"wrapped on this target, so comparing its bytes would certify nothing about line breaking",
			target.name, len(ys), ys, wantElements,
		)
	}
}

// requireShapedTextIsShaped is Story 2.3's OWN feature guard for the
// shaped-text document, and it is the reason registering the legs is not
// a formality.
//
// "Contains a FontFile2" is satisfied by any embedding at all, and every
// pre-2.3 fixture's text was measured to shape to itself — so a matrix
// comparing four legs of a document that was never shaped would agree
// perfectly, byte for byte, and certify nothing about this story. This
// guard asserts on EVERY leg, before any byte comparison, that the
// captured stream actually carries shaping: at least one TJ array with a
// non-zero adjustment (GPOS reached the page) AND at least one CID whose
// /ToUnicode entry is longer than one rune (GSUB collapsed a cluster).
// Both are properties of the PRODUCED bytes, read back off them.
func requireShapedTextIsShaped(t *testing.T, target matrixTarget, raw []byte) {
	t.Helper()

	programs := extractAllFontFile2Programs(t, raw)
	if len(programs) != len(shippedFaceSpecs) {
		t.Fatalf(
			"target %s: the shaped-text fixture must embed exactly %d FontFile2 programs, got %d",
			target.name, len(shippedFaceSpecs), len(programs),
		)
	}
	assertEmbeddedProgramsAreStaticRegular(t, "target "+target.name, programs, 400)

	wantPS := map[string]bool{}
	for _, spec := range shippedFaceSpecs {
		wantPS[spec.PostScriptName] = true
	}
	assertBaseFontNames(t, "target "+target.name, raw, wantPS)

	emitted := readEmittedRuns(t, raw)
	adjusted := 0
	for _, run := range emitted {
		if run.NonZeroAdjust {
			adjusted++
		}
	}
	if adjusted == 0 {
		t.Fatalf(
			"target %s: the shaped-text fixture's content stream carries NO TJ adjustment. Four targets "+
				"agreeing on an unshaped document would be byte-identical and would certify nothing about "+
				"this story (D-000.9).",
			target.name,
		)
	}

	merged := 0
	for _, cmap := range toUnicodeForResources(t, raw) {
		for _, txt := range cmap {
			if len([]rune(txt)) > 1 {
				merged++
			}
		}
	}
	if merged == 0 {
		t.Fatalf(
			"target %s: no CID in the shaped-text fixture extracts as more than one rune, so no ligature or "+
				"merged cluster reached the page — GSUB is not being exercised on this leg",
			target.name,
		)
	}
}

// requireInstancedShippedFaces is AC8's OWN feature guard (V6): asserts
// that raw (a captured multi-script leg) contains exactly the three
// shipped faces' embedded programs, and that EACH ONE is genuinely
// INSTANCED — carries no fvar/gvar/avar table — never merely "contains
// a FontFile2" (which requireFontFile2 alone would accept even from an
// un-instanced, still-variable embedding, proving nothing about AC7's
// actual hazard).
func requireInstancedShippedFaces(t *testing.T, target matrixTarget, raw []byte) {
	t.Helper()
	programs := extractAllFontFile2Programs(t, raw)
	if len(programs) != len(shippedFaceSpecs) {
		t.Fatalf(
			"target %s: multi-script fixture must embed exactly %d FontFile2 programs (one per shipped face "+
				"actually used), got %d (AC8 vacuity guard)",
			target.name, len(shippedFaceSpecs), len(programs),
		)
	}

	// D-000.22's semantic acceptance step, run on EVERY target rather
	// than only where the golden was recorded. It asserts the tables
	// (static, glyf, no CFF2) AND OS/2.usWeightClass == 400, each behind
	// a presence precondition.
	//
	// The weight half is the one that matters here and it is new. The
	// previous guard checked only that `fvar`/`gvar`/`avar` were absent,
	// which the Thin CJK program satisfied perfectly: it WAS genuinely
	// instanced, to a default weight of 100. Four targets then agreed on
	// it, byte for byte, and every one of them was right about the
	// question it was asked (D-000.21).
	assertEmbeddedProgramsAreStaticRegular(t, "target "+target.name, programs, 400)

	// And the PDF-level name, which is what a reader or a validator sees:
	// one six-letter subset tag, one '+', then the embedded program's own
	// PostScript name (ISO 32000-1 §9.6.4 + Table 117).
	wantPS := map[string]bool{}
	for _, spec := range shippedFaceSpecs {
		wantPS[spec.PostScriptName] = true
	}
	assertBaseFontNames(t, "target "+target.name, raw, wantPS)
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

// matrixDocument names one of AC15's two documents this story's matrix
// runs: the fontless fixture (unchanged, AC10a) and the new template+font
// fixture (AC10a's second selector). requireFontFile2 is AC10b's binding
// vacuity guard, applied per document — only the font document must
// contain one; asserting it on the fontless document would be wrong, not
// merely redundant. slug is a filename-safe identifier (Finding 8, QA
// review) used by BOTH TestTargetRenderHash and .github/workflows/matrix.yml
// to name one hash artifact PER TARGET PER DOCUMENT — before this fix,
// CI's per-target jobs published only the fontless hash
// (TestTargetRenderHash captured only minimal-rect), so a textshape
// upgrade that broke font cross-target byte-identity would pass CI even
// though TestCrossTargetByteIdentity (the local, matrix-tagged gate)
// would catch it.
type matrixDocument struct {
	label               string
	slug                string
	capture             func(t *testing.T, target matrixTarget, binPath string) []byte
	fixtureRelPath      []string
	requireFontFile2    bool
	requireImageXObject bool // AC23 (D-1.8.6): applied per document, on the same basis as requireFontFile2
	// extraGuard is Story 2.2's (AC8) generalization: a document-specific
	// feature guard beyond requireFontFile2/requireImageXObject, run on
	// EVERY captured leg before any byte comparison — the same shape,
	// widened because "contains a FontFile2" is not sufficient to prove
	// AC7's instancing hazard was actually exercised (V6).
	extraGuard func(t *testing.T, target matrixTarget, raw []byte)

	// wantPages is how many /Type /Page objects this document's render must
	// carry, passed to assertWellFormedPDF on EVERY leg.
	//
	// It is a per-document field rather than a constant because Story 2.6
	// added the first document with more than one page — and because the
	// checker previously hard-coded 1, it could not be pointed at that
	// document at all. Zero reads as 1, so every pre-2.6 entry keeps its
	// existing meaning without being edited.
	wantPages int
}

// pageCount returns the document's declared page count, treating the zero
// value as 1 so the seven single-page entries need no edit.
func (d matrixDocument) pageCount() int {
	if d.wantPages == 0 {
		return 1
	}
	return d.wantPages
}

var matrixDocuments = []matrixDocument{
	{
		label:            "minimal-rect (fontless)",
		slug:             "minimal-rect",
		capture:          captureRender,
		fixtureRelPath:   []string{"fixtures", "minimal-rect", "expected.json"},
		requireFontFile2: false,
	},
	{
		label:            "font-text (template+font)",
		slug:             "font-text",
		capture:          captureFontRender,
		fixtureRelPath:   []string{"fixtures", "font-text", "expected.json"},
		requireFontFile2: true,
	},
	{
		label:               "image-embed (template+image)",
		slug:                "image-embed",
		capture:             captureImageRender,
		fixtureRelPath:      []string{"fixtures", "image-embed", "expected.json"},
		requireImageXObject: true,
	},
	{
		label:            "multi-script-fallback (shipped faces, fallback chain)",
		slug:             "multi-script-fallback",
		capture:          captureMultiScriptRender,
		fixtureRelPath:   []string{"fixtures", "multi-script-fallback", "expected.json"},
		requireFontFile2: true,
		extraGuard:       requireInstancedShippedFaces,
	},
	{
		// Story 2.3's AC10 fixture. REGISTERED IN-STORY, RUN AT THE
		// EPIC 2 BOUNDARY GATE: 2.3 is not one of D-000.4's per-story
		// matrix overrides (those are 1.2, 1.5, 1.8, 2.4 and 4.7), so
		// the Docker legs are deliberately not run here. Registering it
		// is still this story's obligation — AD-21 binds every feature
		// to ship its golden, and a fixture recorded but never added to
		// this list is one the matrix silently never covers.
		label:            "shaped-text (shaping, all three scripts)",
		slug:             "shaped-text",
		capture:          captureShapedTextRender,
		fixtureRelPath:   []string{"fixtures", "shaped-text", "expected.json"},
		requireFontFile2: true,
		extraGuard:       requireShapedTextIsShaped,
	},
	{
		// Story 2.4's AC15 fixture. REGISTERED AND RUN IN-STORY:
		// D-000.4 names 2.4 explicitly as a per-story matrix override
		// ("line breaking feeds every measurement"), alongside 1.2, 1.5,
		// 1.8 and 4.7. Unlike shaped-text above, this document's four
		// legs are executed by this story, not deferred to the Epic 2
		// gate.
		label:            "wrapped-text (line breaking, all three scripts)",
		slug:             "wrapped-text",
		capture:          captureWrappedTextRender,
		fixtureRelPath:   []string{"fixtures", "wrapped-text", "expected.json"},
		requireFontFile2: true,
		extraGuard:       requireWrappedTextIsWrapped,
	},
	{
		// Story 2.5's AC7 fixture: the only registered document with
		// content in ALL THREE bands and four pairwise-distinct
		// geometric inputs.
		//
		// REGISTERED HERE; ITS FOUR LEGS ARE NOT RUN BY THIS STORY.
		// D-000.4 (override criterion) DECLINED Story 2.5's per-story
		// override: "an override is warranted when a story introduces a
		// new SOURCE OF CROSS-TARGET DIVERGENCE — float arithmetic, a
		// vendor call, a compressor, a new dependency — not merely
		// because it records a new golden. Story 2.5 is integer band
		// arithmetic on geom.Length." The legs are the Epic 2 boundary
		// gate's, and registration now is what gives that gate something
		// to compare (D-2.5.1 sanctions this as the gate's fourth
		// obligation; a golden registered in one list and not the other
		// is "a matrix leg nobody compares, reported as green").
		label:            "three-band-page (band composition, all three bands populated)",
		slug:             "three-band-page",
		capture:          captureThreeBandRender,
		fixtureRelPath:   []string{"fixtures", "three-band-page", "expected.json"},
		requireFontFile2: true,
		extraGuard:       requireThreeBandPageUsesAllThreeBands,
	},
	{
		// Story 2.6. The legs are DEFERRED to the Epic 2 boundary gate, on
		// the same D-000.4 reasoning that deferred three-band-page's: the
		// override criterion is a NEW SOURCE OF CROSS-TARGET DIVERGENCE —
		// float arithmetic, a vendor call, a compressor, a new dependency —
		// not merely a new golden. Pagination is integer comparison and
		// subtraction on geom.Length; it calls no shaper, adds no
		// dependency, and go list -m all stays at exactly two modules.
		//
		// D-2.6.2 SANCTIONS this as the gate's FIFTH obligation, and the
		// criterion it ruled on is that this is the ONLY cross-target
		// artifact for a shipped FR30: refusing it would ship FR30 with
		// none. Registration now is what gives the gate something to
		// compare — "a golden registered in one list and not the other is a
		// matrix leg nobody compares, reported as green."
		label:            "multi-page (pagination, running header and footer on every page)",
		slug:             "multi-page",
		capture:          captureMultiPageRender,
		fixtureRelPath:   []string{"fixtures", "multi-page", "expected.json"},
		requireFontFile2: true,
		extraGuard:       requireMultiPageIsGenuinelyMultiPage,
		wantPages:        2,
	},
}

// TestCrossTargetByteIdentity is AC1's single local entry point and AC7's
// pairwise-plus-golden check, run all in one process, for BOTH documents
// AC15 requires (AC10a, the D-000.4 override for this story). For every
// target it builds ONE render/toolchain test binary (both selectors live
// in the same binary — TestMain, render_test.go), runs AC3's toolchain
// gate before anything else, then captures, structurally validates
// (AC6) and hashes EACH document's render from that same binary. It then
// implements AC17's two-case divergence report per document: all four
// agreeing with each other but not the golden is a legitimate versioned
// change under AD-22 to be investigated and re-recorded by a human; any
// pairwise disagreement is NFR1 falsified. It never re-records a
// fixture, majority-votes, or auto-selects the host's value (D-1.2.2).
func TestCrossTargetByteIdentity(t *testing.T) {
	assertExactlyFourMatrixTargets(t)

	root := repoRootFromTest(t)
	buildDir := matrixBuildDir(t, root)

	fixtures := make(map[string]expectedFixture, len(matrixDocuments))
	for _, doc := range matrixDocuments {
		fixtures[doc.label] = loadExpectedFixture(t, filepath.Join(append([]string{root}, doc.fixtureRelPath...)...))
	}
	assertFixturesShareToolchain(t, fixtures)

	resultsByDoc := map[string][]renderLegResult{}
	toolchainLegs := 0
	for _, target := range matrixTargets {
		binPath := buildRenderTestBinary(t, root, target, buildDir)

		// AC3: toolchain gate runs before any hash is compared, no skip —
		// once per target binary, shared by both documents it renders.
		assertToolchainWitness(t, target, binPath, fixtures[matrixDocuments[0].label].GoToolchain)
		toolchainLegs++

		for _, doc := range matrixDocuments {
			raw := doc.capture(t, target, binPath)
			assertWellFormedPDF(t, target.name+" ("+doc.label+")", raw, doc.pageCount()) // AC6, every leg's own output

			if doc.requireFontFile2 && !containsFontFile2(raw) {
				// AC10b: the font child's output must be confirmed to
				// actually contain a FontFile2 BEFORE any byte
				// comparison runs — without this, AC10/AC15 can pass on
				// two identical fontless renders and certify nothing
				// (F-6, D-000.9).
				t.Fatalf(
					"target %s (%s): rendered output does not contain a FontFile2 (AC10b vacuity guard) — "+
						"the D-000.4 matrix override for this story is only meaningful if the subprocess "+
						"font path actually embeds a font",
					target.name, doc.label,
				)
			}
			if doc.requireImageXObject && !containsImageXObject(raw) {
				// AC23 (D-1.8.6): the image child's output must be
				// confirmed to actually contain an image XObject
				// (/Subtype /Image) BEFORE any byte comparison runs, on
				// every leg — otherwise the matrix could compare four
				// identical IMAGELESS PDFs and report byte-identity,
				// exactly what Story 1.5 discovered about fonts.
				t.Fatalf(
					"target %s (%s): rendered output does not contain an image XObject (AC23 vacuity guard)",
					target.name, doc.label,
				)
			}
			if doc.extraGuard != nil {
				doc.extraGuard(t, target, raw)
			}

			sum := sha256.Sum256(raw)
			hash := hex.EncodeToString(sum[:])
			resultsByDoc[doc.label] = append(resultsByDoc[doc.label], renderLegResult{target: target, hash: hash, bytes: raw})
		}
	}

	// AC11: assert exactly four toolchain legs, against the literal
	// wantMatrixLegs constant, never against len(matrixTargets) — that
	// comparison is the loop's own output against its own input and
	// cannot fail (Blocker 1).
	if toolchainLegs != wantMatrixLegs {
		t.Fatalf("expected exactly %d toolchain legs, got %d (AC11)", wantMatrixLegs, toolchainLegs)
	}

	for _, doc := range matrixDocuments {
		results := resultsByDoc[doc.label]
		if len(results) != wantMatrixLegs {
			t.Fatalf("%s: expected exactly %d render legs, got %d (AC11)", doc.label, wantMatrixLegs, len(results))
		}
		checkCrossTargetResults(t, buildDir, doc.label, results, fixtures[doc.label])
	}
}

// checkCrossTargetResults implements AC17's two-case divergence report
// for one document's four render legs.
func checkCrossTargetResults(t *testing.T, buildDir, docLabel string, results []renderLegResult, fixture expectedFixture) {
	t.Helper()

	table := strings.Join(renderTableRows(results), "\n")
	t.Logf("cross-target render matrix (%s):\n%s", docLabel, table)

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

	diagDir := filepath.Join(buildDir, "diverged", strings.ReplaceAll(docLabel, " ", "_"))
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
			"%s: all four targets agree with EACH OTHER but NOT with the recorded golden "+
				"(sha256=%s):\n%s\n\n"+
				"This is a legitimate versioned change under AD-22 (e.g. a Go toolchain bump, or a "+
				"deliberate, documented rendering-format change) to be investigated and "+
				"deliberately re-recorded by a human — never auto-fixed by this harness. The "+
				"harness never re-records the fixture, majority-votes, or auto-selects the host's "+
				"value (D-1.2.2). Diverging bytes written under %s.",
			docLabel, fixture.SHA256, table, diagDir,
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
		"%s: cross-target byte identity FAILED — targets disagree with EACH OTHER, not merely with "+
			"the golden. This falsifies NFR1, the product's core claim:\n%s\n\n%s\n\n"+
			"Diverging bytes written under %s. This harness never re-records the fixture, "+
			"majority-votes three targets against one, or auto-selects the host's value (D-1.2.2). "+
			"A real cross-target divergence is an owner decision, not one this harness or a "+
			"developer resolves by choosing a value.",
		docLabel, table, strings.Join(diagLines, "\n"), diagDir,
	)
}

// TestTargetRenderHash is the single-target mode CI's per-target jobs
// call (AC1, AC13): it exercises the same code path as
// TestCrossTargetByteIdentity's per-leg body — toolchain gate,
// structural validation, golden comparison — for exactly the one target
// named by FOLIO_MATRIX_TARGET, and writes ONE HASH FILE PER DOCUMENT
// (matrixDocuments, Finding 8, QA review) the workflow can publish as
// build artifacts. Before this fix, this test captured ONLY
// minimal-rect (the fontless document), so CI's per-target jobs never
// regression-tested cross-target byte-identity for font-text — the
// document whose determinism depends on textshape's internal glyph-map
// ordering, "the first property in the project whose determinism
// depends on a third-party module's internal ordering" (story header).
// It contains no hash logic of its own beyond what
// TestCrossTargetByteIdentity already has; the workflow's compare job
// re-derives the pairwise property from the published hash files rather
// than reimplementing hashing in YAML.
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

	fixtures := make(map[string]expectedFixture, len(matrixDocuments))
	for _, doc := range matrixDocuments {
		fixtures[doc.label] = loadExpectedFixture(t, filepath.Join(append([]string{root}, doc.fixtureRelPath...)...))
	}
	assertFixturesShareToolchain(t, fixtures)

	binPath := buildRenderTestBinary(t, root, target, buildDir)
	assertToolchainWitness(t, target, binPath, fixtures[matrixDocuments[0].label].GoToolchain)

	for _, doc := range matrixDocuments {
		raw := doc.capture(t, target, binPath)
		// doc.pageCount(), NOT a literal 1: this loop runs over EVERY
		// registered document, so a hard-coded page count here would fatal
		// on the multi-page document at the GATE — the one place this
		// matrix-tagged path runs, and the last place anyone wants to
		// discover it. Same defect class as the one that let the first
		// multi-page golden ship: a checker encoding the single-page shape
		// as an invariant.
		assertWellFormedPDF(t, target.name+" ("+doc.label+")", raw, doc.pageCount())

		if doc.requireFontFile2 && !containsFontFile2(raw) {
			// AC10b's vacuity guard, same reasoning as
			// TestCrossTargetByteIdentity: a subprocess that silently
			// fell back to the fontless path must not be laundered into
			// a passing per-target CI job either.
			t.Fatalf(
				"target %s (%s): rendered output does not contain a FontFile2 (AC10b vacuity guard)",
				target.name, doc.label,
			)
		}
		if doc.requireImageXObject && !containsImageXObject(raw) {
			// AC23's same vacuity guard, applied to the per-target CI
			// entry point.
			t.Fatalf(
				"target %s (%s): rendered output does not contain an image XObject (AC23 vacuity guard)",
				target.name, doc.label,
			)
		}
		if doc.extraGuard != nil {
			doc.extraGuard(t, target, raw)
		}

		sum := sha256.Sum256(raw)
		hash := hex.EncodeToString(sum[:])

		hashOutPath := filepath.Join(buildDir, "hash."+target.goos+"-"+target.goarch+"."+doc.slug+".txt")
		if err := os.WriteFile(hashOutPath, []byte(hash+"\n"), 0o644); err != nil {
			t.Fatalf("write hash output for %s (%s): %v", target.name, doc.label, err)
		}
		t.Logf("target %s (%s): sha256=%s (%d bytes; written to %s)", target.name, doc.label, hash, len(raw), hashOutPath)

		fixture := fixtures[doc.label]
		if hash != fixture.SHA256 {
			t.Fatalf("target %s (%s): sha256 %s does not match recorded golden %s",
				target.name, doc.label, hash, fixture.SHA256)
		}
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
