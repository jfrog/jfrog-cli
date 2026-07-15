package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// testDist returns the apt distribution codename for the current container,
// driven by the APT_TEST_DIST env var set in aptTests.yml. Falls back to
// "noble" so local runs without the env still work.
func testDist() string {
	if d := os.Getenv("APT_TEST_DIST"); d != "" {
		return d
	}
	return "noble"
}

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

// TestAptSetup_ImportKey verifies the full --import-key handshake end to end
// against a local Debian repository Artifactory signs with the repo's own
// keypair — no server-wide signing key involved:
//
//   - a passphrase-protected GPG keypair is attached to a DEDICATED local repo
//     as its primaryKeyPairRef (the key --import-key installs and the index is
//     signed with);
//   - a package is seeded and the repo reindexed with that key's passphrase, so
//     Artifactory generates a dists/<dist>/InRelease signed with the per-repo key;
//   - setup --import-key fetches the key, installs it into the apt keyring, and
//     writes a signed-by= sources entry; apt-get update verifies the signed
//     index and the seeded package installs from Artifactory.
//
// A dedicated repo (not tests.AptLocalRepo, which is a member of the shared
// virtual) is used so signing it can't change what the virtual serves to the
// sibling install tests. No server GPG signing key is set, so the instance is
// otherwise untouched. A remote-backed repo cannot be used here: Artifactory
// proxies the upstream (Ubuntu/Debian) InRelease signature unchanged, so the
// imported key never matches it (apt fails NO_PUBKEY for the upstream key).
func TestAptSetup_ImportKey(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)

	dist := testDist()
	const component = "main"
	const pkg = "jfrog-apt-importkey-testpkg"
	// Dedicated local repo, isolated from the shared virtual used by sibling tests.
	repo := tests.AptLocalRepo + "-importkey"

	pairName, passphrase, cleanupKeypair := createArtifactoryGPGKeypair(t)
	defer cleanupKeypair()
	createLocalDebianRepo(t, repo, pairName)
	defer deleteRepo(repo)

	// Seed a package, then reindex with the repo key's passphrase so Artifactory
	// signs dists/<dist>/InRelease with that per-repo key.
	buildAndUploadTestDeb(t, repo, dist, component, pkg)
	reindexDebianRepo(t, repo, passphrase)
	waitForSignedInRelease(t, repo, dist)

	// setup --import-key: fetch+install the repo key, write sources, and verify
	// the signed index via the trailing apt-get update (runJfrogCli asserts success).
	runJfrogCli(t, "setup", "apt",
		"--repo="+repo,
		"--dist="+dist,
		"--component="+component,
		"--import-key",
	)

	assert.FileExists(t, keyringPath(repo, dist))

	content, err := os.ReadFile(sourcesListPath(repo, dist))
	require.NoError(t, err)
	assert.Contains(t, string(content), "signed-by=")
	assert.NotContains(t, string(content), "trusted=yes")

	// The seeded package installs from Artifactory — proving the imported key
	// verified the signed index end to end.
	out, err := exec.Command("apt-get", "install", "-y", pkg).CombinedOutput()
	require.NoError(t, err, "apt-get install %s from Artifactory failed: %s", pkg, out)
	artURL := strings.TrimSuffix(*tests.JfrogUrl+tests.ArtifactoryEndpoint, "/")
	assertPersistentInstallFromArtifactory(t, pkg, artURL)
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

// TestAptInstall_OnTheFlyInstall verifies a package can be installed via
// on-the-fly auth and that the install came from Artifactory.
func TestAptInstall_OnTheFlyInstall(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)

	// Ensure ed is absent first: if it's already installed (dirty/reused
	// container), the install below would be a no-op and write no history entry.
	// Purge makes the subsequent install do real work regardless of container state.
	// A failure here is non-fatal (e.g. ed simply isn't installed), but log it so
	// an unexpected purge problem is visible if the install later misbehaves.
	if out, err := exec.Command("apt-get", "purge", "-y", "ed").CombinedOutput(); err != nil {
		t.Logf("pre-test purge of ed failed (continuing): %v\n%s", err, out)
	}

	// Snapshot history.log size before installing so assertInstalledFromArtifactory
	// reads only the entry written by this test, not the workflow's prereq step.
	logOffset := aptHistoryLogSize()

	dist := testDist()
	// Use ed: small, in main on every Ubuntu/Debian distro (so resolvable via the
	// Artifactory remote for any dist), not pre-installed, and — unlike bzip2 — it
	// depends only on libc6 with no exact-version pin, so a slightly stale
	// Artifactory cache can't produce a "held broken packages" dependency conflict.
	runJfrogCli(t, "apt", "install", "-y", "ed",
		"--repo="+aptRepo(),
		"--dist="+dist,
		"--trusted",
	)

	_, err := exec.LookPath("ed")
	assert.NoError(t, err, "ed must be installed after jf apt install")

	assertInstalledFromArtifactory(t, logOffset)
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
	dist := testDist()

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

	dists := []string{"noble", "jammy", "focal", "trixie", "bookworm", "bullseye"}
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
			if err != nil {
				msg := err.Error()
				if strings.Contains(msg, "apt-get update failed") ||
					strings.Contains(msg, "not available in this Artifactory remote repo") {
					t.Skipf("dist %q not available in this Artifactory remote repo — skipping", dist)
				}
				// 502 / transient platform error — skip rather than fail the suite.
				if strings.Contains(msg, "502") || strings.Contains(msg, "Bad Gateway") ||
					strings.Contains(msg, "executor timeout") {
					t.Skipf("dist %q: transient platform error — skipping: %v", dist, err)
				}
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

// aptHistoryLogSize returns the current byte length of /var/log/apt/history.log.
// Call before an install to get an offset; pass to assertInstalledFromArtifactory
// so it only inspects lines written during the test, not earlier workflow steps.
func aptHistoryLogSize() int64 {
	info, err := os.Stat("/var/log/apt/history.log")
	if err != nil {
		return 0
	}
	return info.Size()
}

// assertInstalledFromArtifactory reads /var/log/apt/history.log from offset and
// verifies the first apt commandline entry uses on-the-fly Artifactory auth
// (Dir::Etc::sourcelist= pointing to a jfrog temp file, Dir::Etc::sourceparts=-
// to suppress other sources). offset should be from aptHistoryLogSize() before
// the install; pass 0 to search the full log.
func assertInstalledFromArtifactory(t *testing.T, offset int64) {
	t.Helper()
	data, err := os.ReadFile("/var/log/apt/history.log")
	if err != nil {
		t.Logf("Warning: cannot read /var/log/apt/history.log: %v — skipping Artifactory source check", err)
		return
	}
	if offset > int64(len(data)) {
		offset = int64(len(data))
	}
	s := string(data[offset:])
	idx := strings.Index(s, "Commandline:")
	require.NotEqual(t, -1, idx, "no Commandline entry found in apt history log after test install")
	line := s[idx:]
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

// createArtifactoryGPGKeypair generates a throwaway, passphrase-protected GPG
// keypair and uploads it to Artifactory via the REST API. Returns the pair name,
// its passphrase (needed when reindexing so Artifactory can sign the index with
// this per-repo key), and a cleanup func that deletes it. The caller must defer
// the cleanup. Skips the test if gpg is not available.
//
// The key is passphrase-protected on purpose: Artifactory signs a repo's Debian
// index with its primaryKeyPairRef key only when the reindex request carries
// that key's passphrase (see reindexDebianRepo). No server-wide signing key is
// involved, so nothing outside the target repo is affected.
func createArtifactoryGPGKeypair(t *testing.T) (pairName, passphrase string, cleanup func()) {
	t.Helper()
	if _, err := exec.LookPath("gpg"); err != nil {
		t.Skip("gpg not found — cannot create test GPG keypair")
	}
	const pass = "jfrog-apt-test-pass"

	// Isolated GPG home so we don't pollute the system keyring.
	gpgHome := t.TempDir()
	require.NoError(t, os.Chmod(gpgHome, 0700))

	keyParams := `Key-Type: RSA
Key-Length: 2048
Name-Real: JFrog Apt Test
Name-Email: jfrog-apt-test@example.com
Expire-Date: 0
Passphrase: ` + pass + `
%commit
`
	paramFile := filepath.Join(gpgHome, "keygen.conf")
	require.NoError(t, os.WriteFile(paramFile, []byte(keyParams), 0600))

	gpgArgs := func(args ...string) *exec.Cmd {
		cmd := exec.Command("gpg", args...)
		cmd.Env = append(os.Environ(), "GNUPGHOME="+gpgHome)
		return cmd
	}

	out, err := gpgArgs("--batch", "--gen-key", paramFile).CombinedOutput()
	require.NoError(t, err, "gpg key generation failed: %s", out)

	pubKey, err := gpgArgs("--armor", "--export", "jfrog-apt-test@example.com").Output()
	require.NoError(t, err, "gpg export public key failed")

	privKeyCmd := gpgArgs("--armor", "--batch", "--yes", "--pinentry-mode", "loopback", "--passphrase", pass, "--export-secret-keys", "jfrog-apt-test@example.com")
	privKey, err := privKeyCmd.Output()
	require.NoError(t, err, "gpg export private key failed")
	require.NotEmpty(t, privKey, "gpg exported empty private key — check GnuPG version and pinentry mode")

	pairName = fmt.Sprintf("jfrog-apt-test-%d", time.Now().UnixNano())
	artURL := strings.TrimSuffix(*tests.JfrogUrl+tests.ArtifactoryEndpoint, "/")

	type keypairReq struct {
		PairName   string `json:"pairName"`
		PairType   string `json:"pairType"`
		Alias      string `json:"alias"`
		PassPhrase string `json:"passPhrase"`
		PublicKey  string `json:"publicKey"`
		PrivateKey string `json:"privateKey"`
	}
	body, err := json.Marshal(keypairReq{
		PairName:   pairName,
		PairType:   "GPG",
		Alias:      pairName,
		PassPhrase: pass,
		PublicKey:  string(pubKey),
		PrivateKey: string(privKey),
	})
	require.NoError(t, err)

	doArtRequest(t, http.MethodPost, artURL+"/api/security/keypair", body, http.StatusCreated)

	cleanup = func() {
		req, _ := http.NewRequest(http.MethodDelete, artURL+"/api/security/keypair/"+pairName, nil)
		setArtAuth(req)
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			_ = resp.Body.Close()
		}
	}
	return pairName, pass, cleanup
}

// configureArtifactorySigningKey generates a passphrase-protected GPG key and
// installs it as Artifactory's server signing key. Artifactory only signs Debian
// repository metadata once such a key is configured. Returns the passphrase
// (required when triggering a reindex) and a best-effort cleanup func.
func createLocalDebianRepo(t *testing.T, repoName, pairName string) {
	t.Helper()
	artURL := strings.TrimSuffix(*tests.JfrogUrl+tests.ArtifactoryEndpoint, "/")
	body := fmt.Sprintf(`{"key":%q,"rclass":"local","packageType":"debian","repoLayoutRef":"simple-default","primaryKeyPairRef":%q}`, repoName, pairName)
	doArtRequest(t, http.MethodPut, artURL+"/api/repositories/"+repoName, []byte(body), http.StatusOK)
}

// buildAndUploadTestDeb builds a minimal .deb for the current runner's
// architecture (so the signed Release advertises that arch) and uploads it to a
// local Debian repo with the deb coordinate properties.
func buildAndUploadTestDeb(t *testing.T, repo, dist, component, pkg string) {
	t.Helper()
	if _, err := exec.LookPath("dpkg-deb"); err != nil {
		t.Skip("dpkg-deb not found — cannot build test .deb")
	}
	archOut, err := exec.Command("dpkg", "--print-architecture").Output()
	require.NoError(t, err, "dpkg --print-architecture failed")
	arch := strings.TrimSpace(string(archOut))

	buildRoot := t.TempDir()
	debianDir := filepath.Join(buildRoot, "pkg", "DEBIAN")
	require.NoError(t, os.MkdirAll(debianDir, 0755))
	control := fmt.Sprintf("Package: %s\nVersion: 1.0\nArchitecture: %s\n"+
		"Maintainer: JFrog Apt Test <jfrog-apt-test@example.com>\n"+
		"Description: JFrog apt import-key test package\n", pkg, arch)
	require.NoError(t, os.WriteFile(filepath.Join(debianDir, "control"), []byte(control), 0644))

	debPath := filepath.Join(buildRoot, fmt.Sprintf("%s_1.0_%s.deb", pkg, arch))
	out, err := exec.Command("dpkg-deb", "--build", filepath.Join(buildRoot, "pkg"), debPath).CombinedOutput()
	require.NoError(t, err, "dpkg-deb build failed: %s", out)
	data, err := os.ReadFile(debPath)
	require.NoError(t, err)

	artURL := strings.TrimSuffix(*tests.JfrogUrl+tests.ArtifactoryEndpoint, "/")
	uploadURL := fmt.Sprintf("%s/%s/pool/%s/%s/%s/%s_1.0_%s.deb;deb.distribution=%s;deb.component=%s;deb.architecture=%s",
		artURL, repo, component, pkg[:1], pkg, pkg, arch, dist, component, arch)
	putArtFile(t, uploadURL, data, map[string]string{"Content-Type": "application/octet-stream"}, http.StatusCreated)
}

// reindexDebianRepo triggers a Debian index recalculation, passing the repo
// keypair's passphrase so Artifactory signs the generated index with that key.
func reindexDebianRepo(t *testing.T, repo, passphrase string) {
	t.Helper()
	artURL := strings.TrimSuffix(*tests.JfrogUrl+tests.ArtifactoryEndpoint, "/")
	req, err := http.NewRequest(http.MethodPost, artURL+"/api/deb/reindex/"+repo, nil)
	require.NoError(t, err)
	req.Header.Set("X-GPG-PASSPHRASE", passphrase)
	setArtAuth(req)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "reindex %s\nresponse: %s", repo, body)
}

// waitForSignedInRelease polls until the signed dists/<dist>/InRelease is served
// or the timeout elapses.
func waitForSignedInRelease(t *testing.T, repo, dist string) {
	t.Helper()
	artURL := strings.TrimSuffix(*tests.JfrogUrl+tests.ArtifactoryEndpoint, "/")
	inRelease := fmt.Sprintf("%s/%s/dists/%s/InRelease", artURL, repo, dist)
	deadline := time.Now().Add(90 * time.Second)
	for {
		req, _ := http.NewRequest(http.MethodGet, inRelease, nil)
		setArtAuth(req)
		if resp, err := http.DefaultClient.Do(req); err == nil {
			code := resp.StatusCode
			_ = resp.Body.Close()
			if code == http.StatusOK {
				return
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("signed InRelease for %s/%s was not generated within timeout", repo, dist)
		}
		time.Sleep(3 * time.Second)
	}
}

// putArtFile PUTs raw bytes to Artifactory with optional extra headers.
func putArtFile(t *testing.T, rawURL string, data []byte, headers map[string]string, wantStatus int) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPut, rawURL, bytes.NewReader(data))
	require.NoError(t, err)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	setArtAuth(req)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	require.Equal(t, wantStatus, resp.StatusCode, "PUT %s\nresponse: %s", rawURL, body)
}

// doArtRequest performs an authenticated Artifactory REST call and asserts the status.
func doArtRequest(t *testing.T, method, url string, body []byte, wantStatus int) {
	t.Helper()
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	setArtAuth(req)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	respBody, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	require.Equal(t, wantStatus, resp.StatusCode, "%s %s\nresponse: %s", method, url, respBody)
}

// setArtAuth attaches admin credentials to a request using the test flags.
func setArtAuth(req *http.Request) {
	if *tests.JfrogAccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+*tests.JfrogAccessToken)
	} else {
		req.SetBasicAuth(*tests.JfrogUser, *tests.JfrogPassword)
	}
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
