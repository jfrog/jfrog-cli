package main

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RTECO-1596: a plain download must not include "property" in its main AQL query, because Artifactory then joins the
// item properties table, which can be very slow on large Oracle instances. The symlink properties are fetched by a
// separate query, so symlinks are still restored.

var aqlQueryPattern = regexp.MustCompile(`items\.find\(.*`)

const (
	symlinkTreePath  = "bi/public_keys"
	symlinkPropsAql  = `"@symlink.dest"`
	includePropsAql  = `"property"`
	symlinkTreeProps = "team=ecosystem;owner=cli"
)

// Runs the command with a DEBUG logger and returns the AQL queries it sent, in order.
func execAndCaptureAql(args ...string) ([]string, error) {
	previousLog := log.Logger
	defer log.SetLogger(previousLog)
	buffer := &bytes.Buffer{}
	debugLog := log.NewLogger(log.DEBUG, nil)
	debugLog.SetOutputWriter(buffer)
	debugLog.SetLogsWriter(buffer, 0)
	log.SetLogger(debugLog)

	err := artifactoryCli.Exec(args...)
	var queries []string
	for _, line := range strings.Split(buffer.String(), "\n") {
		if query := aqlQueryPattern.FindString(line); query != "" {
			queries = append(queries, strings.TrimSpace(query))
		}
	}
	return queries, err
}

// Runs the command from dir. Older clients resolve a relative symlink target against the working directory when
// validating it, newer ones against the link directory. Running from the link directory works with both.
func execAndCaptureAqlInDir(t *testing.T, dir string, args ...string) ([]string, error) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	defer func() {
		require.NoError(t, os.Chdir(wd))
	}()
	return execAndCaptureAql(args...)
}

// Uploads a tree like the one in RTECO-1596 to <repo>/bi/public_keys:
// key1.pem, key2.pem, archive/key3.pem, notes.txt (with properties), and 'link', a symlink to key1.pem.
// The symlink target is relative, because newer clients only restore symlinks that stay inside the download directory.
func uploadSymlinkTree(t *testing.T, extraArgs ...string) {
	localDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(localDir, "archive"), 0755))
	for name, content := range map[string]string{"key1.pem": "key1", "key2.pem": "key2", "archive/key3.pem": "key3", "notes.txt": "notes"} {
		require.NoError(t, os.WriteFile(filepath.Join(localDir, name), []byte(content), 0644))
	}
	require.NoError(t, os.Symlink("key1.pem", filepath.Join(localDir, "link")))

	target := tests.RtRepo1 + "/" + symlinkTreePath + "/"
	runRt(t, append([]string{"u", filepath.Join(localDir, "key*.pem"), target, "--flat"}, extraArgs...)...)
	runRt(t, append([]string{"u", filepath.Join(localDir, "archive", "key3.pem"), target + "archive/", "--flat"}, extraArgs...)...)
	runRt(t, append([]string{"u", filepath.Join(localDir, "notes.txt"), target, "--flat", "--target-props=" + symlinkTreeProps}, extraArgs...)...)
	runRt(t, append([]string{"u", filepath.Join(localDir, "link"), target, "--flat", "--symlinks"}, extraArgs...)...)
}

func assertFlatSymlinkTreeDownloaded(t *testing.T, outDir string) {
	for _, name := range []string{"key1.pem", "key2.pem", "key3.pem", "notes.txt"} {
		assert.FileExists(t, filepath.Join(outDir, name))
	}
	validateSymLink(filepath.Join(outDir, "link"), filepath.Join(outDir, "key1.pem"), t)
}

func TestDownloadMainAqlDoesNotIncludeProperties(t *testing.T) {
	if coreutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t, "")
	defer cleanArtifactoryTest()
	uploadSymlinkTree(t)
	pattern := tests.RtRepo1 + "/" + symlinkTreePath + "/"

	flatOut := filepath.Join(tests.Out, "flat") + string(filepath.Separator)
	flatQueries, err := execAndCaptureAql("dl", pattern, flatOut, "--flat")
	require.NoError(t, err)
	require.Len(t, flatQueries, 2, "expected the main query and the symlink properties query")
	assert.NotContains(t, flatQueries[0], includePropsAql, "the main query must not include the item properties")
	assert.Contains(t, flatQueries[1], symlinkPropsAql)
	assert.Contains(t, flatQueries[1], `"@symlink.destsha1"`)
	assertFlatSymlinkTreeDownloaded(t, flatOut)

	// --flat only changes where files are written locally, not the queries.
	treeOut := filepath.Join(tests.Out, "tree") + string(filepath.Separator)
	treeQueries, err := execAndCaptureAql("dl", pattern, treeOut)
	require.NoError(t, err)
	assert.Equal(t, flatQueries, treeQueries)
	assert.FileExists(t, filepath.Join(treeOut, symlinkTreePath, "archive", "key3.pem"))
	validateSymLink(filepath.Join(treeOut, symlinkTreePath, "link"), filepath.Join(treeOut, symlinkTreePath, "key1.pem"), t)
}

func TestDownloadValidateSymlinksWithSeparatePropsQuery(t *testing.T) {
	if coreutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t, "")
	defer cleanArtifactoryTest()
	uploadSymlinkTree(t)
	pattern := tests.RtRepo1 + "/" + symlinkTreePath + "/link"

	// Only the link is downloaded, so its target is created locally first.
	validOut := createDirWithFile(t, "valid", "key1.pem", "key1")
	queries, err := execAndCaptureAqlInDir(t, validOut, "dl", pattern, validOut+string(filepath.Separator), "--flat", "--validate-symlinks")
	require.NoError(t, err)
	require.Len(t, queries, 2)
	assert.NotContains(t, queries[0], includePropsAql)
	validateSymLink(filepath.Join(validOut, "link"), filepath.Join(validOut, "key1.pem"), t)

	// symlink.destsha1 still reaches the CLI: a changed target fails the validation.
	changedOut := createDirWithFile(t, "changed", "key1.pem", "changed")
	_, err = execAndCaptureAqlInDir(t, changedOut, "dl", pattern, changedOut+string(filepath.Separator), "--flat", "--validate-symlinks")
	assert.Error(t, err)
}

// Creates tests.Out/<dir> with one file and returns the directory's absolute path.
func createDirWithFile(t *testing.T, dir, fileName, content string) string {
	absDir, err := filepath.Abs(filepath.Join(tests.Out, dir))
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(absDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(absDir, fileName), []byte(content), 0644))
	return absDir
}

func TestDownloadByBuildKeepsPropertiesInMainAql(t *testing.T) {
	if coreutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t, "")
	buildNumber := "1596"
	defer func() {
		inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
		cleanArtifactoryTest()
	}()
	uploadSymlinkTree(t, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumber)
	runRt(t, "bp", tests.RtBuildName1, buildNumber)

	pattern := tests.RtRepo1 + "/" + symlinkTreePath + "/"
	build := "--build=" + tests.RtBuildName1 + "/" + buildNumber

	// Build search: not changed by RTECO-1596, the symlinks are still restored.
	buildOut := filepath.Join(tests.Out, "build") + string(filepath.Separator)
	_, err := execAndCaptureAql("dl", pattern, buildOut, "--flat", build)
	require.NoError(t, err)
	assertFlatSymlinkTreeDownloaded(t, buildOut)

	// A pattern with a build and a props filter is searched by pattern and then filtered by build. The build filter
	// reads the build properties of the results, so the main query keeps the properties and no symlink query is sent.
	filteredOut := filepath.Join(tests.Out, "filtered") + string(filepath.Separator)
	queries, err := execAndCaptureAql("dl", pattern, filteredOut, "--flat", build, "--exclude-props=team=ecosystem")
	require.NoError(t, err)
	assert.True(t, containsQuery(queries, includePropsAql), "expected a query that includes the item properties: %v", queries)
	assert.False(t, containsQuery(queries, symlinkPropsAql), "expected no symlink properties query: %v", queries)
	for _, name := range []string{"key1.pem", "key2.pem", "key3.pem"} {
		assert.FileExists(t, filepath.Join(filteredOut, name))
	}
	assert.NoFileExists(t, filepath.Join(filteredOut, "notes.txt"))
	validateSymLink(filepath.Join(filteredOut, "link"), filepath.Join(filteredOut, "key1.pem"), t)
}

func containsQuery(queries []string, part string) bool {
	for _, query := range queries {
		if strings.Contains(query, part) {
			return true
		}
	}
	return false
}

func TestDownloadWithPropsFilterAndSymlink(t *testing.T) {
	if coreutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t, "")
	defer cleanArtifactoryTest()
	uploadSymlinkTree(t)

	// tests.Out also holds the CLI home directory in CI, so download to a subdirectory.
	outDir := filepath.Join(tests.Out, "props")
	queries, err := execAndCaptureAql("dl", tests.RtRepo1+"/"+symlinkTreePath+"/", outDir+string(filepath.Separator), "--flat", "--props=team=ecosystem")
	require.NoError(t, err)
	require.Len(t, queries, 2)
	assert.Contains(t, queries[0], `"@team"`, "the props filter stays in the search criteria")
	assert.NotContains(t, queries[0], includePropsAql)

	entries, err := os.ReadDir(outDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "notes.txt", entries[0].Name())
}

func TestDownloadSpecWithSymlink(t *testing.T) {
	if coreutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t, "")
	defer cleanArtifactoryTest()
	uploadSymlinkTree(t)

	out, err := filepath.Abs(tests.Out)
	require.NoError(t, err)
	specFile := filepath.Join(t.TempDir(), "download-spec.json")
	spec := `{"files":[{"pattern":"` + tests.RtRepo1 + "/" + symlinkTreePath + `/*","target":"` + filepath.ToSlash(out) + `/","flat":"true"}]}`
	require.NoError(t, os.WriteFile(specFile, []byte(spec), 0644))

	queries, err := execAndCaptureAql("dl", "--spec="+specFile)
	require.NoError(t, err)
	require.Len(t, queries, 2)
	assert.NotContains(t, queries[0], includePropsAql)
	assert.Contains(t, queries[1], symlinkPropsAql)
	assertFlatSymlinkTreeDownloaded(t, out)
}

func TestDownloadNoMatchesWithSeparatePropsQuery(t *testing.T) {
	initArtifactoryTest(t, "")
	defer cleanArtifactoryTest()

	outDir := filepath.Join(tests.Out, "none")
	queries, err := execAndCaptureAql("dl", tests.RtRepo1+"/does/not/exist/", outDir+string(filepath.Separator), "--flat")
	require.NoError(t, err)
	require.NotEmpty(t, queries)
	assert.NotContains(t, queries[0], includePropsAql)
	assert.NoDirExists(t, outDir)
}
