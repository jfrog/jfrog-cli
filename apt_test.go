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

	buildinfo "github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/jfrog-cli-core/v2/common/build"
	"github.com/jfrog/jfrog-cli/inttestutils"
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
// imported key never matches it (apt fails NO_PUBKEY for the upstream key) —
// this holds even when the remote is wrapped in a virtual, which passes the
// upstream signature through. See TestAptInstall_PersistentConfigVirtualRemote,
// which covers the virtual+remote (no local) path with --trusted instead.
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

// TestAptInstall_PersistentConfigVirtualRemote mirrors the manual repro end to
// end for the "virtual repo backed only by a remote, no local member" case:
// persistent `jf setup apt` against that virtual, then a NO-FLAGS `jf apt
// install` that must detect the persistent config setup just wrote and install a
// remote-backed package using the embedded credentials.
//
// --trusted (not --import-key) is used deliberately: a virtual backed by a
// remote proxies the upstream (Ubuntu/Debian) InRelease signature unchanged — it
// does not re-sign with an Artifactory key — so an imported Artifactory key can
// never verify it (apt fails NO_PUBKEY for the upstream key). --import-key
// against an Artifactory-signed local repo is covered by TestAptSetup_ImportKey.
//
// Steps:
//   - create a remote proxying the arch-correct upstream and a remote-only virtual;
//   - `jf setup apt --trusted` — persistent [trusted=yes] source + embedded creds,
//     verified by the trailing apt-get update (runJfrogCli asserts success);
//   - `jf apt install <pkg>` with NO --repo/--dist — must log "Using persistent
//     Artifactory apt configuration" and install the remote-backed package,
//     sourced from Artifactory (Pin-Priority 1001).
func TestAptInstall_PersistentConfigVirtualRemote(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)

	dist := testDist()
	const component = "main"
	// ed: small, in main on every Ubuntu/Debian dist, not pre-installed, and
	// depends only on libc6 with no exact-version pin (no downgrade conflicts).
	const pkg = "ed"

	// Short, self-contained repo keys — Artifactory caps repo names at 58 chars,
	// so do NOT derive from the already-timestamped global repo vars. dist keeps
	// concurrent matrix containers from colliding; the hex tail adds uniqueness.
	uniq := fmt.Sprintf("%x", time.Now().UnixNano()&0xffffff)
	remoteRepo := fmt.Sprintf("cli-apt-pcvr-%s-r-%s", dist, uniq)
	virtualRepo := fmt.Sprintf("cli-apt-pcvr-%s-v-%s", dist, uniq)

	upstreamURL, arch := aptRemoteUpstream(t, dist)
	createRemoteDebianRepo(t, remoteRepo, upstreamURL, arch)
	defer deleteRepo(remoteRepo)

	createVirtualDebianRepo(t, virtualRepo, remoteRepo)
	defer deleteRepo(virtualRepo) // registered last → deleted first (before its remote member)

	// Ensure pkg is absent so the install below does real work (a reused container
	// may already have it). Non-fatal if it isn't installed.
	if out, err := exec.Command("apt-get", "purge", "-y", pkg).CombinedOutput(); err != nil {
		t.Logf("pre-test purge of %s failed (continuing): %v\n%s", pkg, err, out)
	}

	// Persistent setup against the remote-only virtual.
	runJfrogCli(t, "setup", "apt",
		"--repo="+virtualRepo,
		"--dist="+dist,
		"--component="+component,
		"--trusted",
	)

	sourcesFile := sourcesListPath(virtualRepo, dist)
	require.FileExists(t, sourcesFile)
	require.FileExists(t, prefPath(virtualRepo, dist))
	content, err := os.ReadFile(sourcesFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "[trusted=yes]")
	assert.Contains(t, string(content), virtualRepo)

	// The crux: no --repo / no --dist. Must detect the persistent config written
	// above and install the remote-backed package via its embedded credentials.
	runJfrogCli(t, "apt", "install", "-y", pkg)

	_, err = exec.LookPath(pkg)
	assert.NoError(t, err, "%s must be installed via persistent-config 'jf apt install'", pkg)
	assertPersistentInstallFromArtifactory(t, pkg, *tests.JfrogUrl)
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

// ── auth + setup with existing virtual repo ───────────────────────────────────

// TestAptInstall_PersistentSetupJfAptInstall verifies the "setup once, then
// jf apt install without flags" flow against the pre-configured virtual repository
// (cli-apt-virtual, backed by ubuntu-remote + debian-remote + local members).
//
// After `jf setup apt` writes a persistent sources.list with embedded credentials
// and a Pin-Priority: 1001 preferences file, a bare `jf apt install` (no --repo /
// --dist) must detect the persistent config, route the install through Artifactory,
// and succeed.
func TestAptInstall_PersistentSetupJfAptInstall(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)

	repo := aptRepo()
	dist := testDist()
	const pkg = "ed"

	if out, err := exec.Command("apt-get", "purge", "-y", pkg).CombinedOutput(); err != nil {
		t.Logf("pre-test purge of %s skipped: %v\n%s", pkg, err, out)
	}

	// Write persistent sources.list + pinning file for the shared virtual repo.
	runJfrogCli(t, "setup", "apt",
		"--repo="+repo,
		"--dist="+dist,
		"--component=main",
		"--trusted",
	)

	require.FileExists(t, sourcesListPath(repo, dist))
	require.FileExists(t, prefPath(repo, dist))

	// Install WITHOUT --repo/--dist. Must log "Using persistent Artifactory apt
	// configuration" and proxy the install through the configured virtual repo.
	runJfrogCli(t, "apt", "install", "-y", pkg)

	_, err := exec.LookPath(pkg)
	assert.NoError(t, err, "%s must be installed via persistent-config 'jf apt install'", pkg)
	assertPersistentInstallFromArtifactory(t, pkg, *tests.JfrogUrl)
}

// TestAptInstall_BuildInfoWithVirtualRepo verifies end-to-end build-info collection
// against the pre-configured virtual repository (cli-apt-virtual).
//
// Flow:
//  1. jf apt install jq with --build-name/--build-number; build-info is saved locally.
//  2. Local build-info is validated: module type debian, deps type deb, SHA256 present.
//  3. jf rt bp publishes the build-info to Artifactory.
//  4. The published build-info is fetched and round-trip-validated.
func TestAptInstall_BuildInfoWithVirtualRepo(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.AptBuildName, artHttpDetails)

	repo := aptRepo()
	dist := testDist()
	const buildNumber = "1"
	// jq: small utility whose closure is jq + libjq1 + libonig5; all three are served
	// by the ubuntu/debian remotes in the virtual repo, so all three resolve with
	// checksums from the Packages index and appear in the build-info output.
	const pkg = "jq"

	if out, err := exec.Command("apt-get", "purge", "-y", pkg).CombinedOutput(); err != nil {
		t.Logf("pre-test purge of %s skipped: %v\n%s", pkg, err, out)
	}

	runJfrogCli(t, "apt", "install", "-y", pkg,
		"--repo="+repo,
		"--dist="+dist,
		"--trusted",
		"--build-name="+tests.AptBuildName,
		"--build-number="+buildNumber,
	)

	_, err := exec.LookPath(pkg)
	require.NoError(t, err, "%s must be installed after jf apt install", pkg)

	// Validate the locally persisted build-info before it is published.
	validateAptLocalBuildInfo(t, tests.AptBuildName, buildNumber)

	// Publish and validate the round-trip through Artifactory.
	runJfrogCli(t, "rt", "bp", tests.AptBuildName, buildNumber)

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.AptBuildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found, "build %s/%s must appear in Artifactory after bp", tests.AptBuildName, buildNumber)

	bi := publishedBuildInfo.BuildInfo
	require.Len(t, bi.Modules, 1, "apt build-info must have exactly one module")
	mod := bi.Modules[0]
	assert.Equal(t, string(buildinfo.Debian), string(mod.Type))
	assert.NotEmpty(t, mod.Dependencies, "apt build-info must have at least one dependency")
	for _, dep := range mod.Dependencies {
		assert.Equal(t, "deb", dep.Type, "dep %s must have type deb", dep.Id)
		assert.NotEmpty(t, dep.Sha256, "dep %s must have SHA256 checksum", dep.Id)
	}
}

// validateAptLocalBuildInfo validates the build-info persisted locally after
// `jf apt install --build-name/--build-number` before it is published to Artifactory.
func validateAptLocalBuildInfo(t *testing.T, buildName, buildNumber string) {
	t.Helper()
	buildInfoService := build.CreateBuildInfoService()
	aptBuild, err := buildInfoService.GetOrCreateBuildWithProject(buildName, buildNumber, "")
	require.NoError(t, err)
	bi, err := aptBuild.ToBuildInfo()
	require.NoError(t, err)
	require.NotEmpty(t, bi.Started)
	if !assert.Len(t, bi.Modules, 1, "apt build-info must have exactly one module") {
		return
	}
	mod := bi.Modules[0]
	assert.Equal(t, string(buildinfo.Debian), string(mod.Type))
	assert.NotEmpty(t, mod.Dependencies, "apt build-info must have at least one dependency")
	for _, dep := range mod.Dependencies {
		assert.Equal(t, "deb", dep.Type, "dep %s must have type deb", dep.Id)
		assert.NotEmpty(t, dep.Sha256, "dep %s must have SHA256 checksum", dep.Id)
		assert.NotEmpty(t, dep.Scopes, "dep %s must have at least one scope", dep.Id)
	}
}

// ── build info dep properties ─────────────────────────────────────────────────

// TestAptInstall_BuildInfoDepProperties verifies scenarios #42–#45:
//   - #42: dependency ID format is name:version:arch
//   - #43: Depends/Pre-Depends → required, Recommends → recommended
//   - #44: sha256/sha1/md5 populated from Packages index
//   - #45: requestedBy chains are acyclic (no package in its own ancestry)
func TestAptInstall_BuildInfoDepProperties(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)
	defer func() { _ = build.RemoveBuildDir(tests.AptBuildName, "2", "") }()

	if out, err := exec.Command("apt-get", "purge", "-y", "jq").CombinedOutput(); err != nil {
		t.Logf("purge jq: %v\n%s", err, out)
	}

	runJfrogCli(t, "apt", "install", "-y", "jq",
		"--repo="+aptRepo(),
		"--dist="+testDist(),
		"--trusted",
		"--build-name="+tests.AptBuildName,
		"--build-number=2",
	)

	buildInfoService := build.CreateBuildInfoService()
	aptBuild, err := buildInfoService.GetOrCreateBuildWithProject(tests.AptBuildName, "2", "")
	require.NoError(t, err)
	bi, err := aptBuild.ToBuildInfo()
	require.NoError(t, err)
	require.Len(t, bi.Modules, 1)

	for _, dep := range bi.Modules[0].Dependencies {
		// #42: ID must be name:version:arch
		parts := strings.SplitN(dep.Id, ":", 3)
		assert.Len(t, parts, 3, "dep %s: ID must be name:version:arch", dep.Id)
		assert.NotEmpty(t, parts[0], "dep %s: name must not be empty", dep.Id)
		assert.NotEmpty(t, parts[1], "dep %s: version must not be empty", dep.Id)
		assert.NotEmpty(t, parts[2], "dep %s: arch must not be empty", dep.Id)

		// type must be deb
		assert.Equal(t, "deb", dep.Type, "dep %s: type must be deb", dep.Id)

		// #43: scope must be one of the three valid values
		for _, scope := range dep.Scopes {
			assert.Contains(t, []string{"required", "recommended", "optional"}, scope,
				"dep %s: unexpected scope %q", dep.Id, scope)
		}

		// #44: sha256 must be populated
		assert.NotEmpty(t, dep.Sha256, "dep %s: sha256 must be populated", dep.Id)

		// #45: package must not appear in its own requestedBy ancestry
		for _, path := range dep.RequestedBy {
			assert.NotContains(t, path, dep.Id,
				"dep %s must not appear in its own requestedBy chain: %v", dep.Id, path)
		}
	}
}

// ── build info flag combinations ──────────────────────────────────────────────

// TestAptInstall_BuildModule verifies scenario #35: --module overrides module ID.
func TestAptInstall_BuildModule(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)
	defer func() { _ = build.RemoveBuildDir(tests.AptBuildName, "3", "") }()

	const moduleID = "my-custom-apt-module"

	if out, err := exec.Command("apt-get", "purge", "-y", "jq").CombinedOutput(); err != nil {
		t.Logf("purge jq: %v\n%s", err, out)
	}

	runJfrogCli(t, "apt", "install", "-y", "jq",
		"--repo="+aptRepo(),
		"--dist="+testDist(),
		"--trusted",
		"--build-name="+tests.AptBuildName,
		"--build-number=3",
		"--module="+moduleID,
	)

	buildInfoService := build.CreateBuildInfoService()
	aptBuild, err := buildInfoService.GetOrCreateBuildWithProject(tests.AptBuildName, "3", "")
	require.NoError(t, err)
	bi, err := aptBuild.ToBuildInfo()
	require.NoError(t, err)
	require.Len(t, bi.Modules, 1)
	assert.Equal(t, moduleID, bi.Modules[0].Id, "--module must override the default module ID")
}

// TestAptInstall_BuildNameOnlyError verifies scenario #36:
// --build-name without --build-number → CLI rejects the partial flags with an error.
// JFrog CLI enforces that both flags must be provided together; a partial pair is
// an error rather than silently skipping build info collection.
func TestAptInstall_BuildNameOnlyError(t *testing.T) {
	initAptTest(t)
	defer cleanAptTest(t)

	err := runJfrogCliWithoutAssertion("apt", "install", "--dry-run", "-y", "jq",
		"--repo="+aptRepo(),
		"--dist="+testDist(),
		"--trusted",
		"--build-name=cli-apt-nameonly-test",
		// no --build-number
	)
	assert.Error(t, err, "--build-name without --build-number must return an error")
}

// TestAptInstall_BuildNumberOnlyError verifies scenario #37:
// --build-number without --build-name → CLI rejects with an error.
func TestAptInstall_BuildNumberOnlyError(t *testing.T) {
	initAptTest(t)
	defer cleanAptTest(t)

	err := runJfrogCliWithoutAssertion("apt", "install", "--dry-run", "-y", "jq",
		"--repo="+aptRepo(),
		"--dist="+testDist(),
		"--trusted",
		"--build-number=1",
		// no --build-name
	)
	assert.Error(t, err, "--build-number without --build-name must return an error")
}

// TestAptInstall_NoBuildFlagsNoBuildInfo verifies scenario #38:
// no build flags → install succeeds, no build info produced.
func TestAptInstall_NoBuildFlagsNoBuildInfo(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)

	const buildName = "cli-apt-nobuild-test"
	const buildNum = "1"

	runJfrogCli(t, "apt", "install", "--dry-run", "-y", "jq",
		"--repo="+aptRepo(),
		"--dist="+testDist(),
		"--trusted",
		// no --build-name, no --build-number
	)

	buildInfoService := build.CreateBuildInfoService()
	aptBuild, err := buildInfoService.GetOrCreateBuildWithProject(buildName, buildNum, "")
	require.NoError(t, err)
	bi, err := aptBuild.ToBuildInfo()
	require.NoError(t, err)
	assert.Empty(t, bi.Modules, "no build flags must not produce build info")
}

// TestAptInstall_BuildFlagsFromEnvVars verifies scenario #39:
// JFROG_CLI_BUILD_NAME + JFROG_CLI_BUILD_NUMBER env vars → build info captured.
func TestAptInstall_BuildFlagsFromEnvVars(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)
	defer func() { _ = build.RemoveBuildDir(tests.AptBuildName, "4", "") }()

	t.Setenv("JFROG_CLI_BUILD_NAME", tests.AptBuildName)
	t.Setenv("JFROG_CLI_BUILD_NUMBER", "4")

	if out, err := exec.Command("apt-get", "purge", "-y", "jq").CombinedOutput(); err != nil {
		t.Logf("purge jq: %v\n%s", err, out)
	}

	// No --build-name / --build-number flags; env vars must supply them.
	runJfrogCli(t, "apt", "install", "-y", "jq",
		"--repo="+aptRepo(),
		"--dist="+testDist(),
		"--trusted",
	)

	buildInfoService := build.CreateBuildInfoService()
	aptBuild, err := buildInfoService.GetOrCreateBuildWithProject(tests.AptBuildName, "4", "")
	require.NoError(t, err)
	bi, err := aptBuild.ToBuildInfo()
	require.NoError(t, err)
	require.Len(t, bi.Modules, 1, "build info must be captured from env vars")
	assert.NotEmpty(t, bi.Modules[0].Dependencies)
}

// ── dispatch ──────────────────────────────────────────────────────────────────

// TestAptInstall_DpkgQueryDispatch verifies scenario #29:
// jf apt dpkg-query dispatches to dpkg-query without auth injection.
func TestAptInstall_DpkgQueryDispatch(t *testing.T) {
	initAptTest(t)
	defer cleanAptTest(t)

	// base-files is always installed; dpkg-query must find it.
	runJfrogCli(t, "apt", "dpkg-query", "-W",
		"-f=${Package}\\t${Version}\\n", "base-files")
}

// ── closure bounded ───────────────────────────────────────────────────────────

// TestAptInstall_ClosureBounded verifies scenario #71:
// apt-cache closure is bounded to installed packages via --installed --no-suggests.
// curl's full archive closure is ~23,000 packages; with bounding it is <20.
func TestAptInstall_ClosureBounded(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)
	defer func() { _ = build.RemoveBuildDir(tests.AptBuildName, "5", "") }()

	if out, err := exec.Command("apt-get", "purge", "-y", "curl").CombinedOutput(); err != nil {
		t.Logf("purge curl: %v\n%s", err, out)
	}

	runJfrogCli(t, "apt", "install", "-y", "--allow-downgrades", "curl",
		"--repo="+aptRepo(),
		"--dist="+testDist(),
		"--trusted",
		"--build-name="+tests.AptBuildName,
		"--build-number=5",
	)

	buildInfoService := build.CreateBuildInfoService()
	aptBuild, err := buildInfoService.GetOrCreateBuildWithProject(tests.AptBuildName, "5", "")
	require.NoError(t, err)
	bi, err := aptBuild.ToBuildInfo()
	require.NoError(t, err)
	require.Len(t, bi.Modules, 1)

	depCount := len(bi.Modules[0].Dependencies)
	// On a minimal Ubuntu 24.04 image with --installed --no-suggests, curl's
	// closure is typically 2–6 packages, never thousands.
	assert.Less(t, depCount, 20,
		"curl dep count %d exceeds expected bound; --installed --no-suggests must be active", depCount)
	assert.Greater(t, depCount, 0, "curl must have at least one dependency")
}

// ── Artifactory unreachable ───────────────────────────────────────────────────

// TestAptInstall_ArtifactoryUnreachable verifies scenario #66:
// on-the-fly install against a nonexistent repo returns a clear error;
// apt must not silently fall back to system sources (Dir::Etc::sourceparts=-).
func TestAptInstall_ArtifactoryUnreachable(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)

	err := runJfrogCliWithoutAssertion("apt", "install", "-y", "curl",
		"--repo=repo-does-not-exist-xyz-abc",
		"--dist="+testDist(),
		"--trusted",
	)
	require.Error(t, err,
		"install against nonexistent repo must fail (no silent fallback to system sources)")
}

// ── full CI pipeline ─────────────────────────────────────────────────────────

// TestAptInstall_FullPipeline verifies scenario #65:
// jf setup apt → jf apt install with build flags → jf rt bp → build in Artifactory.
func TestAptInstall_FullPipeline(t *testing.T) {
	initAptTest(t)
	requireRoot(t)
	defer cleanAptTest(t)
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.AptBuildName, artHttpDetails)

	const buildNumber = "6"
	const pkg = "jq"
	dist := testDist()

	// 1. Persistent setup against the shared virtual repo.
	runJfrogCli(t, "setup", "apt",
		"--repo="+aptRepo(),
		"--dist="+dist,
		"--component=main",
		"--trusted",
	)
	require.FileExists(t, sourcesListPath(aptRepo(), dist))

	if out, err := exec.Command("apt-get", "purge", "-y", pkg).CombinedOutput(); err != nil {
		t.Logf("purge %s: %v\n%s", pkg, err, out)
	}

	// 2. Install using persistent config (no --repo/--dist needed).
	runJfrogCli(t, "apt", "install", "-y", pkg,
		"--build-name="+tests.AptBuildName,
		"--build-number="+buildNumber,
	)

	_, err := exec.LookPath(pkg)
	require.NoError(t, err, "%s must be installed", pkg)

	// 3. Validate local build info.
	validateAptLocalBuildInfo(t, tests.AptBuildName, buildNumber)

	// 4. Publish to Artifactory.
	runJfrogCli(t, "rt", "bp", tests.AptBuildName, buildNumber)

	// 5. Verify round-trip.
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.AptBuildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found, "build info must be in Artifactory after jf rt bp")
	require.Len(t, publishedBuildInfo.BuildInfo.Modules, 1)
	assert.Equal(t, string(buildinfo.Debian), string(publishedBuildInfo.BuildInfo.Modules[0].Type))
	assert.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules[0].Dependencies)
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
	const pass = "jfrog-apt-test-pass" //#nosec G101 -- test-only passphrase for a throwaway GPG keypair, not a real credential

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

// createRemoteDebianRepo creates a remote Debian repo proxying upstreamURL, with
// the runner's architecture added to the indexed architectures so binary-<arch>
// Packages are served.
func createRemoteDebianRepo(t *testing.T, repoName, upstreamURL, arch string) {
	t.Helper()
	artURL := strings.TrimSuffix(*tests.JfrogUrl+tests.ArtifactoryEndpoint, "/")
	body := fmt.Sprintf(`{"key":%q,"rclass":"remote","packageType":"debian","url":%q,"debianDefaultArchitectures":%q,"listRemoteFolderItems":true}`,
		repoName, upstreamURL, arch)
	doArtRequest(t, http.MethodPut, artURL+"/api/repositories/"+repoName, []byte(body), http.StatusOK)
}

// createVirtualDebianRepo creates a virtual Debian repo aggregating members.
// Note: a virtual backed by a remote proxies the upstream (Ubuntu/Debian)
// InRelease signature unchanged — it does NOT re-sign with an Artifactory key —
// so callers verifying via --import-key won't work here; use --trusted.
func createVirtualDebianRepo(t *testing.T, repoName string, members ...string) {
	t.Helper()
	artURL := strings.TrimSuffix(*tests.JfrogUrl+tests.ArtifactoryEndpoint, "/")
	membersJSON, err := json.Marshal(members)
	require.NoError(t, err)
	body := fmt.Sprintf(`{"key":%q,"rclass":"virtual","packageType":"debian","repositories":%s,"debianDefaultArchitectures":"amd64,arm64"}`,
		repoName, membersJSON)
	doArtRequest(t, http.MethodPut, artURL+"/api/repositories/"+repoName, []byte(body), http.StatusOK)
}

// aptRemoteUpstream returns an upstream apt mirror URL + the runner architecture
// appropriate for dist. Ubuntu dists use archive.ubuntu.com on amd64/i386 and
// ports.ubuntu.com/ubuntu-ports on other arches (arm64 etc.); Debian dists use
// deb.debian.org, which serves every architecture from one URL.
func aptRemoteUpstream(t *testing.T, dist string) (upstreamURL, arch string) {
	t.Helper()
	archOut, err := exec.Command("dpkg", "--print-architecture").Output()
	require.NoError(t, err, "dpkg --print-architecture failed")
	arch = strings.TrimSpace(string(archOut))

	switch dist {
	case "bookworm", "bullseye", "trixie", "buster", "sid":
		return "http://deb.debian.org/debian", arch
	default: // ubuntu codenames
		if arch == "amd64" || arch == "i386" {
			return "http://archive.ubuntu.com/ubuntu", arch
		}
		return "http://ports.ubuntu.com/ubuntu-ports", arch
	}
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
