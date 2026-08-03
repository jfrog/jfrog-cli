package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	biutils "github.com/jfrog/build-info-go/utils"
	agentTestutil "github.com/jfrog/jfrog-cli-artifactory/agent/common/testutil"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/generic"
	artUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	coreBuild "github.com/jfrog/jfrog-cli-core/v2/common/build"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
)

// Agent Skills e2e suite (`jf agent skills`).
//
// Run:
//
//	go test -v github.com/jfrog/jfrog-cli --timeout 0 --test.agentSkills
//
// Skills differ from agent plugins in ways that matter for these tests:
//   - Default install scope is project (plugins default to global).
//   - No marketplace index; version discovery uses the Skills REST API.
//   - Publish has an Xray gate (--skip-scan / JFROG_CLI_SKIP_SKILLS_SCAN).
//   - Evidence quiet-failure override is JFROG_SKILLS_DISABLE_QUIET_FAILURE.
//   - Custom agents live under "skills-agents" in agent-config.json.
//
// Indexing note: upload to storage can succeed before the Skills versions/list/search
// APIs see the artifact. publishTestSkill waits on SkillVersionExists; list/search/
// check-updates helpers also retry with backoff when they hit those APIs directly.
//
// Publish does not validate the target repository package type: publishing to a generic
// local repo succeeds, so there is no wrong-repo-type failure to assert here.

// ---------------------------------------------------------------------------
// Init / cleanup
// ---------------------------------------------------------------------------

func InitAgentSkillsTests() {
	initArtifactoryCli()
	cleanUpOldRepositories()
	tests.AddTimestampToGlobalVars()
	createRequiredRepos()
}

func CleanAgentSkillsTests() {
	deleteCreatedRepos()
}

func initAgentSkillsTest(t *testing.T) {
	if !*tests.TestAgentSkills {
		t.Skip("Skipping Agent Skills test. To run Agent Skills test add the '-test.agentSkills=true' option.")
	}
	createJfrogHomeConfig(t, false)
	require.True(t, isRepoExist(tests.AgentSkillsLocalRepo), "agent skills local repo does not exist: "+tests.AgentSkillsLocalRepo)
	// Shared test Artifactory has no evidence/One-Model service. Disable the quiet-failure
	// evidence gate so install/update do not block on 403.
	t.Setenv("JFROG_SKILLS_DISABLE_QUIET_FAILURE", "true")
}

func cleanAgentSkillsTest() {
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.AgentSkillsBuildName, artHttpDetails)
	tests.CleanFileSystem()
}

// runAgentSkillsCmd executes `jf agent skills <args...>`.
func runAgentSkillsCmd(t *testing.T, args ...string) error {
	t.Helper()
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	return jfrogCli.Exec(append([]string{"agent", "skills"}, args...)...)
}

// runAgentSkillsCmdWithOutput executes `jf agent skills <args...>` and returns captured stdout.
func runAgentSkillsCmdWithOutput(t *testing.T, args ...string) (string, error) {
	t.Helper()
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	return jfrogCli.RunCliCmdWithOutputs(t, append([]string{"agent", "skills"}, args...)...)
}

// assertErrorContainsAll requires a non-nil error whose message contains every substring.
// Prefer this over loose OR-chains that pass on unrelated failures.
func assertErrorContainsAll(t *testing.T, err error, substrings ...string) {
	t.Helper()
	require.Error(t, err)
	msg := err.Error()
	for _, substring := range substrings {
		assert.Contains(t, msg, substring, "error %q should contain %q", msg, substring)
	}
}

// createTestSkill copies the fixture and rewrites SKILL.md frontmatter name/version.
func createTestSkill(t *testing.T, slug, version string) string {
	t.Helper()
	skillSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "agent_skills", "test-skill")
	skillPath, cleanup := coretests.CreateTempDirWithCallbackAndAssert(t)
	t.Cleanup(cleanup)

	require.NoError(t, biutils.CopyDir(skillSrc, skillPath, true, nil))
	writeSkillMD(t, skillPath, slug, version, "Integration test skill")
	return skillPath
}

func writeSkillMD(t *testing.T, skillDir, slug, version, description string) {
	t.Helper()
	content := fmt.Sprintf("---\nname: %s\ndescription: %s\nversion: %s\n---\n\n# %s\n\nBody content for agent skills e2e.\n",
		slug, description, version, slug)
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)) // #nosec G306 -- test fixture
}

// skillArtifactPath is the generated (hyphen) publish layout:
// {repo}/{slug}/{version}/{slug}-{version}.zip
func skillArtifactPath(repo, slug, version string) string {
	return repo + "/" + slug + "/" + version + "/" + slug + "-" + version + ".zip"
}

// skillPrebuiltArtifactPath is the flat-upload path for zip/{slug}_{version}.zip.
// Install still resolves the hyphen filename, so prebuilt publish and install remain mismatched.
func skillPrebuiltArtifactPath(repo, slug, version string) string {
	return repo + "/" + slug + "/" + version + "/" + slug + "_" + version + ".zip"
}

// skillArtifactExists reports storage presence via the file-info API. GetItemProps is not
// usable here: it maps a 404 "no properties" response to (nil, nil), so it cannot distinguish
// a missing artifact from an artifact without properties.
func skillArtifactExists(artifactPath string) (bool, error) {
	sm, err := artUtils.CreateServiceManager(serverDetails, -1, 0, false)
	if err != nil {
		return false, err
	}
	info, err := sm.FileInfo(artifactPath)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}
	return info != nil, nil
}

func assertSkillArtifactExists(t *testing.T, artifactPath string) {
	t.Helper()
	exists, err := skillArtifactExists(artifactPath)
	require.NoError(t, err)
	require.True(t, exists, "artifact should exist at %s", artifactPath)
}

func assertSkillArtifactMissing(t *testing.T, artifactPath string) {
	t.Helper()
	exists, err := skillArtifactExists(artifactPath)
	require.NoError(t, err)
	require.False(t, exists, "artifact should not exist at %s", artifactPath)
}

func assertSkillExists(t *testing.T, slug, version string) {
	t.Helper()
	assertSkillArtifactExists(t, skillArtifactPath(tests.AgentSkillsLocalRepo, slug, version))
}

// assertSkillAbsent retries the storage file-info lookup until the artifact is gone,
// since DELETE visibility can briefly lag.
func assertSkillAbsent(t *testing.T, slug, version string) {
	t.Helper()
	path := skillArtifactPath(tests.AgentSkillsLocalRepo, slug, version)
	require.NoError(t, retryWithBackoffSkills(t, "wait for deleted skill artifact "+path, func() error {
		exists, err := skillArtifactExists(path)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("artifact still present at %s", path)
		}
		return nil
	}))
}

// waitForSkillIndexed polls the Skills versions API until slug@version is visible.
// Install, update, list --repo, search, collision checks, and delete --dry-run all depend
// on this index; a successful storage upload alone is not enough.
func waitForSkillIndexed(t *testing.T, slug, version string) {
	t.Helper()
	description := fmt.Sprintf("wait for Skills API to index %s@%s", slug, version)
	require.NoError(t, retryWithBackoffSkills(t, description, func() error {
		sm, err := artUtils.CreateServiceManager(serverDetails, -1, 0, false)
		if err != nil {
			return err
		}
		exists, err := sm.SkillVersionExists(tests.AgentSkillsLocalRepo, slug, version)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Skills API does not yet list %s@%s", slug, version)
		}
		return nil
	}))
}

// publishTestSkill publishes a fixture skill, asserts storage presence, and waits for Skills API indexing.
func publishTestSkill(t *testing.T, slug, version string) {
	t.Helper()
	skillPath := createTestSkill(t, slug, version)
	require.NoError(t, runAgentSkillsCmd(t,
		"publish", skillPath,
		"--repo="+tests.AgentSkillsLocalRepo,
		"--skip-scan",
	))
	assertSkillExists(t, slug, version)
	waitForSkillIndexed(t, slug, version)
}

// agentSkillHarnessCase covers built-in skills harnesses. Project scope is the default
// for skills (unlike plugins). Multi-harness install is allowed; list rejects commas.
// Path maps intentionally duplicate agent/skills/common/agents.go so e2e fails if layout drifts.
type agentSkillHarnessCase struct {
	name      string
	harnesses []string
}

func agentSkillHarnessCases() []agentSkillHarnessCase {
	return []agentSkillHarnessCase{
		{name: "cursor", harnesses: []string{"cursor"}},
		{name: "claude-code", harnesses: []string{"claude-code"}},
		{name: "github-copilot", harnesses: []string{"github-copilot"}},
		{name: "windsurf", harnesses: []string{"windsurf"}},
		{name: "codex", harnesses: []string{"codex"}},
		{name: "cross-agent", harnesses: []string{"cross-agent"}},
		{name: "cursor,claude-code", harnesses: []string{"cursor", "claude-code"}},
	}
}

func skillsHarnessFlag(harnesses []string) string {
	return strings.Join(harnesses, ",")
}

func skillProjectRelDir(harness string) string {
	switch harness {
	case "claude-code":
		return filepath.Join(".claude", "skills")
	case "cursor":
		return filepath.Join(".cursor", "skills")
	case "github-copilot":
		return filepath.Join(".github", "skills")
	case "windsurf":
		return filepath.Join(".windsurf", "skills")
	case "codex":
		return filepath.Join(".codex", "skills")
	case "cross-agent":
		return filepath.Join(".agents", "skills")
	default:
		return filepath.Join("."+harness, "skills")
	}
}

func skillGlobalInstallDir(homeDir, harness, slug string) string {
	switch harness {
	case "claude-code":
		return filepath.Join(homeDir, ".claude", "skills", slug)
	case "cursor":
		return filepath.Join(homeDir, ".cursor", "skills", slug)
	case "github-copilot":
		return filepath.Join(homeDir, ".copilot", "skills", slug)
	case "windsurf":
		return filepath.Join(homeDir, ".codeium", "windsurf", "skills", slug)
	case "codex":
		return filepath.Join(homeDir, ".codex", "skills", slug)
	case "cross-agent":
		return filepath.Join(homeDir, ".agents", "skills", slug)
	default:
		return filepath.Join(homeDir, "."+harness, "skills", slug)
	}
}

// skillInfoManifest mirrors .jfrog/skill-info.json written by install/update.
type skillInfoManifest struct {
	SchemaVersion    int    `json:"schemaVersion"`
	Repo             string `json:"repo"`
	Slug             string `json:"slug"`
	InstalledVersion string `json:"installedVersion"`
	Scope            string `json:"scope"`
	Agent            string `json:"agent"`
	ProjectDir       string `json:"projectDir"`
}

// assertSkillInfoManifest validates .jfrog/skill-info.json written by install/update.
func assertSkillInfoManifest(t *testing.T, installDir, slug, version, agent, scope, repo string) {
	t.Helper()
	manifestPath := filepath.Join(installDir, ".jfrog", "skill-info.json")
	require.FileExists(t, manifestPath)
	data, err := os.ReadFile(manifestPath) // #nosec G304 -- path under test temp dir
	require.NoError(t, err)
	var manifest skillInfoManifest
	require.NoError(t, json.Unmarshal(data, &manifest))
	assert.Equal(t, 1, manifest.SchemaVersion, "schemaVersion")
	assert.Equal(t, slug, manifest.Slug, "slug")
	assert.Equal(t, version, manifest.InstalledVersion, "installedVersion")
	assert.Equal(t, agent, manifest.Agent, "agent")
	assert.Equal(t, scope, manifest.Scope, "scope")
	assert.Equal(t, repo, manifest.Repo, "repo")
	switch scope {
	case "project":
		assert.NotEmpty(t, manifest.ProjectDir, "projectDir should be set for project scope")
	case "path", "global":
		assert.Empty(t, manifest.ProjectDir, "projectDir should be omitted for %s scope", scope)
	}
}

// extractJSONPayload returns the first JSON array/object in CLI output, preferring a
// document that starts on its own line so log prefixes containing '[' do not win.
func extractJSONPayload(t *testing.T, out string) string {
	t.Helper()
	trimmed := strings.TrimSpace(out)
	arrayStart := strings.LastIndex(trimmed, "\n[")
	if arrayStart >= 0 {
		arrayStart++
	} else {
		arrayStart = strings.Index(trimmed, "[")
	}
	objectStart := strings.LastIndex(trimmed, "\n{")
	if objectStart >= 0 {
		objectStart++
	} else {
		objectStart = strings.Index(trimmed, "{")
	}
	switch {
	case arrayStart >= 0 && (objectStart < 0 || arrayStart < objectStart):
		end := strings.LastIndex(trimmed, "]")
		require.Greater(t, end, arrayStart, "JSON array must have a closing bracket, got: %q", out)
		return trimmed[arrayStart : end+1]
	case objectStart >= 0:
		end := strings.LastIndex(trimmed, "}")
		require.Greater(t, end, objectStart, "JSON object must have a closing brace, got: %q", out)
		return trimmed[objectStart : end+1]
	default:
		t.Fatalf("output must contain JSON, got: %q", out)
		return ""
	}
}

// Skills API / property indexing can lag after upload. Budget mirrors agent plugins e2e.
const (
	skillsIndexFirstWait = 2 * time.Second
	skillsIndexMaxWait   = 10 * time.Second
	skillsIndexBudget    = 90 * time.Second
)

// retryWithBackoffSkills runs operation until it succeeds or the Skills index budget is spent,
// doubling the wait between attempts. Failures name the timed-out operation.
func retryWithBackoffSkills(t *testing.T, description string, operation func() error) error {
	t.Helper()
	deadline := time.Now().Add(skillsIndexBudget)
	wait := skillsIndexFirstWait
	for attempt := 1; ; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}
		if time.Now().Add(wait).After(deadline) {
			return fmt.Errorf("%s did not succeed within %s (%d attempts): %w",
				description, skillsIndexBudget, attempt, err)
		}
		t.Logf("%s: attempt %d failed (%v); retrying in %s", description, attempt, err, wait)
		time.Sleep(wait)
		if wait *= 2; wait > skillsIndexMaxWait {
			wait = skillsIndexMaxWait
		}
	}
}

type agentSkillsSearchRow struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Repository  string `json:"repository"`
	Description string `json:"description"`
}

type agentSkillsRepoListRow struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Source      string `json:"source"`
}

type agentSkillsLocalListRow struct {
	Name           string `json:"name"`
	Version        string `json:"version"`
	Description    string `json:"description"`
	Repo           string `json:"repo"`
	Path           string `json:"path"`
	RegistryLatest string `json:"registryLatest"`
	Status         string `json:"status"`
}

type agentSkillsSummaryResult struct {
	Agent  string `json:"agent"`
	Scope  string `json:"scope"`
	Path   string `json:"path"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

type agentSkillsSummaryJSON struct {
	Slug    string                     `json:"slug"`
	Version string                     `json:"version"`
	Results []agentSkillsSummaryResult `json:"results"`
}

func parseSkillsSearchRows(out string) ([]agentSkillsSearchRow, error) {
	payload := out
	if start := strings.Index(out, "["); start >= 0 {
		end := strings.LastIndex(out, "]")
		if end <= start {
			return nil, fmt.Errorf("search output does not contain a JSON array, got: %q", out)
		}
		payload = out[start : end+1]
	}
	var rows []agentSkillsSearchRow
	if err := json.Unmarshal([]byte(payload), &rows); err != nil {
		return nil, fmt.Errorf("parse search JSON: %w", err)
	}
	return rows, nil
}

// searchSkillsRowsWithRetry waits for Skills API search (or --prop property search) to list slug@version.
func searchSkillsRowsWithRetry(t *testing.T, slug, version string, searchArgs ...string) []agentSkillsSearchRow {
	t.Helper()
	var found []agentSkillsSearchRow
	description := fmt.Sprintf("wait for skills search to list %s %s", slug, version)
	require.NoError(t, retryWithBackoffSkills(t, description, func() error {
		out, err := runAgentSkillsCmdWithOutput(t, searchArgs...)
		if err != nil {
			return err
		}
		rows, err := parseSkillsSearchRows(out)
		if err != nil {
			return err
		}
		for _, row := range rows {
			if row.Name == slug && row.Version == version {
				found = rows
				return nil
			}
		}
		return fmt.Errorf("search does not list skill %q at version %q yet; output: %s", slug, version, out)
	}))
	return found
}

func assertSearchSkillsRowsContain(t *testing.T, rows []agentSkillsSearchRow, slug, version, repo string) {
	t.Helper()
	for _, row := range rows {
		if row.Name == slug {
			assert.Equal(t, version, row.Version, "search row version for %q", slug)
			assert.Equal(t, repo, row.Repository, "search row repository for %q", slug)
			return
		}
	}
	t.Fatalf("search rows did not contain skill %q; rows: %+v", slug, rows)
}

// listRepoSkillsWithRetry waits for list --repo (Skills list API) to include slug@version.
func listRepoSkillsWithRetry(t *testing.T, slug, version string) []agentSkillsRepoListRow {
	t.Helper()
	var found []agentSkillsRepoListRow
	description := fmt.Sprintf("wait for list --repo to show %s@%s", slug, version)
	require.NoError(t, retryWithBackoffSkills(t, description, func() error {
		out, err := runAgentSkillsCmdWithOutput(t,
			"list", "--repo="+tests.AgentSkillsLocalRepo, "--format=json")
		if err != nil {
			return err
		}
		var rows []agentSkillsRepoListRow
		if err := json.Unmarshal([]byte(extractJSONPayload(t, out)), &rows); err != nil {
			return fmt.Errorf("parse list --repo JSON: %w", err)
		}
		for _, row := range rows {
			if row.Name == slug && row.Version == version {
				found = rows
				return nil
			}
		}
		return fmt.Errorf("list --repo does not yet include %s@%s; output: %s", slug, version, out)
	}))
	return found
}

// waitForLocalListStatus waits until list --check-updates reports wantStatus for slug and
// returns the matching row so callers can assert the registry version it resolved.
// Pass projectDir="" to use --global instead of --project-dir.
func waitForLocalListStatus(t *testing.T, harness, projectDir, slug, wantStatus string) agentSkillsLocalListRow {
	t.Helper()
	args := []string{"list", "--harness=" + harness, "--check-updates", "--format=json"}
	if projectDir != "" {
		args = append(args, "--project-dir="+projectDir)
	} else {
		args = append(args, "--global")
	}
	var matched agentSkillsLocalListRow
	description := fmt.Sprintf("wait for check-updates status %q for %s", wantStatus, slug)
	require.NoError(t, retryWithBackoffSkills(t, description, func() error {
		out, err := runAgentSkillsCmdWithOutput(t, args...)
		if err != nil {
			return err
		}
		var rows []agentSkillsLocalListRow
		if err := json.Unmarshal([]byte(extractJSONPayload(t, out)), &rows); err != nil {
			return fmt.Errorf("parse local list JSON: %w", err)
		}
		for _, row := range rows {
			if row.Name != slug {
				continue
			}
			if !strings.EqualFold(row.Status, wantStatus) {
				return fmt.Errorf("status=%q want %q (latest=%q)", row.Status, wantStatus, row.RegistryLatest)
			}
			matched = row
			return nil
		}
		return fmt.Errorf("skill %q not in local list output: %s", slug, out)
	}))
	return matched
}

func parseSkillsSummaryJSON(t *testing.T, out string) agentSkillsSummaryJSON {
	t.Helper()
	var summary agentSkillsSummaryJSON
	require.NoError(t, json.Unmarshal([]byte(extractJSONPayload(t, out)), &summary), "install/update summary JSON")
	return summary
}

// ---------------------------------------------------------------------------
// Publish
// ---------------------------------------------------------------------------

// TestAgentSkillsPublish covers publish happy path, artifact layout, and checksums
// (plan scenarios 8, 9, 19, 75).
func TestAgentSkillsPublish(t *testing.T) {
	initAgentSkillsTest(t)
	defer cleanAgentSkillsTest()

	slug := "publish-skill"
	version := "1.0.0"
	publishTestSkill(t, slug, version)

	artifactPath := skillArtifactPath(tests.AgentSkillsLocalRepo, slug, version)
	searchSpec := spec.NewBuilder().Pattern(artifactPath).BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(searchSpec)
	reader, err := searchCmd.Search()
	require.NoError(t, err)
	// Read-side Close: search reader has no write flush; nothing actionable on failure.
	defer func() { _ = reader.Close() }()
	item := new(artUtils.SearchResult)
	require.NoError(t, reader.NextRecord(item), "published zip must be searchable at %s", artifactPath)
	assert.Equal(t, artifactPath, item.Path, "artifact layout must be {repo}/{slug}/{version}/{slug}-{version}.zip")
	assert.NotEmpty(t, item.Sha256, "Artifactory must store sha256")
	assert.NotEmpty(t, item.Md5, "Artifactory must store md5")
	assert.NotEmpty(t, item.Sha1, "Artifactory must store sha1")
}

// TestAgentSkillsPublishPrebuiltZip verifies that zip/{slug}_{version}.zip is uploaded
// as-is under the underscore basename (plan scenario 98). Install still resolves the
// hyphen filename, so this test only asserts publish/storage layout.
func TestAgentSkillsPublishPrebuiltZip(t *testing.T) {
	initAgentSkillsTest(t)
	defer cleanAgentSkillsTest()

	slug := "prebuilt-zip-skill"
	version := "1.0.0"
	skillDir := createTestSkill(t, slug, version)

	zipSubDir := filepath.Join(skillDir, "zip")
	require.NoError(t, os.MkdirAll(zipSubDir, 0755)) // #nosec G301 -- test directory
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	f, err := zw.Create("placeholder.txt")
	require.NoError(t, err)
	_, err = f.Write([]byte("placeholder"))
	require.NoError(t, err)
	require.NoError(t, zw.Close())
	zipBytes := zipBuf.Bytes()
	require.NoError(t, os.WriteFile(
		filepath.Join(zipSubDir, slug+"_"+version+".zip"), zipBytes, 0644, // #nosec G306 -- test fixture
	))

	require.NoError(t, runAgentSkillsCmd(t,
		"publish", skillDir,
		"--repo="+tests.AgentSkillsLocalRepo,
		"--skip-scan",
	))

	assertSkillArtifactExists(t, skillPrebuiltArtifactPath(tests.AgentSkillsLocalRepo, slug, version))
	// Hyphen layout must not be fabricated for a prebuilt underscore zip.
	assertSkillArtifactMissing(t, skillArtifactPath(tests.AgentSkillsLocalRepo, slug, version))
}

// TestAgentSkillsPublishVersionResolution covers frontmatter version, --version override,
// and quiet/CI collision (plan scenarios 10, 11, 13).
func TestAgentSkillsPublishVersionResolution(t *testing.T) {
	initAgentSkillsTest(t)
	defer cleanAgentSkillsTest()

	t.Run("frontmatter-version", func(t *testing.T) {
		slug := "frontmatter-version-skill"
		skillPath := createTestSkill(t, slug, "1.2.3")
		require.NoError(t, runAgentSkillsCmd(t, "publish", skillPath,
			"--repo="+tests.AgentSkillsLocalRepo, "--skip-scan"))
		assertSkillExists(t, slug, "1.2.3")
		waitForSkillIndexed(t, slug, "1.2.3")
	})

	t.Run("version-flag-override", func(t *testing.T) {
		slug := "version-override-skill"
		skillPath := createTestSkill(t, slug, "1.0.0")
		require.NoError(t, runAgentSkillsCmd(t, "publish", skillPath,
			"--repo="+tests.AgentSkillsLocalRepo, "--version=2.0.0", "--skip-scan"))
		assertSkillExists(t, slug, "2.0.0")
		waitForSkillIndexed(t, slug, "2.0.0")
	})

	t.Run("collision-ci", func(t *testing.T) {
		slug := "collision-skill"
		publishTestSkill(t, slug, "1.0.0")
		t.Setenv("CI", "true")
		err := runAgentSkillsCmd(t, "publish", createTestSkill(t, slug, "1.0.0"),
			"--repo="+tests.AgentSkillsLocalRepo, "--skip-scan", "--quiet")
		assertErrorContainsAll(t, err,
			"version 1.0.0 of skill 'collision-skill' already exists")
	})
}

// TestAgentSkillsPublishValidation merges SKILL.md / slug / semver / path validation
// (plan scenarios 14–17, 77, 96).
func TestAgentSkillsPublishValidation(t *testing.T) {
	initAgentSkillsTest(t)
	defer cleanAgentSkillsTest()

	cases := []struct {
		name        string
		setup       func(t *testing.T) (skillPath string, extraArgs []string)
		errContains []string
	}{
		{
			name: "missing-skill-md",
			setup: func(t *testing.T) (string, []string) {
				dir := t.TempDir()
				return dir, nil
			},
			errContains: []string{"SKILL.md"},
		},
		{
			name: "missing-name",
			setup: func(t *testing.T) (string, []string) {
				dir := t.TempDir()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\ndescription: x\nversion: 1.0.0\n---\n"), 0644)) // #nosec G306
				return dir, nil
			},
			errContains: []string{"SKILL.md missing required 'name' field"},
		},
		{
			name: "invalid-frontmatter",
			setup: func(t *testing.T) (string, []string) {
				dir := t.TempDir()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# no frontmatter\n"), 0644)) // #nosec G306
				return dir, nil
			},
			errContains: []string{"YAML frontmatter delimiter"},
		},
		{
			name: "invalid-slug",
			setup: func(t *testing.T) (string, []string) {
				return createTestSkill(t, "Bad_Slug", "1.0.0"), nil
			},
			errContains: []string{"invalid slug"},
		},
		{
			name: "invalid-semver",
			setup: func(t *testing.T) (string, []string) {
				return createTestSkill(t, "bad-semver-skill", "1.0.0"), []string{"--version=not-a-version"}
			},
			errContains: []string{"invalid version", "expected format major.minor.patch"},
		},
		{
			name: "missing-path-arg",
			setup: func(t *testing.T) (string, []string) {
				return "", nil
			},
			errContains: []string{"usage: jf agent skills publish"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			skillPath, extra := tc.setup(t)
			args := []string{"publish"}
			if skillPath != "" {
				args = append(args, skillPath)
			}
			args = append(args, "--repo="+tests.AgentSkillsLocalRepo, "--skip-scan")
			args = append(args, extra...)
			err := runAgentSkillsCmd(t, args...)
			assertErrorContainsAll(t, err, tc.errContains...)
		})
	}
}

// ---------------------------------------------------------------------------
// Install
// ---------------------------------------------------------------------------

// TestAgentSkillsInstallHarnesses covers project/global installs across built-in
// harnesses, multi-harness, specific version, latest, and skill-info.json
// (plan scenarios 23–26, 28–30).
func TestAgentSkillsInstallHarnesses(t *testing.T) {
	initAgentSkillsTest(t)
	defer cleanAgentSkillsTest()

	slug := "install-harness-skill"
	publishTestSkill(t, slug, "1.0.0")
	publishTestSkill(t, slug, "2.0.0")

	for _, tc := range agentSkillHarnessCases() {
		t.Run("project/"+tc.name, func(t *testing.T) {
			projectDir := t.TempDir()
			require.NoError(t, runAgentSkillsCmd(t,
				"install", slug,
				"--repo="+tests.AgentSkillsLocalRepo,
				"--harness="+skillsHarnessFlag(tc.harnesses),
				"--project-dir="+projectDir,
				"--version=1.0.0",
			))
			for _, harness := range tc.harnesses {
				installDir := filepath.Join(projectDir, skillProjectRelDir(harness), slug)
				require.DirExists(t, installDir)
				assert.FileExists(t, filepath.Join(installDir, "SKILL.md"))
				assertSkillInfoManifest(t, installDir, slug, "1.0.0", harness, "project", tests.AgentSkillsLocalRepo)
			}
		})
	}

	t.Run("global/cursor", func(t *testing.T) {
		homeDir := t.TempDir()
		t.Setenv("HOME", homeDir)
		t.Setenv("USERPROFILE", homeDir)
		require.NoError(t, runAgentSkillsCmd(t,
			"install", slug,
			"--repo="+tests.AgentSkillsLocalRepo,
			"--harness=cursor",
			"--global",
			"--version=2.0.0",
		))
		installDir := skillGlobalInstallDir(homeDir, "cursor", slug)
		require.DirExists(t, installDir)
		assertSkillInfoManifest(t, installDir, slug, "2.0.0", "cursor", "global", tests.AgentSkillsLocalRepo)
	})

	t.Run("latest-via-path", func(t *testing.T) {
		installBase := t.TempDir()
		require.NoError(t, runAgentSkillsCmd(t,
			"install", slug,
			"--repo="+tests.AgentSkillsLocalRepo,
			"--path="+installBase,
		))
		assertSkillInfoManifest(t, filepath.Join(installBase, slug), slug, "2.0.0", "(path)", "path", tests.AgentSkillsLocalRepo)
	})
}

// TestAgentSkillsInstallPathAndFlags covers --path install and mutual exclusions
// (plan scenarios 27, 33, 34).
func TestAgentSkillsInstallPathAndFlags(t *testing.T) {
	initAgentSkillsTest(t)
	defer cleanAgentSkillsTest()

	slug := "path-flags-skill"
	publishTestSkill(t, slug, "1.0.0")

	t.Run("path-install", func(t *testing.T) {
		base := t.TempDir()
		require.NoError(t, runAgentSkillsCmd(t,
			"install", slug,
			"--repo="+tests.AgentSkillsLocalRepo,
			"--path="+base,
			"--version=1.0.0",
		))
		assert.FileExists(t, filepath.Join(base, slug, "SKILL.md"))
		assertSkillInfoManifest(t, filepath.Join(base, slug), slug, "1.0.0", "(path)", "path", tests.AgentSkillsLocalRepo)
	})

	cases := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "path-and-harness",
			args: []string{"install", slug, "--repo=" + tests.AgentSkillsLocalRepo, "--path=" + t.TempDir(), "--harness=cursor"},
			want: []string{"--path cannot be combined with --harness"},
		},
		{
			name: "global-and-project-dir",
			args: []string{"install", slug, "--repo=" + tests.AgentSkillsLocalRepo, "--harness=cursor", "--global", "--project-dir=" + t.TempDir()},
			want: []string{"--global and --project-dir are mutually exclusive"},
		},
		{
			name: "unknown-harness",
			args: []string{"install", slug, "--repo=" + tests.AgentSkillsLocalRepo, "--harness=not-a-real-agent", "--project-dir=" + t.TempDir()},
			want: []string{"unknown agent"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertErrorContainsAll(t, runAgentSkillsCmd(t, tc.args...), tc.want...)
		})
	}
}

// TestAgentSkillsInstallNotFound covers missing skill and missing version of an
// existing skill (plan scenario 32).
func TestAgentSkillsInstallNotFound(t *testing.T) {
	initAgentSkillsTest(t)
	defer cleanAgentSkillsTest()

	t.Run("missing-skill", func(t *testing.T) {
		err := runAgentSkillsCmd(t,
			"install", "missing-skill-xyz",
			"--repo="+tests.AgentSkillsLocalRepo,
			"--path="+t.TempDir(),
		)
		assertErrorContainsAll(t, err, "skill 'missing-skill-xyz' not found")
	})

	t.Run("missing-version", func(t *testing.T) {
		slug := "missing-version-skill"
		publishTestSkill(t, slug, "1.0.0")
		err := runAgentSkillsCmd(t,
			"install", slug,
			"--repo="+tests.AgentSkillsLocalRepo,
			"--path="+t.TempDir(),
			"--version=9.9.9",
		)
		assertErrorContainsAll(t, err, "version '9.9.9' not found", "Available versions")
	})
}

// TestAgentSkillsInstallEvidenceGate covers quiet-failure evidence behavior
// (plan scenarios 35, 36).
func TestAgentSkillsInstallEvidenceGate(t *testing.T) {
	initAgentSkillsTest(t)
	defer cleanAgentSkillsTest()

	slug := "evidence-gate-skill"
	publishTestSkill(t, slug, "1.0.0")

	t.Run("ci-gate", func(t *testing.T) {
		t.Setenv("CI", "true")
		t.Setenv("JFROG_SKILLS_DISABLE_QUIET_FAILURE", "")
		err := runAgentSkillsCmd(t,
			"install", slug,
			"--repo="+tests.AgentSkillsLocalRepo,
			"--path="+t.TempDir(),
			"--quiet",
		)
		if err == nil {
			t.Skip("evidence quiet-failure gate not enforced on this Artifactory instance")
		}
		assertErrorContainsAll(t, err, "evidence verification failed", "JFROG_SKILLS_DISABLE_QUIET_FAILURE")
	})

	t.Run("disabled", func(t *testing.T) {
		t.Setenv("CI", "true")
		t.Setenv("JFROG_SKILLS_DISABLE_QUIET_FAILURE", "true")
		require.NoError(t, runAgentSkillsCmd(t,
			"install", slug,
			"--repo="+tests.AgentSkillsLocalRepo,
			"--path="+t.TempDir(),
			"--quiet",
		))
	})
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

// TestAgentSkillsUpdate covers update, dry-run, force, not-installed, and JSON
// (plan scenarios 38–41, 43).
func TestAgentSkillsUpdate(t *testing.T) {
	initAgentSkillsTest(t)
	defer cleanAgentSkillsTest()

	slug := "update-skill"
	publishTestSkill(t, slug, "1.0.0")
	publishTestSkill(t, slug, "2.0.0")

	projectDir := t.TempDir()
	require.NoError(t, runAgentSkillsCmd(t,
		"install", slug,
		"--repo="+tests.AgentSkillsLocalRepo,
		"--harness=cursor",
		"--project-dir="+projectDir,
		"--version=1.0.0",
	))
	installDir := filepath.Join(projectDir, ".cursor", "skills", slug)
	assertSkillInfoManifest(t, installDir, slug, "1.0.0", "cursor", "project", tests.AgentSkillsLocalRepo)

	t.Run("dry-run", func(t *testing.T) {
		require.NoError(t, runAgentSkillsCmd(t,
			"update", slug,
			"--repo="+tests.AgentSkillsLocalRepo,
			"--harness=cursor",
			"--project-dir="+projectDir,
			"--dry-run",
		))
		assertSkillInfoManifest(t, installDir, slug, "1.0.0", "cursor", "project", tests.AgentSkillsLocalRepo)
	})

	t.Run("update-to-latest", func(t *testing.T) {
		require.NoError(t, runAgentSkillsCmd(t,
			"update", slug,
			"--repo="+tests.AgentSkillsLocalRepo,
			"--harness=cursor",
			"--project-dir="+projectDir,
		))
		assertSkillInfoManifest(t, installDir, slug, "2.0.0", "cursor", "project", tests.AgentSkillsLocalRepo)
	})

	t.Run("force-reinstall", func(t *testing.T) {
		out, err := runAgentSkillsCmdWithOutput(t,
			"update", slug,
			"--repo="+tests.AgentSkillsLocalRepo,
			"--harness=cursor",
			"--project-dir="+projectDir,
			"--force",
			"--format=json",
		)
		require.NoError(t, err)
		summary := parseSkillsSummaryJSON(t, out)
		assert.Equal(t, slug, summary.Slug)
		assert.Equal(t, "2.0.0", summary.Version)
		require.NotEmpty(t, summary.Results)
		assert.Equal(t, "ok", summary.Results[0].Status)
		assertSkillInfoManifest(t, installDir, slug, "2.0.0", "cursor", "project", tests.AgentSkillsLocalRepo)
	})

	// A published-but-not-installed skill fails per target; the returned error is generic and
	// the "not installed" reason is reported in the summary detail.
	t.Run("not-installed", func(t *testing.T) {
		out, err := runAgentSkillsCmdWithOutput(t,
			"update", slug,
			"--repo="+tests.AgentSkillsLocalRepo,
			"--harness=cursor",
			"--project-dir="+t.TempDir(),
			"--format=json",
		)
		assertErrorContainsAll(t, err, "update failed for all targets")
		summary := parseSkillsSummaryJSON(t, out)
		require.NotEmpty(t, summary.Results)
		assert.Equal(t, "failed", summary.Results[0].Status)
		assert.Contains(t, summary.Results[0].Detail, "not installed")
	})

	t.Run("not-in-repo", func(t *testing.T) {
		assertErrorContainsAll(t, runAgentSkillsCmd(t,
			"update", "never-published-skill",
			"--repo="+tests.AgentSkillsLocalRepo,
			"--harness=cursor",
			"--project-dir="+t.TempDir(),
		), "skill 'never-published-skill' not found in repository")
	})
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

// TestAgentSkillsDelete covers delete, dry-run, and missing --version
// (plan scenarios 45–47).
func TestAgentSkillsDelete(t *testing.T) {
	initAgentSkillsTest(t)
	defer cleanAgentSkillsTest()

	slug := "delete-skill"
	publishTestSkill(t, slug, "1.0.0")
	publishTestSkill(t, slug, "2.0.0")

	t.Run("dry-run", func(t *testing.T) {
		require.NoError(t, runAgentSkillsCmd(t,
			"delete", slug,
			"--repo="+tests.AgentSkillsLocalRepo,
			"--version=1.0.0",
			"--dry-run",
		))
		assertSkillExists(t, slug, "1.0.0")
	})

	t.Run("delete-one-version", func(t *testing.T) {
		require.NoError(t, runAgentSkillsCmd(t,
			"delete", slug,
			"--repo="+tests.AgentSkillsLocalRepo,
			"--version=1.0.0",
		))
		assertSkillAbsent(t, slug, "1.0.0")
		assertSkillExists(t, slug, "2.0.0")
	})

	t.Run("missing-version-flag", func(t *testing.T) {
		assertErrorContainsAll(t, runAgentSkillsCmd(t,
			"delete", slug,
			"--repo="+tests.AgentSkillsLocalRepo,
		), "--version is required")
	})
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

// TestAgentSkillsList covers repo/harness list modes, empty list, mutual exclusions,
// comma rejection, and limit (plan scenarios 48–50, 53–56, 58).
func TestAgentSkillsList(t *testing.T) {
	initAgentSkillsTest(t)
	defer cleanAgentSkillsTest()

	slug := "list-skill"
	publishTestSkill(t, slug, "1.0.0")
	publishTestSkill(t, slug, "2.0.0")

	t.Run("repo-json", func(t *testing.T) {
		rows := listRepoSkillsWithRetry(t, slug, "2.0.0")
		var matched bool
		for _, row := range rows {
			if row.Name == slug {
				matched = true
				assert.Equal(t, "2.0.0", row.Version, "list --repo should surface the latest indexed version")
				assert.Equal(t, "Repo: "+tests.AgentSkillsLocalRepo, row.Source)
				break
			}
		}
		require.True(t, matched, "list --repo JSON must include %s", slug)
	})

	t.Run("harness-after-install", func(t *testing.T) {
		projectDir := t.TempDir()
		require.NoError(t, runAgentSkillsCmd(t,
			"install", slug,
			"--repo="+tests.AgentSkillsLocalRepo,
			"--harness=cursor",
			"--project-dir="+projectDir,
			"--version=2.0.0",
		))
		out, err := runAgentSkillsCmdWithOutput(t,
			"list",
			"--harness=cursor",
			"--project-dir="+projectDir,
			"--format=json",
		)
		require.NoError(t, err)
		var rows []agentSkillsLocalListRow
		require.NoError(t, json.Unmarshal([]byte(extractJSONPayload(t, out)), &rows))
		require.NotEmpty(t, rows)
		assert.Equal(t, slug, rows[0].Name)
		assert.Equal(t, "2.0.0", rows[0].Version)
		assert.Equal(t, tests.AgentSkillsLocalRepo, rows[0].Repo)
		// list always reports forward-slash paths, including on Windows.
		assert.Contains(t, rows[0].Path, ".cursor/skills/"+slug)
	})

	// With no skills directory the command logs a notice and prints no JSON document.
	t.Run("empty-local", func(t *testing.T) {
		out, err := runAgentSkillsCmdWithOutput(t,
			"list",
			"--harness=cursor",
			"--project-dir="+t.TempDir(),
			"--format=json",
		)
		require.NoError(t, err)
		payload := strings.TrimSpace(out)
		if payload == "" {
			return
		}
		var rows []agentSkillsLocalListRow
		require.NoError(t, json.Unmarshal([]byte(extractJSONPayload(t, out)), &rows))
		assert.Empty(t, rows, "empty local list must not contain skills")
	})

	flagCases := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "neither-mode",
			args: []string{"list"},
			want: []string{"requires exactly one of"},
		},
		{
			name: "repo-and-harness",
			args: []string{"list", "--repo=" + tests.AgentSkillsLocalRepo, "--harness=cursor"},
			want: []string{"--repo and --harness are mutually exclusive"},
		},
		{
			name: "comma-harness",
			args: []string{"list", "--harness=cursor,claude-code"},
			want: []string{"one harness name"},
		},
		{
			name: "limit-zero",
			args: []string{"list", "--repo=" + tests.AgentSkillsLocalRepo, "--limit=0"},
			want: []string{"--limit must be a positive integer"},
		},
	}
	for _, tc := range flagCases {
		t.Run(tc.name, func(t *testing.T) {
			assertErrorContainsAll(t, runAgentSkillsCmd(t, tc.args...), tc.want...)
		})
	}

	t.Run("limit-repo", func(t *testing.T) {
		for _, extra := range []string{"list-limit-a", "list-limit-b", "list-limit-c"} {
			publishTestSkill(t, extra, "1.0.0")
		}
		out, err := runAgentSkillsCmdWithOutput(t,
			"list", "--repo="+tests.AgentSkillsLocalRepo, "--format=json", "--limit=2")
		require.NoError(t, err)
		var rows []agentSkillsRepoListRow
		require.NoError(t, json.Unmarshal([]byte(extractJSONPayload(t, out)), &rows))
		assert.Len(t, rows, 2)
	})
}

// TestAgentSkillsListCheckUpdates covers behind → current status after update
// (plan scenario 52). Status is filled via Skills versions API, so assertions retry.
func TestAgentSkillsListCheckUpdates(t *testing.T) {
	initAgentSkillsTest(t)
	defer cleanAgentSkillsTest()

	slug := "check-updates-skill"
	publishTestSkill(t, slug, "1.0.0")
	publishTestSkill(t, slug, "2.0.0")

	projectDir := t.TempDir()
	require.NoError(t, runAgentSkillsCmd(t,
		"install", slug,
		"--repo="+tests.AgentSkillsLocalRepo,
		"--harness=cursor",
		"--project-dir="+projectDir,
		"--version=1.0.0",
	))

	behindRow := waitForLocalListStatus(t, "cursor", projectDir, slug, "behind")
	assert.Equal(t, "2.0.0", behindRow.RegistryLatest)

	require.NoError(t, runAgentSkillsCmd(t,
		"update", slug,
		"--repo="+tests.AgentSkillsLocalRepo,
		"--harness=cursor",
		"--project-dir="+projectDir,
	))
	currentRow := waitForLocalListStatus(t, "cursor", projectDir, slug, "current")
	assert.Equal(t, "2.0.0", currentRow.RegistryLatest)
}

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------

// TestAgentSkillsSearch covers match, property search, no-match, empty/blank query
// (plan scenarios 59–62). Default search and --prop use different indexes; both retry.
func TestAgentSkillsSearch(t *testing.T) {
	initAgentSkillsTest(t)
	defer cleanAgentSkillsTest()

	slug := "search-skill"
	publishTestSkill(t, slug, "1.0.0")

	t.Run("match", func(t *testing.T) {
		rows := searchSkillsRowsWithRetry(t, slug, "1.0.0",
			"search", slug,
			"--repo="+tests.AgentSkillsLocalRepo,
			"--format=json",
		)
		require.NotEmpty(t, rows)
		assertSearchSkillsRowsContain(t, rows, slug, "1.0.0", tests.AgentSkillsLocalRepo)
	})

	t.Run("prop-match", func(t *testing.T) {
		// --prop queries Artifactory property search on skill.name. That property is
		// server-stamped asynchronously; skip when this Artifactory build does not set it.
		sm, err := artUtils.CreateServiceManager(serverDetails, -1, 0, false)
		require.NoError(t, err)
		props, err := sm.GetItemProps(skillArtifactPath(tests.AgentSkillsLocalRepo, slug, "1.0.0"))
		require.NoError(t, err)
		if props == nil || len(props.Properties["skill.name"]) == 0 {
			t.Skip("Artifactory did not stamp skill.name; property search not available on this instance")
		}
		rows := searchSkillsRowsWithRetry(t, slug, "1.0.0",
			"search", slug,
			"--repo="+tests.AgentSkillsLocalRepo,
			"--format=json",
			"--prop",
		)
		require.NotEmpty(t, rows)
		assertSearchSkillsRowsContain(t, rows, slug, "1.0.0", tests.AgentSkillsLocalRepo)
	})

	t.Run("no-matches", func(t *testing.T) {
		query := "nonexistent-skill-xyzzy-abc123"
		searchArgs := []string{"search", query, "--repo=" + tests.AgentSkillsLocalRepo, "--format=json"}
		out, err := runAgentSkillsCmdWithOutput(t, searchArgs...)
		require.NoError(t, err)
		assert.Empty(t, strings.TrimSpace(out))

		_, logBuf, previousLog := coretests.RedirectLogOutputToBuffer()
		t.Cleanup(func() { log.SetLogger(previousLog) })
		require.NoError(t, runAgentSkillsCmd(t, searchArgs...))
		assert.Contains(t, logBuf.String(), fmt.Sprintf("No skills found matching '%s'.", query))
	})

	t.Run("empty-query", func(t *testing.T) {
		assertErrorContainsAll(t, runAgentSkillsCmd(t, "search", "--repo="+tests.AgentSkillsLocalRepo),
			"usage: jf agent skills search")
	})

	t.Run("blank-query", func(t *testing.T) {
		assertErrorContainsAll(t, runAgentSkillsCmd(t, "search", "   ", "--repo="+tests.AgentSkillsLocalRepo),
			"search query cannot be empty")
	})
}

// ---------------------------------------------------------------------------
// Config / repo resolution
// ---------------------------------------------------------------------------

// TestAgentSkillsRepoResolution covers --repo, env var, nonexistent repo, and
// wrong package type (plan scenarios 1, 2, 82, 83).
func TestAgentSkillsRepoResolution(t *testing.T) {
	initAgentSkillsTest(t)
	defer cleanAgentSkillsTest()

	slug := "repo-resolution-skill"
	publishTestSkill(t, slug, "1.0.0")

	t.Run("env-var", func(t *testing.T) {
		t.Setenv("JFROG_SKILLS_REPO", tests.AgentSkillsLocalRepo)
		installBase := t.TempDir()
		require.NoError(t, runAgentSkillsCmd(t,
			"install", slug,
			"--path="+installBase,
			"--version=1.0.0",
		))
		assert.FileExists(t, filepath.Join(installBase, slug, "SKILL.md"))
	})

	// Artifactory answers an unknown repo with 405 and no repo name in the body, so the
	// assertion stays on the wrapper the publish command adds.
	t.Run("nonexistent-repo", func(t *testing.T) {
		assertErrorContainsAll(t, runAgentSkillsCmd(t,
			"publish", createTestSkill(t, "bad-repo-skill", "1.0.0"),
			"--repo=does-not-exist-skills-repo",
			"--skip-scan",
		), "upload failed")
	})
}

// TestAgentSkillsMultiRepoCIFail verifies quiet/CI publish fails when multiple Skills
// local repos exist and neither --repo nor JFROG_SKILLS_REPO is set.
func TestAgentSkillsMultiRepoCIFail(t *testing.T) {
	initAgentSkillsTest(t)
	defer cleanAgentSkillsTest()

	// Create a second Skills local repo so discovery finds more than one key in quiet/CI.
	secondRepo := tests.AgentSkillsLocalRepo + "-extra"
	repoConfig := tests.GetTestResourcesPath() + tests.AgentSkillsLocalRepositoryConfig
	repoConfig, err := tests.ReplaceTemplateVariables(repoConfig, "")
	require.NoError(t, err)
	data, err := os.ReadFile(repoConfig) // #nosec G304 -- test template path
	require.NoError(t, err)
	rewritten := strings.Replace(string(data), tests.AgentSkillsLocalRepo, secondRepo, 1)
	tmpCfg := filepath.Join(t.TempDir(), "skills-extra-repo.json")
	// #nosec G306,G703 -- test fixture path is constrained to t.TempDir.
	require.NoError(t, os.WriteFile(tmpCfg, []byte(rewritten), 0644))
	execCreateRepoRest(tmpCfg, secondRepo)
	t.Cleanup(func() { execDeleteRepo(secondRepo) })
	require.True(t, isRepoExist(secondRepo))

	t.Setenv("CI", "true")
	t.Setenv("JFROG_SKILLS_REPO", "")
	err = runAgentSkillsCmd(t,
		"publish", createTestSkill(t, "multi-repo-skill", "1.0.0"),
		"--skip-scan",
		"--quiet",
	)
	assertErrorContainsAll(t, err,
		"multiple skills repositories found",
		tests.AgentSkillsLocalRepo,
		secondRepo,
	)
}

// ---------------------------------------------------------------------------
// Custom agent lifecycle
// ---------------------------------------------------------------------------

// TestAgentSkillsCustomAgentLifecycle covers skills-agents config override and
// install → list → check-updates → update (plan scenarios 6, 88).
func TestAgentSkillsCustomAgentLifecycle(t *testing.T) {
	initAgentSkillsTest(t)
	defer cleanAgentSkillsTest()

	const (
		agentName  = "my-agent"
		slug       = "my-agent-lifecycle-skill"
		oldVersion = "1.0.0"
		newVersion = "2.0.0"
	)

	publishTestSkill(t, slug, oldVersion)
	publishTestSkill(t, slug, newVersion)

	customGlobalDir := t.TempDir()
	agentTestutil.WriteAgentConfig(t, os.Getenv("JFROG_CLI_HOME_DIR"), `{
		"skills-agents": {
			"my-agent": {
				"globalDir": "`+filepath.ToSlash(customGlobalDir)+`",
				"projectDir": ".my-agent/skills"
			}
		}
	}`)

	require.NoError(t, runAgentSkillsCmd(t,
		"install", slug,
		"--repo="+tests.AgentSkillsLocalRepo,
		"--harness="+agentName,
		"--global",
		"--version="+oldVersion,
	))
	installDir := filepath.Join(customGlobalDir, slug)
	assertSkillInfoManifest(t, installDir, slug, oldVersion, agentName, "global", tests.AgentSkillsLocalRepo)

	listOut, err := runAgentSkillsCmdWithOutput(t,
		"list", "--harness="+agentName, "--global", "--format=json")
	require.NoError(t, err)
	var localRows []agentSkillsLocalListRow
	require.NoError(t, json.Unmarshal([]byte(extractJSONPayload(t, listOut)), &localRows))
	require.NotEmpty(t, localRows)
	assert.Equal(t, slug, localRows[0].Name)
	assert.Equal(t, oldVersion, localRows[0].Version)

	behindRow := waitForLocalListStatus(t, agentName, "", slug, "behind")
	assert.Equal(t, newVersion, behindRow.RegistryLatest)

	require.NoError(t, runAgentSkillsCmd(t,
		"update", slug,
		"--repo="+tests.AgentSkillsLocalRepo,
		"--harness="+agentName,
		"--global",
	))
	assertSkillInfoManifest(t, installDir, slug, newVersion, agentName, "global", tests.AgentSkillsLocalRepo)
	currentRow := waitForLocalListStatus(t, agentName, "", slug, "current")
	assert.Equal(t, newVersion, currentRow.RegistryLatest)
}

// ---------------------------------------------------------------------------
// Build info
// ---------------------------------------------------------------------------

// TestAgentSkillsPublishBuildInfo covers publish build-info capture and artifact
// build props (plan scenarios 64–66).
func TestAgentSkillsPublishBuildInfo(t *testing.T) {
	initAgentSkillsTest(t)
	defer cleanAgentSkillsTest()

	slug := "buildinfo-skill"
	version := "1.0.0"
	skillPath := createTestSkill(t, slug, version)
	buildName := tests.AgentSkillsBuildName
	buildNumber := t.Name()
	// Best-effort teardown of local build-info dir; leftover dirs do not affect assertions.
	t.Cleanup(func() { _ = coreBuild.RemoveBuildDir(buildName, buildNumber, "") })

	require.NoError(t, runAgentSkillsCmd(t,
		"publish", skillPath,
		"--repo="+tests.AgentSkillsLocalRepo,
		"--skip-scan",
		"--build-name="+buildName,
		"--build-number="+buildNumber,
	))
	assertSkillExists(t, slug, version)
	waitForSkillIndexed(t, slug, version)

	require.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found)
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules)
	var moduleFound bool
	for _, module := range publishedBuildInfo.BuildInfo.Modules {
		if module.Id != slug {
			continue
		}
		moduleFound = true
		require.NotEmpty(t, module.Artifacts)
		assert.NotEmpty(t, module.Artifacts[0].Sha256)
		assert.NotEmpty(t, module.Artifacts[0].Md5)
		assert.NotEmpty(t, module.Artifacts[0].Sha1)
		break
	}
	require.True(t, moduleFound, "build-info must contain module id %q", slug)

	sm, err := artUtils.CreateServiceManager(serverDetails, -1, 0, false)
	require.NoError(t, err)
	props, err := sm.GetItemProps(skillArtifactPath(tests.AgentSkillsLocalRepo, slug, version))
	require.NoError(t, err)
	require.NotNil(t, props, "published artifact must carry build properties")
	require.Contains(t, props.Properties, "build.name")
	require.Contains(t, props.Properties, "build.number")
	require.Contains(t, props.Properties, "build.timestamp")
	assert.Contains(t, props.Properties["build.name"], buildName)
	assert.Contains(t, props.Properties["build.number"], buildNumber)
	assert.NotEmpty(t, props.Properties["build.timestamp"])
}

// ---------------------------------------------------------------------------
// Round-trip / CI
// ---------------------------------------------------------------------------

// TestAgentSkillsRoundTrip verifies publish → install preserves SKILL.md bytes
// (plan scenario 86).
func TestAgentSkillsRoundTrip(t *testing.T) {
	initAgentSkillsTest(t)
	defer cleanAgentSkillsTest()

	slug := "roundtrip-skill"
	version := "1.0.0"
	skillPath := createTestSkill(t, slug, version)
	sourceMD, err := os.ReadFile(filepath.Join(skillPath, "SKILL.md")) // #nosec G304
	require.NoError(t, err)

	require.NoError(t, runAgentSkillsCmd(t,
		"publish", skillPath,
		"--repo="+tests.AgentSkillsLocalRepo,
		"--skip-scan",
	))
	assertSkillExists(t, slug, version)
	waitForSkillIndexed(t, slug, version)

	installBase := t.TempDir()
	require.NoError(t, runAgentSkillsCmd(t,
		"install", slug,
		"--repo="+tests.AgentSkillsLocalRepo,
		"--path="+installBase,
		"--version="+version,
	))
	installedMD, err := os.ReadFile(filepath.Join(installBase, slug, "SKILL.md")) // #nosec G304
	require.NoError(t, err)
	assert.Equal(t, string(sourceMD), string(installedMD))
}

// TestAgentSkillsCIPipeline covers quiet CI publish(+bp) → install → list → update
// (plan scenario 102).
func TestAgentSkillsCIPipeline(t *testing.T) {
	initAgentSkillsTest(t)
	defer cleanAgentSkillsTest()

	t.Setenv("CI", "true")
	t.Setenv("JFROG_SKILLS_DISABLE_QUIET_FAILURE", "true")

	slug := "ci-pipeline-skill"
	buildName := tests.AgentSkillsBuildName
	buildNumber := t.Name()
	// Best-effort teardown of local build-info dir; leftover dirs do not affect assertions.
	t.Cleanup(func() { _ = coreBuild.RemoveBuildDir(buildName, buildNumber, "") })

	skillPath := createTestSkill(t, slug, "1.0.0")
	require.NoError(t, runAgentSkillsCmd(t,
		"publish", skillPath,
		"--repo="+tests.AgentSkillsLocalRepo,
		"--skip-scan",
		"--quiet",
		"--build-name="+buildName,
		"--build-number="+buildNumber,
	))
	assertSkillExists(t, slug, "1.0.0")
	waitForSkillIndexed(t, slug, "1.0.0")
	require.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found)
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules)

	publishTestSkill(t, slug, "2.0.0")

	projectDir := t.TempDir()
	require.NoError(t, runAgentSkillsCmd(t,
		"install", slug,
		"--repo="+tests.AgentSkillsLocalRepo,
		"--harness=cursor",
		"--project-dir="+projectDir,
		"--version=1.0.0",
		"--quiet",
	))
	out, err := runAgentSkillsCmdWithOutput(t,
		"list", "--harness=cursor", "--project-dir="+projectDir, "--format=json")
	require.NoError(t, err)
	assert.Contains(t, out, slug)

	require.NoError(t, runAgentSkillsCmd(t,
		"update", slug,
		"--repo="+tests.AgentSkillsLocalRepo,
		"--harness=cursor",
		"--project-dir="+projectDir,
		"--quiet",
	))
	assertSkillInfoManifest(t, filepath.Join(projectDir, ".cursor", "skills", slug),
		slug, "2.0.0", "cursor", "project", tests.AgentSkillsLocalRepo)
}

// ---------------------------------------------------------------------------
// Flag validation
// ---------------------------------------------------------------------------

// TestAgentSkillsFlagValidation covers unknown flags, missing args, and format fallback
// (plan scenarios 77–78). Unknown --format values fall back to table, matching plugins.
func TestAgentSkillsFlagValidation(t *testing.T) {
	initAgentSkillsTest(t)
	defer cleanAgentSkillsTest()

	t.Run("invalid-format-falls-back", func(t *testing.T) {
		// Skills list treats unknown --format values as table output (same as plugins).
		assert.NoError(t, runAgentSkillsCmd(t,
			"list", "--repo="+tests.AgentSkillsLocalRepo, "--format=xml"))
	})

	t.Run("unknown-flag", func(t *testing.T) {
		for _, subcommand := range []string{"publish", "install", "update", "delete", "list", "search"} {
			t.Run(subcommand, func(t *testing.T) {
				assert.Error(t, runAgentSkillsCmd(t, subcommand, "--this-flag-does-not-exist=xyz"),
					"subcommand %q must reject unknown flags", subcommand)
			})
		}
	})

	t.Run("missing-install-slug", func(t *testing.T) {
		assertErrorContainsAll(t, runAgentSkillsCmd(t,
			"install", "--repo="+tests.AgentSkillsLocalRepo, "--path="+t.TempDir()),
			"usage: jf agent skills install")
	})
}
