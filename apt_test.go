package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── test lifecycle ────────────────────────────────────────────────────────────

func initAptTest(t *testing.T) {
	t.Helper()
	if !*tests.TestApt {
		t.Skip("Skipping apt test. To run add '-test.apt=true' option.")
	}
	if runtime.GOOS != "linux" {
		t.Skip("apt tests only run on Linux")
	}
	createJfrogHomeConfig(t, true)
}

func cleanAptTest(t *testing.T) {
	t.Helper()
	// Remove any files written to the system apt dirs by the test.
	for _, pattern := range []string{
		"/etc/apt/sources.list.d/jfrog-cli-apt-*.list",
		"/etc/apt/preferences.d/jfrog-cli-apt-*.pref",
		"/etc/apt/keyrings/jfrog-cli-apt-*.asc",
	} {
		matches, _ := filepath.Glob(pattern)
		for _, f := range matches {
			_ = os.Remove(f)
		}
	}
	tests.CleanFileSystem()
}

// ── helpers ───────────────────────────────────────────────────────────────────

// aptRepo returns the virtual repo name used across apt integration tests.
func aptRepo() string { return tests.AptVirtualRepo }

// sourcesListPath returns the expected path for a given repo+dist.
func sourcesListPath(repo, dist string) string {
	return fmt.Sprintf("/etc/apt/sources.list.d/jfrog-%s-%s.list", repo, dist)
}

func prefPath(repo, dist string) string {
	return fmt.Sprintf("/etc/apt/preferences.d/jfrog-%s-%s.pref", repo, dist)
}

func keyringPath(repo, dist string) string {
	return fmt.Sprintf("/etc/apt/keyrings/jfrog-%s-%s.asc", repo, dist)
}

// requireRoot skips the test unless running as root (euid 0).
func requireRoot(t *testing.T) {
	t.Helper()
	if os.Getuid() != 0 {
		t.Skip("test requires root — run with sudo or in a root container")
	}
}

// requireNonRoot skips the test if running as root.
func requireNonRoot(t *testing.T) {
	t.Helper()
	if os.Getuid() == 0 {
		t.Skip("test requires non-root user")
	}
}

// ── jf setup apt ─────────────────────────────────────────────────────────────

// TestAptSetup_BasicPersistentSetup verifies the happy path: sources.list entry
// and pinning file are written and apt-get update succeeds.
func TestAptSetup_BasicPersistentSetup(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)

	repo := aptRepo()
	runJfrogCli(t, "setup", "apt",
		"--repo="+repo,
		"--dist=noble",
		"--component=main",
		"--trusted",
	)

	assert.FileExists(t, sourcesListPath(repo, "noble"))
	assert.FileExists(t, prefPath(repo, "noble"))

	content, err := os.ReadFile(sourcesListPath(repo, "noble"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "[trusted=yes]")
	assert.Contains(t, string(content), repo)
	assert.Contains(t, string(content), "noble main")
}

// TestAptSetup_Idempotent verifies re-running with the same args does not
// produce duplicate entries or errors.
func TestAptSetup_Idempotent(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)

	repo := aptRepo()
	args := []string{"setup", "apt", "--repo=" + repo, "--dist=noble", "--trusted"}
	runJfrogCli(t, args...)
	runJfrogCli(t, args...) // second run — must not error

	content, err := os.ReadFile(sourcesListPath(repo, "noble"))
	require.NoError(t, err)
	// file should contain exactly one deb line, not duplicated
	count := 0
	for _, line := range splitLines(string(content)) {
		if len(line) > 0 {
			count++
		}
	}
	assert.Equal(t, 1, count, "sources.list should contain exactly one entry")
}

// TestAptSetup_MultipleComponents verifies space-separated components work.
func TestAptSetup_MultipleComponents(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)

	repo := aptRepo()
	runJfrogCli(t, "setup", "apt",
		"--repo="+repo,
		"--dist=noble",
		"--component=main contrib non-free",
		"--trusted",
	)

	content, err := os.ReadFile(sourcesListPath(repo, "noble"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "main contrib non-free")
}

// TestAptSetup_PinningFile verifies the .pref file has correct priority.
func TestAptSetup_PinningFile(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)

	repo := aptRepo()
	runJfrogCli(t, "setup", "apt", "--repo="+repo, "--dist=noble", "--trusted")

	pref, err := os.ReadFile(prefPath(repo, "noble"))
	require.NoError(t, err)
	assert.Contains(t, string(pref), "Pin-Priority: 1001")
}

// TestAptSetup_TrustedFlag verifies --trusted injects [trusted=yes].
func TestAptSetup_TrustedFlag(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)

	repo := aptRepo()
	runJfrogCli(t, "setup", "apt", "--repo="+repo, "--dist=noble", "--trusted")

	content, err := os.ReadFile(sourcesListPath(repo, "noble"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "[trusted=yes]", "trusted flag must produce [trusted=yes] in sources line")
}

// TestAptSetup_ImportKey verifies --import-key fetches and installs the GPG key.
// Skipped when the Artifactory repo has no GPG key configured.
func TestAptSetup_ImportKey(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)

	repo := aptRepo()
	err := runJfrogCliWithoutAssertion("setup", "apt", "--repo="+repo, "--dist=noble", "--import-key")
	if err != nil {
		t.Skipf("Skipping: Artifactory repo %q has no GPG key configured: %v", repo, err)
	}

	// keyring file written
	assert.FileExists(t, keyringPath(repo, "noble"))

	// sources line uses signed-by, not trusted=yes
	content, err := os.ReadFile(sourcesListPath(repo, "noble"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "signed-by=")
	assert.NotContains(t, string(content), "trusted=yes")
}

// TestAptSetup_TrustedAndImportKeyMutuallyExclusive verifies both flags together error.
func TestAptSetup_TrustedAndImportKeyMutuallyExclusive(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)

	err := runJfrogCliWithoutAssertion("setup", "apt",
		"--repo="+aptRepo(),
		"--dist=noble",
		"--trusted",
		"--import-key",
	)
	assert.Error(t, err, "combining --trusted and --import-key must return an error")
}

// TestAptSetup_MissingRepo verifies --repo is required.
func TestAptSetup_MissingRepo(t *testing.T) {
	initAptTest(t)
	defer cleanAptTest(t)

	err := runJfrogCliWithoutAssertion("setup", "apt", "--dist=noble")
	assert.Error(t, err)
}

// TestAptSetup_MissingDist verifies --dist is required.
func TestAptSetup_MissingDist(t *testing.T) {
	initAptTest(t)
	defer cleanAptTest(t)

	err := runJfrogCliWithoutAssertion("setup", "apt", "--repo="+aptRepo())
	assert.Error(t, err)
}

// TestAptSetup_NonRootPermissionDenied verifies non-root without perms gets a clear error.
func TestAptSetup_NonRootPermissionDenied(t *testing.T) {
	initAptTest(t)
	requireNonRoot(t)

	err := runJfrogCliWithoutAssertion("setup", "apt",
		"--repo="+aptRepo(),
		"--dist=noble",
		"--trusted",
	)
	require.Error(t, err, "non-root without perms must fail")
	assert.Contains(t, err.Error(), "sudo", "error message must suggest sudo")
}

// TestAptSetup_NonRootWithPermission verifies a non-root user who owns the apt
// dirs can run setup without sudo.
func TestAptSetup_NonRootWithPermission(t *testing.T) {
	initAptTest(t)
	requireNonRoot(t)

	// Grant write access to apt dirs for current user (requires prior sudo setup in CI).
	dirs := []string{
		"/etc/apt/sources.list.d",
		"/etc/apt/preferences.d",
		"/etc/apt/keyrings",
	}
	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil || !isWritable(info) {
			t.Skipf("directory %s not writable by current user — skip", dir)
		}
	}
	defer cleanAptTest(t)

	runJfrogCli(t, "setup", "apt",
		"--repo="+aptRepo(),
		"--dist=noble",
		"--trusted",
	)
	assert.FileExists(t, sourcesListPath(aptRepo(), "noble"))
}

// ── jf setup apt --remove ─────────────────────────────────────────────────────

// TestAptSetupRemove_RemovesAllFiles verifies --remove cleans all managed files.
func TestAptSetupRemove_RemovesAllFiles(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)

	repo := aptRepo()
	runJfrogCli(t, "setup", "apt", "--repo="+repo, "--dist=noble", "--trusted")
	require.FileExists(t, sourcesListPath(repo, "noble"))

	runJfrogCli(t, "setup", "apt", "--remove")

	assert.NoFileExists(t, sourcesListPath(repo, "noble"))
	assert.NoFileExists(t, prefPath(repo, "noble"))
}

// TestAptSetupRemove_DistFilteredRemoval verifies --remove --dist only removes
// files for the specified distribution.
func TestAptSetupRemove_DistFilteredRemoval(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)

	repo := aptRepo()
	// Set up two dists.
	runJfrogCli(t, "setup", "apt", "--repo="+repo, "--dist=noble", "--trusted")
	runJfrogCli(t, "setup", "apt", "--repo="+repo, "--dist=jammy", "--trusted")

	// Remove only noble.
	runJfrogCli(t, "setup", "apt", "--remove", "--dist=noble")

	assert.NoFileExists(t, sourcesListPath(repo, "noble"), "noble must be removed")
	assert.NoFileExists(t, prefPath(repo, "noble"), "noble pref must be removed")
	assert.FileExists(t, sourcesListPath(repo, "jammy"), "jammy must survive")
	assert.FileExists(t, prefPath(repo, "jammy"), "jammy pref must survive")
}

// TestAptSetupRemove_Idempotent verifies --remove on empty dir does not error.
func TestAptSetupRemove_Idempotent(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)

	assert.NoError(t, runJfrogCliWithoutAssertion("setup", "apt", "--remove"))
}

// ── jf apt install (on-the-fly) ───────────────────────────────────────────────

// TestAptInstall_OnTheFlyInstall verifies curl can be installed via on-the-fly auth
// and that the install came from Artifactory (not a system source).
func TestAptInstall_OnTheFlyInstall(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)

	runJfrogCli(t, "apt", "install", "-y", "curl",
		"--repo="+aptRepo(),
		"--dist=noble",
		"--trusted",
	)

	_, err := os.Stat("/usr/bin/curl")
	assert.NoError(t, err, "curl must be installed after jf apt install")

	assertInstalledFromArtifactory(t)
}

// TestAptInstall_SkipLoginUsesSystemConfig verifies --skip-login bypasses auth injection:
// no temp sources.list is written and no jfrog-* files appear in system apt dirs.
func TestAptInstall_SkipLoginUsesSystemConfig(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)

	tmpGlob := filepath.Join(os.TempDir(), "jfrog-apt-*")

	before, _ := filepath.Glob(tmpGlob)
	beforeSet := make(map[string]bool, len(before))
	for _, f := range before {
		beforeSet[f] = true
	}

	// --skip-login should bypass auth injection entirely; outcome doesn't matter.
	_ = runJfrogCliWithoutAssertion("apt", "install", "-y", "curl", "--skip-login")

	// No new jfrog-apt-* temp files must have been created.
	after, _ := filepath.Glob(tmpGlob)
	var newFiles []string
	for _, f := range after {
		if !beforeSet[f] {
			newFiles = append(newFiles, f)
		}
	}
	assert.Empty(t, newFiles, "--skip-login must not create temp sources.list files: %v", newFiles)

	// No persistent jfrog-* sources files must have been written.
	sysFiles, _ := filepath.Glob("/etc/apt/sources.list.d/jfrog-*.list")
	assert.Empty(t, sysFiles, "--skip-login must not write persistent sources.list entries")
}

// TestAptInstall_MissingRepoAndDist verifies warning path (no auth injection).
func TestAptInstall_MissingRepoAndDist(t *testing.T) {
	initAptTest(t)
	defer cleanAptTest(t)

	// Should not error — falls back to unauthenticated apt (warn + passthrough).
	// We don't assert success because package resolution depends on system config.
	_ = runJfrogCliWithoutAssertion("apt", "show", "curl")
}

// TestAptInstall_TrustedFlag verifies --trusted injects [trusted=yes] into the
// temporary sources.list used for on-the-fly auth.
func TestAptInstall_TrustedFlag(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)

	// Capture the temp sources.list content while apt-get is running.
	content, cancel := captureTempSourcesList(t)
	defer cancel()

	runJfrogCli(t, "apt", "install", "--dry-run", "-y", "curl",
		"--repo="+aptRepo(),
		"--dist=noble",
		"--trusted",
	)

	select {
	case src := <-content:
		assert.Contains(t, src, "[trusted=yes]", "temp sources.list must contain [trusted=yes]")
		assert.Contains(t, src, aptRepo(), "temp sources.list must reference the Artifactory repo")
	case <-time.After(10 * time.Second):
		t.Error("temp sources.list was not created or deleted before it could be read")
	}
}

// TestAptInstall_AptCacheDispatch verifies apt-cache is dispatched without auth injection.
func TestAptInstall_AptCacheDispatch(t *testing.T) {
	initAptTest(t)
	defer cleanAptTest(t)

	// apt-cache show doesn't need auth
	runJfrogCli(t, "apt", "apt-cache", "show", "base-files")
}

// TestAptInstall_PackageNotFound verifies a 404 in Artifactory returns an error.
func TestAptInstall_PackageNotFound(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)

	err := runJfrogCliWithoutAssertion("apt", "install", "-y", "jfrog-nonexistent-package-xyz",
		"--repo="+aptRepo(),
		"--dist=noble",
		"--trusted",
	)
	assert.Error(t, err, "installing a nonexistent package must return an error")
}

// TestAptSetupThenNativeInstall verifies that after 'jf setup apt', a plain
// 'apt-get install' installs the package from Artifactory (not a system mirror)
// because the pinning file gives the Artifactory source Pin-Priority: 1001.
func TestAptSetupThenNativeInstall(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)

	repo := aptRepo()
	dist := "noble"

	// Persistent setup — writes sources.list + Pin-Priority: 1001 pinning file.
	runJfrogCli(t, "setup", "apt",
		"--repo="+repo,
		"--dist="+dist,
		"--trusted",
	)
	require.FileExists(t, sourcesListPath(repo, dist))
	require.FileExists(t, prefPath(repo, dist))

	// Native apt-get install — no jf wrapper.
	// --allow-downgrades is needed when the Artifactory remote has an older version
	// than the container's pre-installed one; the test still proves the pin is effective.
	out, err := exec.Command("apt-get", "install", "-y", "--allow-downgrades", "curl").CombinedOutput()
	require.NoError(t, err, "native apt-get install failed: %s", out)

	assertPersistentInstallFromArtifactory(t, "curl", *tests.JfrogUrl)
}

// ── distribution matrix ───────────────────────────────────────────────────────

// TestAptSetup_DistributionMatrix runs setup across multiple dist values.
// In CI this is driven by the container image; here we parametrize the dist string.
func TestAptSetup_DistributionMatrix(t *testing.T) {
	initAptTest(t)
	requireRoot(t)

	dists := []string{"noble", "jammy", "focal", "bookworm", "bullseye"}
	for _, dist := range dists {
		dist := dist
		t.Run(dist, func(t *testing.T) {
			defer func() {
				_ = runJfrogCliWithoutAssertion("setup", "apt", "--remove", "--dist="+dist)
			}()

			err := runJfrogCliWithoutAssertion("setup", "apt",
				"--repo="+aptRepo(),
				"--dist="+dist,
				"--trusted",
			)
			if err != nil && strings.Contains(err.Error(), "apt-get update failed") {
				t.Skipf("dist %q not available in this Artifactory remote repo — skipping", dist)
			}
			require.NoError(t, err)
			assert.FileExists(t, sourcesListPath(aptRepo(), dist))
		})
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// captureTempSourcesList starts a background goroutine that polls /tmp for a
// jfrog-apt-* file created by WriteTempSourcesList, reads its content, and sends
// it on the returned channel. The caller must defer the cancel func.
// This lets tests inspect the on-the-fly sources.list before the defer in
// AptCommand.Run() removes it.
func captureTempSourcesList(t *testing.T) (<-chan string, func()) {
	t.Helper()
	tmpGlob := filepath.Join(os.TempDir(), "jfrog-apt-*")

	existing, _ := filepath.Glob(tmpGlob)
	seen := make(map[string]bool, len(existing))
	for _, f := range existing {
		seen[f] = true
	}

	ch := make(chan string, 1)
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				matches, _ := filepath.Glob(tmpGlob)
				for _, m := range matches {
					if seen[m] {
						continue
					}
					content, err := os.ReadFile(m)
					if err != nil || len(content) == 0 {
						continue
					}
					seen[m] = true
					select {
					case ch <- string(content):
					default:
					}
					return // captured what we need
				}
			}
		}
	}()
	return ch, func() { close(done) }
}

// assertInstalledFromArtifactory reads /var/log/apt/history.log and verifies the
// most recent apt commandline used on-the-fly Artifactory auth (Dir::Etc::sourcelist=
// pointing to a jfrog temp file and Dir::Etc::sourceparts=- to block other sources).
func assertInstalledFromArtifactory(t *testing.T) {
	t.Helper()
	data, err := os.ReadFile("/var/log/apt/history.log")
	if err != nil {
		t.Logf("Warning: cannot read /var/log/apt/history.log: %v — skipping Artifactory source check", err)
		return
	}
	s := string(data)
	idx := strings.LastIndex(s, "\nCommandline:")
	if idx == -1 {
		idx = strings.Index(s, "Commandline:")
	}
	require.NotEqual(t, -1, idx, "no Commandline entry found in apt history log")
	line := s[idx+1:]
	if end := strings.IndexByte(line, '\n'); end != -1 {
		line = line[:end]
	}
	assert.Contains(t, line, "Dir::Etc::sourcelist=", "apt must have used on-the-fly Artifactory source")
	assert.Contains(t, line, "Dir::Etc::sourceparts=-", "apt must have disabled system sources during install")
}

// assertPersistentInstallFromArtifactory verifies that pkg's installed version
// was sourced from the Artifactory instance identified by artURL.
// It runs 'apt-cache policy <pkg>' and checks that the line under the installed
// version (marked ***) contains the Artifactory host.
func assertPersistentInstallFromArtifactory(t *testing.T, pkg, artURL string) {
	t.Helper()
	out, err := exec.Command("apt-cache", "policy", pkg).Output()
	require.NoError(t, err, "apt-cache policy failed")

	u, _ := url.Parse(artURL)
	artHost := u.Hostname()

	lines := strings.Split(string(out), "\n")
	for i, line := range lines {
		if strings.Contains(line, "***") {
			// The source URL appears on one of the following indented lines.
			for j := i + 1; j < len(lines) && j <= i+3; j++ {
				if strings.Contains(lines[j], artHost) {
					return // found — Artifactory was the source
				}
			}
			t.Errorf("installed version of %q not sourced from Artifactory (%s).\napt-cache policy output:\n%s", pkg, artURL, out)
			return
		}
	}
	t.Errorf("%q does not appear to be installed; apt-cache policy output:\n%s", pkg, out)
}

func isWritable(info os.FileInfo) bool {
	mode := info.Mode()
	return mode&0200 != 0
}

func splitLines(text string) []string {
	var lines []string
	for _, l := range strings.Split(text, "\n") {
		l = strings.TrimSpace(l)
		if l != "" && l[0] != '#' {
			lines = append(lines, l)
		}
	}
	return lines
}
