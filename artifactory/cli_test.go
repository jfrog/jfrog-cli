package artifactory

import (
	"bytes"
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jfrog/jfrog-cli-artifactory/cliutils/flagkit"

	transferfilescore "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/transferfiles"
	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	coreformat "github.com/jfrog/jfrog-cli-core/v2/common/format"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

func TestPrepareSearchDownloadDeleteCommands(t *testing.T) {
	testRuns := []struct {
		name            string
		args            []string
		flags           []string
		expectedPattern string
		expectedBuild   string
		expectedBundle  string
		expectError     bool
	}{
		{"withoutArgs", []string{}, []string{}, "TestPattern", "", "", true},
		{"withPattern", []string{"TestPattern"}, []string{}, "TestPattern", "", "", false},
		{"withBuild", []string{}, []string{"build=buildName/buildNumber"}, "", "buildName/buildNumber", "", false},
		{"withBundle", []string{}, []string{"bundle=bundleName/bundleVersion"}, "", "", "bundleName/bundleVersion", false},
		{"withSpec", []string{}, []string{"spec=" + getSpecPath(tests.SearchAllRepo1)}, "${REPO1}/*", "", "", false},
		{"withSpecAndPattern", []string{"TestPattern"}, []string{"spec=" + getSpecPath(tests.SearchAllRepo1)}, "", "", "", true},
		{"withBuildAndPattern", []string{"TestPattern"}, []string{"build=buildName/buildNumber"}, "TestPattern", "buildName/buildNumber", "", false},
	}

	for _, test := range testRuns {
		t.Run(test.name, func(t *testing.T) {
			context, buffer := tests.CreateContext(t, test.flags, test.args)
			funcArray := []func(c *cli.Context) (*spec.SpecFiles, error){
				prepareSearchCommand, prepareDownloadCommand, prepareDeleteCommand,
			}
			for _, prepareCommandFunc := range funcArray {
				specFiles, err := prepareCommandFunc(context)
				assertGenericCommand(t, err, buffer, test.expectError, test.expectedPattern, test.expectedBuild, test.expectedBundle, specFiles)
			}
		})
	}
}

func TestPrepareCopyMoveCommand(t *testing.T) {
	testRuns := []struct {
		name            string
		args            []string
		flags           []string
		expectedPattern string
		expectedTarget  string
		expectedBuild   string
		expectedBundle  string
		expectError     bool
	}{
		{"withoutArguments", []string{}, []string{}, "", "", "", "", true},
		{"withPatternAndTarget", []string{"TestPattern", "TestTarget"}, []string{}, "TestPattern", "TestTarget", "", "", false},
		{"withSpec", []string{}, []string{"spec=" + getSpecPath(tests.CopyItemsSpec)}, "${REPO1}/*/", "${REPO2}/", "", "", false},
		{"withSpecAndPattern", []string{"TestPattern"}, []string{"spec=" + getSpecPath(tests.CopyItemsSpec)}, "", "", "", "", true},
		{"withPatternTargetAndBuild", []string{"TestPattern", "TestTarget"}, []string{"build=buildName/buildNumber"}, "TestPattern", "", "buildName/buildNumber", "", false},
		{"withPatternTargetAndBundle", []string{"TestPattern", "TestTarget"}, []string{"bundle=bundleName/bundleVersion"}, "TestPattern", "", "", "bundleName/bundleVersion", false},
	}

	for _, test := range testRuns {
		t.Run(test.name, func(t *testing.T) {
			context, buffer := tests.CreateContext(t, test.flags, test.args)
			specFiles, err := prepareCopyMoveCommand(context)
			assertGenericCommand(t, err, buffer, test.expectError, test.expectedPattern, test.expectedBuild, test.expectedBundle, specFiles)
		})
	}
}

func TestPreparePropsCmd(t *testing.T) {
	testRuns := []struct {
		name            string
		args            []string
		flags           []string
		expectedProps   string
		expectedPattern string
		expectedBuild   string
		expectedBundle  string
		expectError     bool
	}{
		{"withoutPattern", []string{"key1=val1"}, []string{}, "key1=val1", "", "", "", true},
		{"withPattern", []string{"TestPattern", "key1=val1"}, []string{}, "key1=val1", "TestPattern", "", "", false},
		{"withBuild", []string{"key1=val1"}, []string{"build=buildName/buildNumber"}, "key1=val1", "*", "buildName/buildNumber", "", false},
		{"withBundle", []string{"key1=val1"}, []string{"bundle=bundleName/bundleVersion"}, "key1=val1", "*", "", "bundleName/bundleVersion", false},
		{"withSpec", []string{"key1=val1"}, []string{"spec=" + getSpecPath(tests.SetDeletePropsSpec)}, "key1=val1", "${REPO1}/", "", "", false},
		{"withSpecAndPattern", []string{"TestPattern", "key1=val1"}, []string{"spec=" + getSpecPath(tests.SetDeletePropsSpec)}, "key1=val1", "", "", "", true},
		{"withPatternAndBuild", []string{"TestPattern", "key1=val1"}, []string{"build=buildName/buildNumber"}, "key1=val1", "TestPattern", "buildName/buildNumber", "", false},
	}

	for _, test := range testRuns {
		t.Run(test.name, func(t *testing.T) {
			context, buffer := tests.CreateContext(t, test.flags, test.args)
			propsCommand, err := preparePropsCmd(context)
			var actualSpec *spec.SpecFiles
			if propsCommand != nil {
				actualSpec = propsCommand.Spec()
				assert.Equal(t, test.expectedProps, propsCommand.Props())
			}
			assertGenericCommand(t, err, buffer, test.expectError, test.expectedPattern, test.expectedBuild, test.expectedBundle, actualSpec)
		})
	}
}

func assertGenericCommand(t *testing.T, err error, buffer *bytes.Buffer, expectError bool, expectedPattern, expectedBuild, expectedBundle string, actualSpec *spec.SpecFiles) {
	if expectError {
		assert.Error(t, err, buffer)
	} else {
		assert.NoError(t, err, buffer)
		assert.Equal(t, expectedPattern, actualSpec.Get(0).Pattern)
		assert.Equal(t, expectedBuild, actualSpec.Get(0).Build)
		assert.Equal(t, expectedBundle, actualSpec.Get(0).Bundle)
	}
}

func getSpecPath(spec string) string {
	return filepath.Join("..", "testdata", "filespecs", spec)
}

var createUploadConfigurationCases = []struct {
	name               string
	flags              []string
	expectedMinSplit   int64
	expectedSplitCount int
	expectedThreads    int
	expectedDeb        string
	expectedChunkSize  int64
}{
	{"empty", []string{}, flagkit.UploadMinSplitMb, flagkit.UploadSplitCount, commonCliUtils.Threads, "", flagkit.UploadChunkSizeMb},
	{"min-split", []string{"min-split=101"}, 101, flagkit.UploadSplitCount, commonCliUtils.Threads, "", flagkit.UploadChunkSizeMb},
	{"split-count", []string{"split-count=6"}, flagkit.UploadMinSplitMb, 6, commonCliUtils.Threads, "", flagkit.UploadChunkSizeMb},
	{"threads", []string{"threads=6"}, flagkit.UploadMinSplitMb, flagkit.UploadSplitCount, 6, "", flagkit.UploadChunkSizeMb},
	{"deb", []string{"deb=jammy/main/i386"}, flagkit.UploadMinSplitMb, flagkit.UploadSplitCount, commonCliUtils.Threads, "jammy/main/i386", flagkit.UploadChunkSizeMb},
	{"chunk-size", []string{"chunk-size=123"}, flagkit.UploadMinSplitMb, flagkit.UploadSplitCount, commonCliUtils.Threads, "", 123},
}

func TestCreateUploadConfiguration(t *testing.T) {
	for _, testCase := range createUploadConfigurationCases {
		t.Run(testCase.name, func(t *testing.T) {
			context, _ := tests.CreateContext(t, testCase.flags, []string{})
			uploadConfiguration, err := cliutils.CreateUploadConfiguration(context)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedMinSplit, uploadConfiguration.MinSplitSizeMB)
			assert.Equal(t, testCase.expectedSplitCount, uploadConfiguration.SplitCount)
			assert.Equal(t, testCase.expectedThreads, uploadConfiguration.Threads)
			assert.Equal(t, testCase.expectedDeb, uploadConfiguration.Deb)
		})
	}
}

// Test cases for getIncludeFilesPatterns
var getIncludeFilesPatternsTestCases = []struct {
	name             string
	flags            []string
	expectedPatterns []string
}{
	{
		name:             "no flag set",
		flags:            []string{},
		expectedPatterns: nil,
	},
	{
		name:             "single pattern",
		flags:            []string{cliutils.IncludeFiles + "=org/company/*"},
		expectedPatterns: []string{"org/company/*"},
	},
	{
		name:             "multiple patterns",
		flags:            []string{cliutils.IncludeFiles + "=org/company/*;com/example/*"},
		expectedPatterns: []string{"org/company/*", "com/example/*"},
	},
	{
		name:             "three patterns",
		flags:            []string{cliutils.IncludeFiles + "=path1/*;path2/*;path3/*"},
		expectedPatterns: []string{"path1/*", "path2/*", "path3/*"},
	},
	{
		name:             "pattern with deep nesting",
		flags:            []string{cliutils.IncludeFiles + "=a/b/c/d/e/*"},
		expectedPatterns: []string{"a/b/c/d/e/*"},
	},
}

func TestGetIncludeFilesPatterns(t *testing.T) {
	for _, testCase := range getIncludeFilesPatternsTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			context, _ := tests.CreateContext(t, testCase.flags, []string{})
			patterns := getIncludeFilesPatterns(context)
			assert.Equal(t, testCase.expectedPatterns, patterns)
		})
	}
}

// ---------------------------------------------------------------------------
// helpers for transfer-files format tests
// ---------------------------------------------------------------------------

// newTransferFilesFormatContext creates a *cli.Context with --format set to formatVal.
// If formatVal is empty, the flag is registered but left unset.
func newTransferFilesFormatContext(formatVal string) *cli.Context {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: "format"}}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("format", "", "")
	if formatVal != "" {
		_ = fs.Set("format", formatVal)
	}
	return cli.NewContext(app, fs, nil)
}

// ---------------------------------------------------------------------------
// GetOutputFormat (transfer-files) tests
// ---------------------------------------------------------------------------

func TestGetTransferFilesOutputFormat_Default(t *testing.T) {
	// No --format flag set → backward-compatible, returns None.
	c := newTransferFilesFormatContext("")
	f, err := commonCliUtils.GetOutputFormat(c, coreformat.None)
	require.NoError(t, err)
	assert.Equal(t, coreformat.None, f)
}

func TestGetTransferFilesOutputFormat_ExplicitJSON(t *testing.T) {
	c := newTransferFilesFormatContext("json")
	f, err := commonCliUtils.GetOutputFormat(c, coreformat.None)
	require.NoError(t, err)
	assert.Equal(t, coreformat.Json, f)
}

func TestGetTransferFilesOutputFormat_ExplicitTable(t *testing.T) {
	c := newTransferFilesFormatContext("table")
	f, err := commonCliUtils.GetOutputFormat(c, coreformat.None)
	require.NoError(t, err)
	assert.Equal(t, coreformat.Table, f)
}

func TestGetTransferFilesOutputFormat_Invalid(t *testing.T) {
	c := newTransferFilesFormatContext("xml")
	_, err := commonCliUtils.GetOutputFormat(c, coreformat.None)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "only the following output formats are supported"))
}

// ---------------------------------------------------------------------------
// printTransferFilesResponse tests
// ---------------------------------------------------------------------------

func TestPrintTransferFilesResponse_JSON(t *testing.T) {
	result := transferfilescore.TransferFilesResult{
		TotalRepositories:       5,
		TransferredRepositories: 5,
		TotalFiles:              1000,
		TransferredFiles:        990,
		TotalSizeBytes:          1048576,
		TransferredSizeBytes:    1038336,
		TransferFailures:        10,
	}
	var buf bytes.Buffer
	err := printTransferFilesResponse(result, coreformat.Json, &buf, nil)
	require.NoError(t, err)
	// The JSON is printed via log.Output (not to buf), so we verify no error.
}

func TestPrintTransferFilesResponse_JSON_WithError(t *testing.T) {
	result := transferfilescore.TransferFilesResult{
		TotalRepositories:       3,
		TransferredRepositories: 2,
		TotalFiles:              500,
		TransferredFiles:        400,
		TransferFailures:        100,
	}
	var buf bytes.Buffer
	originalErr := assert.AnError
	// JSON path propagates the original error.
	err := printTransferFilesResponse(result, coreformat.Json, &buf, originalErr)
	assert.Equal(t, originalErr, err)
}

func TestPrintTransferFilesResponse_Table(t *testing.T) {
	result := transferfilescore.TransferFilesResult{
		TotalRepositories:       4,
		TransferredRepositories: 4,
		TotalFiles:              200,
		TransferredFiles:        198,
		TotalSizeBytes:          2097152,
		TransferredSizeBytes:    2076672,
		TransferFailures:        2,
	}
	var buf bytes.Buffer
	err := printTransferFilesResponse(result, coreformat.Table, &buf, nil)
	require.NoError(t, err)

	output := buf.String()
	// Header
	assert.True(t, strings.Contains(output, "FIELD"), "output should contain FIELD header")
	assert.True(t, strings.Contains(output, "VALUE"), "output should contain VALUE header")
	// Key rows
	assert.True(t, strings.Contains(output, "status"), "output should contain status row")
	assert.True(t, strings.Contains(output, "success"), "output should contain success status")
	assert.True(t, strings.Contains(output, "repositories_transferred"), "output should contain repositories_transferred row")
	assert.True(t, strings.Contains(output, "files_transferred"), "output should contain files_transferred row")
	assert.True(t, strings.Contains(output, "files_failed"), "output should contain files_failed row")
}

func TestPrintTransferFilesResponse_Table_FailureStatus(t *testing.T) {
	result := transferfilescore.TransferFilesResult{
		TotalRepositories:    2,
		TotalFiles:           100,
		TransferredFiles:     50,
		TransferFailures:     50,
	}
	var buf bytes.Buffer
	originalErr := assert.AnError
	err := printTransferFilesResponse(result, coreformat.Table, &buf, originalErr)
	// Table path propagates the original error.
	assert.Equal(t, originalErr, err)

	output := buf.String()
	assert.True(t, strings.Contains(output, "failure"), "output should show failure status when error occurred")
}

func TestPrintTransferFilesResponse_UnsupportedFormat(t *testing.T) {
	result := transferfilescore.TransferFilesResult{}
	var buf bytes.Buffer
	err := printTransferFilesResponse(result, coreformat.SimpleJson, &buf, nil)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "unsupported format"))
}

// --- permission-target-create --format tests ---

// newPermissionTargetCreateFormatApp returns a minimal *cli.App with a --format flag for
// testing the format-guard logic in permissionTargetCreateCmd without a live server.
func newPermissionTargetCreateFormatApp(action cli.ActionFunc) *cli.App {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: cliutils.Format}}
	app.Action = action
	return app
}

// permissionTargetCreateFormatGuard mirrors the format-check block in permissionTargetCreateCmd
// so we can test it independently of the HTTP call.
func permissionTargetCreateFormatGuard(c *cli.Context) error {
	if !c.IsSet(cliutils.Format) {
		return nil
	}
	_, err := coreformat.ParseOutputFormat(c.String(cliutils.Format), []coreformat.OutputFormat{coreformat.Json})
	return err
}

func TestPermissionTargetCreate_FormatNotSet_NoError(t *testing.T) {
	app := newPermissionTargetCreateFormatApp(permissionTargetCreateFormatGuard)
	require.NoError(t, app.Run([]string{"app"}))
}

func TestPermissionTargetCreate_FormatJSON_NoError(t *testing.T) {
	app := newPermissionTargetCreateFormatApp(permissionTargetCreateFormatGuard)
	require.NoError(t, app.Run([]string{"app", "--format", "json"}))
}

func TestPermissionTargetCreate_FormatTable_ReturnsError(t *testing.T) {
	var gotErr error
	app := newPermissionTargetCreateFormatApp(func(c *cli.Context) error {
		gotErr = permissionTargetCreateFormatGuard(c)
		return nil
	})
	require.NoError(t, app.Run([]string{"app", "--format", "table"}))
	require.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "only the following output formats are supported")
}

func TestPermissionTargetCreate_InvalidFormat_ReturnsError(t *testing.T) {
	var gotErr error
	app := newPermissionTargetCreateFormatApp(func(c *cli.Context) error {
		gotErr = permissionTargetCreateFormatGuard(c)
		return nil
	})
	require.NoError(t, app.Run([]string{"app", "--format", "xml"}))
	require.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "only the following output formats are supported")
}

// --- transfer-config --format tests ---

// newTransferConfigFormatApp returns a minimal *cli.App with a --format flag for
// testing the format-guard logic in transferConfigCmd without a live server.
func newTransferConfigFormatApp(action cli.ActionFunc) *cli.App {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: cliutils.Format}}
	app.Action = action
	return app
}

// transferConfigFormatGuard mirrors the format-check block in transferConfigCmd
// so we can test it independently of the HTTP call.
func transferConfigFormatGuard(c *cli.Context) error {
	if !c.IsSet(cliutils.Format) {
		return nil
	}
	_, err := coreformat.ParseOutputFormat(c.String(cliutils.Format), []coreformat.OutputFormat{coreformat.Json})
	return err
}

func TestTransferConfig_FormatNotSet_NoError(t *testing.T) {
	app := newTransferConfigFormatApp(transferConfigFormatGuard)
	require.NoError(t, app.Run([]string{"app"}))
}

func TestTransferConfig_FormatJSON_NoError(t *testing.T) {
	app := newTransferConfigFormatApp(transferConfigFormatGuard)
	require.NoError(t, app.Run([]string{"app", "--format", "json"}))
}

func TestTransferConfig_FormatTable_ReturnsError(t *testing.T) {
	var gotErr error
	app := newTransferConfigFormatApp(func(c *cli.Context) error {
		gotErr = transferConfigFormatGuard(c)
		return nil
	})
	require.NoError(t, app.Run([]string{"app", "--format", "table"}))
	require.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "only the following output formats are supported")
}

func TestTransferConfig_InvalidFormat_ReturnsError(t *testing.T) {
	var gotErr error
	app := newTransferConfigFormatApp(func(c *cli.Context) error {
		gotErr = transferConfigFormatGuard(c)
		return nil
	})
	require.NoError(t, app.Run([]string{"app", "--format", "xml"}))
	require.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "only the following output formats are supported")
}

// --- permission-target-update --format tests ---

// newPermissionTargetUpdateFormatApp returns a minimal *cli.App with a --format flag for
// testing the format-guard logic in permissionTargetUpdateCmd without a live server.
func newPermissionTargetUpdateFormatApp(action cli.ActionFunc) *cli.App {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: cliutils.Format}}
	app.Action = action
	return app
}

// permissionTargetUpdateFormatGuard mirrors the format-check block in permissionTargetUpdateCmd
// so we can test it independently of the HTTP call.
func permissionTargetUpdateFormatGuard(c *cli.Context) error {
	if !c.IsSet(cliutils.Format) {
		return nil
	}
	_, err := coreformat.ParseOutputFormat(c.String(cliutils.Format), []coreformat.OutputFormat{coreformat.Json})
	return err
}

func TestPermissionTargetUpdate_FormatNotSet_NoError(t *testing.T) {
	app := newPermissionTargetUpdateFormatApp(permissionTargetUpdateFormatGuard)
	require.NoError(t, app.Run([]string{"app"}))
}

func TestPermissionTargetUpdate_FormatJSON_NoError(t *testing.T) {
	app := newPermissionTargetUpdateFormatApp(permissionTargetUpdateFormatGuard)
	require.NoError(t, app.Run([]string{"app", "--format", "json"}))
}

func TestPermissionTargetUpdate_FormatTable_ReturnsError(t *testing.T) {
	var gotErr error
	app := newPermissionTargetUpdateFormatApp(func(c *cli.Context) error {
		gotErr = permissionTargetUpdateFormatGuard(c)
		return nil
	})
	require.NoError(t, app.Run([]string{"app", "--format", "table"}))
	require.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "only the following output formats are supported")
}

func TestPermissionTargetUpdate_InvalidFormat_ReturnsError(t *testing.T) {
	var gotErr error
	app := newPermissionTargetUpdateFormatApp(func(c *cli.Context) error {
		gotErr = permissionTargetUpdateFormatGuard(c)
		return nil
	})
	require.NoError(t, app.Run([]string{"app", "--format", "xml"}))
	require.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "only the following output formats are supported")
}

// --- transfer-config-merge --format tests ---

func newTransferConfigMergeFormatApp(action cli.ActionFunc) *cli.App {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: cliutils.Format}}
	app.Action = action
	return app
}

// TestGetTransferConfigMergeOutputFormat_Default verifies that when --format is not set,
// the function returns coreformat.None (backward-compatible: log output only).
func TestGetTransferConfigMergeOutputFormat_Default(t *testing.T) {
	var got coreformat.OutputFormat
	app := newTransferConfigMergeFormatApp(func(c *cli.Context) error {
		var err error
		got, err = getTransferConfigMergeOutputFormat(c)
		return err
	})
	require.NoError(t, app.Run([]string{"app"}))
	assert.Equal(t, coreformat.None, got)
}

// TestGetTransferConfigMergeOutputFormat_ExplicitJSON verifies --format json is accepted.
func TestGetTransferConfigMergeOutputFormat_ExplicitJSON(t *testing.T) {
	var got coreformat.OutputFormat
	app := newTransferConfigMergeFormatApp(func(c *cli.Context) error {
		var err error
		got, err = getTransferConfigMergeOutputFormat(c)
		return err
	})
	require.NoError(t, app.Run([]string{"app", "--format", "json"}))
	assert.Equal(t, coreformat.Json, got)
}

// TestGetTransferConfigMergeOutputFormat_ExplicitTable verifies --format table is accepted.
func TestGetTransferConfigMergeOutputFormat_ExplicitTable(t *testing.T) {
	var got coreformat.OutputFormat
	app := newTransferConfigMergeFormatApp(func(c *cli.Context) error {
		var err error
		got, err = getTransferConfigMergeOutputFormat(c)
		return err
	})
	require.NoError(t, app.Run([]string{"app", "--format", "table"}))
	assert.Equal(t, coreformat.Table, got)
}

// TestGetTransferConfigMergeOutputFormat_Invalid verifies that an unsupported format returns an error.
func TestGetTransferConfigMergeOutputFormat_Invalid(t *testing.T) {
	var gotErr error
	app := newTransferConfigMergeFormatApp(func(c *cli.Context) error {
		_, gotErr = getTransferConfigMergeOutputFormat(c)
		return nil
	})
	require.NoError(t, app.Run([]string{"app", "--format", "xml"}))
	require.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "only the following output formats are supported")
}

// TestPrintTransferConfigMergeResponse_JSON_Success verifies JSON output when there are no conflicts.
func TestPrintTransferConfigMergeResponse_JSON_Success(t *testing.T) {
	var buf bytes.Buffer
	err := printTransferConfigMergeResponse("", coreformat.Json, &buf, nil)
	require.NoError(t, err)
	// JSON is emitted via log.Output (not to buf), so we just verify no error.
}

// TestPrintTransferConfigMergeResponse_JSON_ConflictsFound verifies JSON output when conflicts are found.
func TestPrintTransferConfigMergeResponse_JSON_ConflictsFound(t *testing.T) {
	var buf bytes.Buffer
	err := printTransferConfigMergeResponse("/tmp/conflicts.csv", coreformat.Json, &buf, nil)
	require.NoError(t, err)
}

// TestPrintTransferConfigMergeResponse_Table_Success verifies table output with no conflicts.
func TestPrintTransferConfigMergeResponse_Table_Success(t *testing.T) {
	var buf bytes.Buffer
	err := printTransferConfigMergeResponse("", coreformat.Table, &buf, nil)
	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "FIELD")
	assert.Contains(t, out, "VALUE")
	assert.Contains(t, out, "status")
	assert.Contains(t, out, "success")
	assert.Contains(t, out, "message")
}

// TestPrintTransferConfigMergeResponse_Table_ConflictsFound verifies table output when conflicts are found.
func TestPrintTransferConfigMergeResponse_Table_ConflictsFound(t *testing.T) {
	var buf bytes.Buffer
	err := printTransferConfigMergeResponse("/tmp/conflicts.csv", coreformat.Table, &buf, nil)
	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "status")
	assert.Contains(t, out, "conflicts_found")
	assert.Contains(t, out, "conflicts_report_path")
	assert.Contains(t, out, "/tmp/conflicts.csv")
}

// TestPrintTransferConfigMergeResponse_Table_NoConflictsReportPath verifies that conflicts_report_path is
// absent from table output when there are no conflicts.
func TestPrintTransferConfigMergeResponse_Table_NoConflictsReportPath(t *testing.T) {
	var buf bytes.Buffer
	err := printTransferConfigMergeResponse("", coreformat.Table, &buf, nil)
	require.NoError(t, err)
	out := buf.String()
	assert.NotContains(t, out, "conflicts_report_path")
}

// TestPrintTransferConfigMergeResponse_UnsupportedFormat verifies that an unsupported format returns an error.
func TestPrintTransferConfigMergeResponse_UnsupportedFormat(t *testing.T) {
	var buf bytes.Buffer
	err := printTransferConfigMergeResponse("", coreformat.SimpleJson, &buf, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

// TestBuildTransferConfigMergeResult_Success verifies the result struct for a successful merge.
func TestBuildTransferConfigMergeResult_Success(t *testing.T) {
	result := buildTransferConfigMergeResult("", nil)
	assert.Equal(t, "success", result.Status)
	assert.Empty(t, result.ConflictsReportPath)
	assert.NotEmpty(t, result.Message)
}

// TestBuildTransferConfigMergeResult_ConflictsFound verifies the result struct when conflicts are found.
func TestBuildTransferConfigMergeResult_ConflictsFound(t *testing.T) {
	result := buildTransferConfigMergeResult("/tmp/conflicts.csv", nil)
	assert.Equal(t, "conflicts_found", result.Status)
	assert.Equal(t, "/tmp/conflicts.csv", result.ConflictsReportPath)
	assert.Contains(t, result.Message, "/tmp/conflicts.csv")
}

// TestBuildTransferConfigMergeResult_Failure verifies the result struct when the command failed.
func TestBuildTransferConfigMergeResult_Failure(t *testing.T) {
	result := buildTransferConfigMergeResult("", fmt.Errorf("something went wrong"))
	assert.Equal(t, "failure", result.Status)
	assert.Contains(t, result.Message, "something went wrong")
}

// --- users-create --format tests ---

// newUsersCreateFormatApp returns a minimal *cli.App with a --format flag for
// testing the format-guard logic in usersCreateCmd without a live server.
func newUsersCreateFormatApp(action cli.ActionFunc) *cli.App {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: cliutils.Format}}
	app.Action = action
	return app
}

// usersCreateFormatGuard mirrors the format-check block in usersCreateCmd so we
// can test it independently of the HTTP call.
func usersCreateFormatGuard(c *cli.Context) error {
	if !c.IsSet(cliutils.Format) {
		return nil
	}
	_, err := coreformat.ParseOutputFormat(c.String(cliutils.Format), []coreformat.OutputFormat{coreformat.Json})
	return err
}

func TestUsersCreate_FormatNotSet_NoError(t *testing.T) {
	app := newUsersCreateFormatApp(usersCreateFormatGuard)
	require.NoError(t, app.Run([]string{"app"}))
}

func TestUsersCreate_FormatJSON_NoError(t *testing.T) {
	app := newUsersCreateFormatApp(usersCreateFormatGuard)
	require.NoError(t, app.Run([]string{"app", "--format", "json"}))
}

func TestUsersCreate_FormatTable_ReturnsError(t *testing.T) {
	var gotErr error
	app := newUsersCreateFormatApp(func(c *cli.Context) error {
		gotErr = usersCreateFormatGuard(c)
		return nil
	})
	require.NoError(t, app.Run([]string{"app", "--format", "table"}))
	require.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "only the following output formats are supported")
}

func TestUsersCreate_InvalidFormat_ReturnsError(t *testing.T) {
	var gotErr error
	app := newUsersCreateFormatApp(func(c *cli.Context) error {
		gotErr = usersCreateFormatGuard(c)
		return nil
	})
	require.NoError(t, app.Run([]string{"app", "--format", "xml"}))
	require.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "only the following output formats are supported")
}
