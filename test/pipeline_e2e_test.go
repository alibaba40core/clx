// Package test holds end-to-end tests that exercise the full parse → intent →
// translate → safety → exec pipeline on a synthetic ~/.clx home.
//
// The test matrix matches the Phase 1.6 exit criteria: every seed intent has
// an explain test (cross-platform), and a small set of exec tests run real
// binaries on Linux/macOS/Windows with documented skip reasons (missing
// docker daemon, optional tools, etc.). Shell-native cmdlets use validated
// host scripts (Phase 1.7) on Windows PowerShell/CMD.
package test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/pipeline"
	"github.com/alibaba40core/clx/internal/policy"
)

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// setupCLXHome bootstraps a per-test ~/.clx and seeds a profile keyed to
// (osName, shell). Useful only when the seeded key matches the host runtime;
// otherwise pipeline.Run → LoadOrDetect re-detects and writes the real
// profile alongside the seeded one. Prefer setupCLXHomeForHost for tests that
// must rely on a specific tool list to drive strategy selection.
func setupCLXHome(t *testing.T, osName, shell string, tools []string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	if _, err := config.Bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}
	policy.ResetCache()

	path, err := config.SystemProfilePath()
	if err != nil {
		t.Fatal(err)
	}
	store := environment.NewProfileStore()
	store.UpsertProfile(environment.SystemProfile{
		OS:             osName,
		Shell:          shell,
		AvailableTools: tools,
	})
	if err := environment.SaveStore(context.Background(), path, store); err != nil {
		t.Fatal(err)
	}
}

// setupCLXHomeForHost bootstraps a per-test ~/.clx and seeds a profile keyed
// to the actual host so LoadOrDetect returns it without re-detecting. If
// tools is non-nil, it overrides the detected tool list (useful to force
// grep over rg, find over fd, etc. so strategy selection is deterministic).
func setupCLXHomeForHost(t *testing.T, tools []string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	if _, err := config.Bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}
	policy.ResetCache()

	// First call runs full detection and persists a host-keyed profile.
	p, err := environment.LoadOrDetect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if tools == nil {
		return
	}

	path, err := config.SystemProfilePath()
	if err != nil {
		t.Fatal(err)
	}
	store, err := environment.LoadStore(context.Background(), path)
	if err != nil {
		store = environment.NewProfileStore()
	}
	p.AvailableTools = tools
	store.UpsertProfile(p)
	if err := environment.SaveStore(context.Background(), path, store); err != nil {
		t.Fatal(err)
	}
}

// runResult is the captured outcome of one pipeline.Run call.
type runResult struct {
	code   int
	err    error
	stdout string
	stderr string
}

// runPipeline executes pipeline.Run with captured stdout/stderr buffers.
// It does not assert on the result; callers inspect the returned runResult.
func runPipeline(t *testing.T, cfg config.Config, input string, opts pipeline.Options) runResult {
	t.Helper()
	var stdout, stderr bytes.Buffer
	opts.Stdout = &stdout
	opts.Stderr = &stderr
	code, err := pipeline.Run(context.Background(), cfg, input, opts)
	return runResult{code: code, err: err, stdout: stdout.String(), stderr: stderr.String()}
}

// skipUnlessTool skips the current test if a binary is missing from PATH.
func skipUnlessTool(t *testing.T, name string) {
	t.Helper()
	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("skip: %s not in PATH", name)
	}
}

// writePolicy overwrites ~/.clx/policies/policy.yaml with the given block-list
// and clears the policy load cache so the next pipeline.Run picks it up.
func writePolicy(t *testing.T, blocked []string) {
	t.Helper()
	p, err := config.PolicyPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	b.WriteString("blocked:\n")
	for _, pattern := range blocked {
		fmt.Fprintf(&b, "  - %q\n", pattern)
	}
	if err := os.WriteFile(p, []byte(b.String()), 0o600); err != nil {
		t.Fatal(err)
	}
	policy.ResetCache()
}

// commandLine returns the "Command: ..." line from the explain display, or "".
func commandLine(stdout string) string {
	for _, ln := range strings.Split(stdout, "\n") {
		if strings.HasPrefix(ln, "Command:") {
			return ln
		}
	}
	return ""
}

// -----------------------------------------------------------------------------
// Explain — cross-platform (intent-name assertion only)
//
// Runs on all three CI OSes. Verifies every seed intent reaches the display
// stage with the correct intent name. Strategy-specific assertions live in
// the OS-gated tests below.
// -----------------------------------------------------------------------------

func TestE2EExplainSeedIntents(t *testing.T) {
	setupCLXHomeForHost(t, nil)

	cases := []struct {
		input  string
		intent string
	}{
		{"grep errors logs.txt", "search_text_in_file"},
		{"locate help.txt", "find_file"},
		{"ls .", "list_dir"},
		{"pwd", "current_dir"},
		{"disk usage", "disk_usage"},
		{"git status", "git_status"},
		{"git log -n 5", "git_log"},
		{"git diff", "git_diff"},
		{"git diff main.go", "git_diff_path"},
		{"git branch", "git_branch_list"},
		{"docker ps", "docker_ps"},
		{"docker images", "docker_images"},
		{"docker logs web", "docker_logs"},
		{"ping example.com", "ping_host"},
		{"curl -I https://example.com", "curl_url"},
		{"netstat -an", "netstat_listening"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			r := runPipeline(t, config.Default(), tc.input, pipeline.Options{Explain: true})
			if r.err != nil || r.code != 0 {
				t.Fatalf("code=%d err=%v stderr=%s", r.code, r.err, r.stderr)
			}
			if !strings.Contains(r.stdout, tc.intent) {
				t.Fatalf("expected intent %q in stdout=%q", tc.intent, r.stdout)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Explain — host-strategy assertions
//
// Verify the actually-selected command for the host's profile. Gated by
// runtime.GOOS so each variant only runs on its target CI matrix entry.
// -----------------------------------------------------------------------------

func TestE2EExplainHostStrategyUnix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix host-strategy test")
	}
	// Restrict tools so deterministic Linux/darwin strategies win:
	//   grep over rg (search_text_in_file), find over fd (find_file).
	setupCLXHomeForHost(t, []string{"grep", "ping"})

	cases := []struct {
		input string
		want  string
	}{
		{"pwd", "pwd"},
		{"ls .", "ls"},
		{"grep errors logs.txt", "grep"},
		{"locate help.txt", "find"},
		{"disk usage", "df"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			r := runPipeline(t, config.Default(), tc.input, pipeline.Options{Explain: true})
			if r.code != 0 {
				t.Fatalf("code=%d stderr=%s", r.code, r.stderr)
			}
			line := commandLine(r.stdout)
			if !strings.Contains(line, tc.want) {
				t.Fatalf("expected %q in command line %q (stdout=%q)", tc.want, line, r.stdout)
			}
		})
	}
}

func TestE2EExplainHostStrategyWindowsPowerShell(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows host-strategy test")
	}
	t.Setenv("POWERSHELL_VERSION", "7.4.0")
	t.Setenv("POWERSHELL_DISTRO_NAME", "")
	setupCLXHomeForHost(t, nil)

	cases := []struct {
		input string
		want  string
	}{
		{"pwd", "Get-Location"},
		{"locate help.txt", "Get-ChildItem"},
		{"grep errors logs.txt", "Select-String"},
		{"ls .", "Get-ChildItem"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			r := runPipeline(t, config.Default(), tc.input, pipeline.Options{Explain: true})
			if r.code != 0 {
				t.Fatalf("code=%d stderr=%s", r.code, r.stderr)
			}
			line := commandLine(r.stdout)
			if !strings.Contains(line, tc.want) {
				t.Fatalf("expected %q in command line %q (stdout=%q)", tc.want, line, r.stdout)
			}
		})
	}
}

func TestE2EExplainHostStrategyWindowsBash(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows host-strategy test")
	}
	t.Setenv("MSYSTEM", "MINGW64")
	t.Setenv("SHELL", `/usr/bin/bash`)
	t.Setenv("POWERSHELL_VERSION", "")
	t.Setenv("POWERSHELL_DISTRO_NAME", "")
	setupCLXHomeForHost(t, []string{"grep"})

	cases := []struct {
		input string
		want  string
	}{
		{"pwd", "pwd"},
		{"grep errors logs.txt", "grep"},
		{"ls .", "ls"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			r := runPipeline(t, config.Default(), tc.input, pipeline.Options{Explain: true})
			if r.code != 0 {
				t.Fatalf("code=%d stderr=%s", r.code, r.stderr)
			}
			line := commandLine(r.stdout)
			if !strings.Contains(line, tc.want) {
				t.Fatalf("expected %q in command line %q (stdout=%q)", tc.want, line, r.stdout)
			}
		})
	}
}

func TestE2EExplainHostStrategyWindowsCmd(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows host-strategy test")
	}
	t.Setenv("POWERSHELL_VERSION", "")
	t.Setenv("POWERSHELL_DISTRO_NAME", "")
	t.Setenv("ComSpec", `C:\Windows\System32\cmd.exe`)
	setupCLXHomeForHost(t, nil)

	cases := []struct {
		input string
		want  string
	}{
		{"pwd", "cd"},
		{"locate help.txt", "dir /s"},
		{"grep errors logs.txt", "findstr"},
		{"ls .", "dir ."},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			r := runPipeline(t, config.Default(), tc.input, pipeline.Options{Explain: true})
			if r.code != 0 {
				t.Fatalf("code=%d stderr=%s", r.code, r.stderr)
			}
			line := commandLine(r.stdout)
			if !strings.Contains(line, tc.want) {
				t.Fatalf("expected %q in command line %q (stdout=%q)", tc.want, line, r.stdout)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Exec — real binaries
//
// All exec tests opt out of the default dry-run via cfg.Safety.DryRun = false
// and pass Yes: true to bypass confirm. They are additionally gated by
// runtime.GOOS or exec.LookPath where the underlying binary isn't portable.
// -----------------------------------------------------------------------------

// TestE2EExecPwdUnix executes the Linux/darwin `pwd` binary.
func TestE2EExecPwdUnix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix only")
	}
	skipUnlessTool(t, "pwd")
	setupCLXHomeForHost(t, nil)

	cfg := config.Default()
	cfg.Safety.DryRun = false
	r := runPipeline(t, cfg, "pwd", pipeline.Options{Yes: true})
	if r.code != 0 || r.err != nil {
		t.Fatalf("code=%d err=%v stderr=%s", r.code, r.err, r.stderr)
	}
}

// TestE2EExecListDirUnix executes `ls` on Linux/darwin.
func TestE2EExecListDirUnix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix only")
	}
	skipUnlessTool(t, "ls")
	setupCLXHomeForHost(t, nil)

	cfg := config.Default()
	cfg.Safety.DryRun = false
	r := runPipeline(t, cfg, "ls .", pipeline.Options{Yes: true})
	if r.code != 0 || r.err != nil {
		t.Fatalf("code=%d err=%v stderr=%s", r.code, r.err, r.stderr)
	}
}

// TestE2EExecPwdWindows runs pwd via PowerShell Get-Location host script.
func TestE2EExecPwdWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows only")
	}
	skipUnlessTool(t, "powershell")
	setupCLXHomeForHost(t, nil)

	cfg := config.Default()
	cfg.Safety.DryRun = false
	r := runPipeline(t, cfg, "pwd", pipeline.Options{Yes: true})
	if r.code != 0 || r.err != nil {
		t.Fatalf("code=%d err=%v stderr=%s", r.code, r.err, r.stderr)
	}
}

// TestE2EExecListDirWindows runs ls via Get-ChildItem host script.
func TestE2EExecListDirWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows only")
	}
	skipUnlessTool(t, "powershell")
	setupCLXHomeForHost(t, nil)

	cfg := config.Default()
	cfg.Safety.DryRun = false
	r := runPipeline(t, cfg, "ls .", pipeline.Options{Yes: true})
	if r.code != 0 || r.err != nil {
		t.Fatalf("code=%d err=%v stderr=%s", r.code, r.err, r.stderr)
	}
}

// TestE2EExecSearchTextWindows runs grep via Select-String host script.
func TestE2EExecSearchTextWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows only")
	}
	skipUnlessTool(t, "powershell")
	setupCLXHomeForHost(t, []string{"grep"})

	logFile := filepath.Join(t.TempDir(), "logs.txt")
	if err := os.WriteFile(logFile, []byte("line one\nerrors here\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	cfg.Safety.DryRun = false
	input := fmt.Sprintf("grep errors %s", logFile)
	r := runPipeline(t, cfg, input, pipeline.Options{Yes: true})
	if r.code != 0 || r.err != nil {
		t.Fatalf("code=%d err=%v stderr=%s stdout=%s", r.code, r.err, r.stderr, r.stdout)
	}
}

// TestE2EExecGitStatus is the cross-platform exec anchor. git ships an
// identical CLI on Linux/macOS/Windows, so the rule's default: strategy
// renders `git status` everywhere. The test runs from inside the clx repo
// (a git repo) so git status exits 0 on all CI runners.
func TestE2EExecGitStatus(t *testing.T) {
	skipUnlessTool(t, "git")
	setupCLXHomeForHost(t, nil)

	cfg := config.Default()
	cfg.Safety.DryRun = false
	r := runPipeline(t, cfg, "git status", pipeline.Options{Yes: true})
	if r.code != 0 || r.err != nil {
		t.Fatalf("code=%d err=%v stderr=%s", r.code, r.err, r.stderr)
	}
}

// TestE2EExecDockerPs runs `docker ps` on Linux only. CI macOS/Windows
// runners do not ship with a docker daemon, so the test skips there even if
// the docker CLI happens to be present. On Linux it additionally skips if
// the daemon is unreachable (`docker info` fails).
func TestE2EExecDockerPs(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skipf("docker exec test runs only on linux (no daemon on CI %s runners)", runtime.GOOS)
	}
	skipUnlessTool(t, "docker")
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("docker daemon not reachable")
	}
	setupCLXHomeForHost(t, nil)

	cfg := config.Default()
	cfg.Safety.DryRun = false
	r := runPipeline(t, cfg, "docker ps", pipeline.Options{Yes: true})
	if r.code != 0 || r.err != nil {
		t.Fatalf("code=%d err=%v stderr=%s", r.code, r.err, r.stderr)
	}
}

// -----------------------------------------------------------------------------
// Policy & safety
// -----------------------------------------------------------------------------

// TestE2EPolicyBlocks installs a custom policy that blocks any command
// containing "git", then verifies pipeline.Run rejects `git status` with
// policy.ErrBlocked before reaching the executor. Cross-platform: the
// block-list check runs the same on every OS.
func TestE2EPolicyBlocks(t *testing.T) {
	setupCLXHomeForHost(t, nil)
	writePolicy(t, []string{"git"})

	cfg := config.Default()
	cfg.Safety.DryRun = false
	r := runPipeline(t, cfg, "git status", pipeline.Options{Yes: true})

	if r.code == 0 {
		t.Fatalf("expected non-zero exit, got stdout=%q stderr=%q", r.stdout, r.stderr)
	}
	if !errors.Is(r.err, policy.ErrBlocked) {
		t.Fatalf("expected policy.ErrBlocked, got err=%v", r.err)
	}
	if !strings.Contains(r.stderr, "blocked by policy") {
		t.Fatalf("expected 'blocked by policy' in stderr=%q", r.stderr)
	}
}

// TestE2EConfigDryRunDefault verifies the fresh-install default
// (safety.dry_run: true) prevents real execution even when the caller passes
// Yes: true and no --dry-run flag. This is the Step 2 exit-criterion test.
func TestE2EConfigDryRunDefault(t *testing.T) {
	setupCLXHomeForHost(t, nil)

	r := runPipeline(t, config.Default(), "pwd", pipeline.Options{Yes: true})
	if r.code != 0 || r.err != nil {
		t.Fatalf("code=%d err=%v stderr=%s", r.code, r.err, r.stderr)
	}
	if !strings.Contains(r.stdout, "dry-run:") {
		t.Fatalf("expected dry-run output, got stdout=%q", r.stdout)
	}
}

// -----------------------------------------------------------------------------
// Cwd independence (embedded rules)
// -----------------------------------------------------------------------------

// TestE2EWorksFromNonRepoCwd verifies the pipeline resolves intents when the
// process cwd is outside the source tree (production install on PATH).
func TestE2EWorksFromNonRepoCwd(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	setupCLXHomeForHost(t, nil)
	r := runPipeline(t, config.Default(), "pwd", pipeline.Options{Explain: true})
	if r.err != nil || r.code != 0 {
		t.Fatalf("code=%d err=%v stderr=%s", r.code, r.err, r.stderr)
	}
	if !strings.Contains(r.stdout, "current_dir") {
		t.Fatalf("expected current_dir in stdout=%q", r.stdout)
	}
}

// -----------------------------------------------------------------------------
// Profile persistence (smoke)
// -----------------------------------------------------------------------------

func TestE2EProfileWritten(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	if _, err := config.Bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}
	path, err := config.SystemProfilePath()
	if err != nil {
		t.Fatal(err)
	}
	p, err := environment.LoadOrDetect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if p.OS == "" {
		t.Fatal("expected detect to fill OS")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var store environment.ProfileStore
	if err := json.Unmarshal(data, &store); err != nil {
		t.Fatal(err)
	}
	for _, prof := range store.Profiles {
		if prof.OS != "" {
			return
		}
	}
	t.Fatal("profile store has no profiles with OS")
}
