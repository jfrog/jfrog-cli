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
func execAndCaptureAql(t *testing.T, args ...string) ([]string, error) {
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

// Uploads a tree like the one in RTECO-1596 to <repo>/bi/public_keys:
// key1.pem, key2.pem, archive/key3.pem, notes.txt (with properties), and 'link', a symlink to key1.pem.
// Returns the local path of key1.pem, the target of the uploaded symlink.
func uploadSymlinkTree(t *testing.T, extraArgs ...string) string {
	localDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(localDir, "archive"), 0755))
	for name, content := range map[string]string{"key1.pem": "key1", "key2.pem": "key2", "archive/key3.pem": "key3", "notes.txt": "notes"} {
		require.NoError(t, os.WriteFile(filepath.Join(localDir, name), []byte(content), 0644))
	}
	linkTarget := filepath.Join(localDir, "key1.pem")
	require.NoError(t, os.Symlink(linkTarget, filepath.Join(localDir, "link")))

	target := tests.RtRepo1 + "/" + symlinkTreePath + "/"
	runRt(t, append([]string{"u", filepath.Join(localDir, "key*.pem"), target, "--flat"}, extraArgs...)...)
	runRt(t, append([]string{"u", filepath.Join(localDir, "archive", "key3.pem"), target + "archive/", "--flat"}, extraArgs...)...)
	runRt(t, append([]string{"u", filepath.Join(localDir, "notes.txt"), target, "--flat", "--target-props=" + symlinkTreeProps}, extraArgs...)...)
	runRt(t, append([]string{"u", filepath.Join(localDir, "link"), target, "--flat", "--symlinks"}, extraArgs...)...)
	return linkTarget
}

func assertFlatSymlinkTreeDownloaded(t *testing.T, outDir, linkTarget string) {
	for _, name := range []string{"key1.pem", "key2.pem", "key3.pem", "notes.txt"} {
		assert.FileExists(t, filepath.Join(outDir, name))
	}
	validateSymLink(filepath.Join(outDir, "link"), linkTarget, t)
}

func TestDownloadMainAqlDoesNotIncludeProperties(t *testing.T) {
	if coreutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t, "")
	defer cleanArtifactoryTest()
	linkTarget := uploadSymlinkTree(t)
	pattern := tests.RtRepo1 + "/" + symlinkTreePath + "/"

	flatOut := filepath.Join(tests.Out, "flat") + string(filepath.Separator)
	flatQueries, err := execAndCaptureAql(t, "dl", pattern, flatOut, "--flat")
	require.NoError(t, err)
	require.Len(t, flatQueries, 2, "expected the main query and the symlink properties query")
	assert.NotContains(t, flatQueries[0], includePropsAql, "the main query must not include the item properties")
	assert.Contains(t, flatQueries[1], symlinkPropsAql)
	assert.Contains(t, flatQueries[1], `"@symlink.destsha1"`)
	assertFlatSymlinkTreeDownloaded(t, flatOut, linkTarget)

	// --flat only changes where files are written locally, not the queries.
	treeOut := filepath.Join(tests.Out, "tree") + string(filepath.Separator)
	treeQueries, err := execAndCaptureAql(t, "dl", pattern, treeOut)
	require.NoError(t, err)
	assert.Equal(t, flatQueries, treeQueries)
	assert.FileExists(t, filepath.Join(treeOut, symlinkTreePath, "archive", "key3.pem"))
	validateSymLink(filepath.Join(treeOut, symlinkTreePath, "link"), linkTarget, t)
}

func TestDownloadValidateSymlinksWithSeparatePropsQuery(t *testing.T) {
	if coreutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t, "")
	defer cleanArtifactoryTest()
	linkTarget := uploadSymlinkTree(t)
	pattern := tests.RtRepo1 + "/" + symlinkTreePath + "/link"

	validOut := filepath.Join(tests.Out, "valid") + string(filepath.Separator)
	queries, err := execAndCaptureAql(t, "dl", pattern, validOut, "--flat", "--validate-symlinks")
	require.NoError(t, err)
	require.Len(t, queries, 2)
	assert.NotContains(t, queries[0], includePropsAql)
	validateSymLink(filepath.Join(validOut, "link"), linkTarget, t)

	// symlink.destsha1 still reaches the CLI: a changed target fails the validation.
	require.NoError(t, os.WriteFile(linkTarget, []byte("changed"), 0644))
	changedOut := filepath.Join(tests.Out, "changed") + string(filepath.Separator)
	assert.Error(t, artifactoryCli.Exec("dl", pattern, changedOut, "--flat", "--validate-symlinks"))
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
	linkTarget := uploadSymlinkTree(t, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumber)
	runRt(t, "bp", tests.RtBuildName1, buildNumber)

	pattern := tests.RtRepo1 + "/" + symlinkTreePath + "/"
	build := "--build=" + tests.RtBuildName1 + "/" + buildNumber

	// Build search: not changed by RTECO-1596, the symlinks are still restored.
	buildOut := filepath.Join(tests.Out, "build") + string(filepath.Separator)
	_, err := execAndCaptureAql(t, "dl", pattern, buildOut, "--flat", build)
	require.NoError(t, err)
	assertFlatSymlinkTreeDownloaded(t, buildOut, linkTarget)

	// A pattern with a build and a props filter is searched by pattern and then filtered by build. The build filter
	// reads the build properties of the results, so the main query keeps the properties and no symlink query is sent.
	filteredOut := filepath.Join(tests.Out, "filtered") + string(filepath.Separator)
	queries, err := execAndCaptureAql(t, "dl", pattern, filteredOut, "--flat", build, "--exclude-props=team=ecosystem")
	require.NoError(t, err)
	assert.True(t, containsQuery(queries, includePropsAql), "expected a query that includes the item properties: %v", queries)
	assert.False(t, containsQuery(queries, symlinkPropsAql), "expected no symlink properties query: %v", queries)
	for _, name := range []string{"key1.pem", "key2.pem", "key3.pem"} {
		assert.FileExists(t, filepath.Join(filteredOut, name))
	}
	assert.NoFileExists(t, filepath.Join(filteredOut, "notes.txt"))
	validateSymLink(filepath.Join(filteredOut, "link"), linkTarget, t)
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

	out := tests.Out + string(filepath.Separator)
	queries, err := execAndCaptureAql(t, "dl", tests.RtRepo1+"/"+symlinkTreePath+"/", out, "--flat", "--props=team=ecosystem")
	require.NoError(t, err)
	require.Len(t, queries, 2)
	assert.Contains(t, queries[0], `"@team"`, "the props filter stays in the search criteria")
	assert.NotContains(t, queries[0], includePropsAql)

	entries, err := os.ReadDir(tests.Out)
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
	linkTarget := uploadSymlinkTree(t)

	out, err := filepath.Abs(tests.Out)
	require.NoError(t, err)
	specFile := filepath.Join(t.TempDir(), "download-spec.json")
	spec := `{"files":[{"pattern":"` + tests.RtRepo1 + "/" + symlinkTreePath + `/*","target":"` + filepath.ToSlash(out) + `/","flat":"true"}]}`
	require.NoError(t, os.WriteFile(specFile, []byte(spec), 0644))

	queries, err := execAndCaptureAql(t, "dl", "--spec="+specFile)
	require.NoError(t, err)
	require.Len(t, queries, 2)
	assert.NotContains(t, queries[0], includePropsAql)
	assert.Contains(t, queries[1], symlinkPropsAql)
	assertFlatSymlinkTreeDownloaded(t, out, linkTarget)
}

func TestDownloadNoMatchesWithSeparatePropsQuery(t *testing.T) {
	initArtifactoryTest(t, "")
	defer cleanArtifactoryTest()

	out := tests.Out + string(filepath.Separator)
	queries, err := execAndCaptureAql(t, "dl", tests.RtRepo1+"/does/not/exist/", out, "--flat")
	require.NoError(t, err)
	require.NotEmpty(t, queries)
	assert.NotContains(t, queries[0], includePropsAql)
	assert.NoDirExists(t, tests.Out)
}
