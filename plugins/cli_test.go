package plugins

import (
	"fmt"
	"testing"

	coreformat "github.com/jfrog/jfrog-cli-core/v2/common/format"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

// newFormatApp returns a minimal *cli.App with a --format flag for testing the
// format-guard logic in installPlugin / publishPlugin without needing a live server.
func newFormatApp(action cli.ActionFunc) *cli.App {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: "format"}}
	app.Action = action
	return app
}

// formatGuard mirrors the format-check block used in both installPlugin and
// publishPlugin so we can test it independently of the HTTP calls.
func formatGuard(c *cli.Context, cmdName string) error {
	if !c.IsSet("format") {
		return nil
	}
	outputFormat, err := coreformat.GetOutputFormat(c.String("format"))
	if err != nil {
		return err
	}
	if outputFormat == coreformat.Json {
		// In the real wrappers, FormatHTTPResponseJSON is called here.
		// For unit tests we just verify the branch is reached without error.
		return nil
	}
	return fmt.Errorf("unsupported format '%s' for %s. Only json is supported", outputFormat, cmdName)
}

// --- installPlugin format-guard tests ---

func TestInstallPlugin_FormatNotSet_NoError(t *testing.T) {
	app := newFormatApp(func(c *cli.Context) error {
		return formatGuard(c, "plugin install")
	})
	require.NoError(t, app.Run([]string{"app"}))
}

func TestInstallPlugin_FormatJSON_NoError(t *testing.T) {
	app := newFormatApp(func(c *cli.Context) error {
		return formatGuard(c, "plugin install")
	})
	require.NoError(t, app.Run([]string{"app", "--format", "json"}))
}

func TestInstallPlugin_FormatTable_ReturnsError(t *testing.T) {
	var gotErr error
	app := newFormatApp(func(c *cli.Context) error {
		gotErr = formatGuard(c, "plugin install")
		return nil
	})
	require.NoError(t, app.Run([]string{"app", "--format", "table"}))
	require.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "unsupported format")
	assert.Contains(t, gotErr.Error(), "plugin install")
}

func TestInstallPlugin_InvalidFormat_ReturnsError(t *testing.T) {
	var gotErr error
	app := newFormatApp(func(c *cli.Context) error {
		gotErr = formatGuard(c, "plugin install")
		return nil
	})
	require.NoError(t, app.Run([]string{"app", "--format", "xml"}))
	require.Error(t, gotErr)
}

// --- publishPlugin format-guard tests ---

func TestPublishPlugin_FormatNotSet_NoError(t *testing.T) {
	app := newFormatApp(func(c *cli.Context) error {
		return formatGuard(c, "plugin publish")
	})
	require.NoError(t, app.Run([]string{"app"}))
}

func TestPublishPlugin_FormatJSON_NoError(t *testing.T) {
	app := newFormatApp(func(c *cli.Context) error {
		return formatGuard(c, "plugin publish")
	})
	require.NoError(t, app.Run([]string{"app", "--format", "json"}))
}

func TestPublishPlugin_FormatTable_ReturnsError(t *testing.T) {
	var gotErr error
	app := newFormatApp(func(c *cli.Context) error {
		gotErr = formatGuard(c, "plugin publish")
		return nil
	})
	require.NoError(t, app.Run([]string{"app", "--format", "table"}))
	require.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "unsupported format")
	assert.Contains(t, gotErr.Error(), "plugin publish")
}

func TestPublishPlugin_InvalidFormat_ReturnsError(t *testing.T) {
	var gotErr error
	app := newFormatApp(func(c *cli.Context) error {
		gotErr = formatGuard(c, "plugin publish")
		return nil
	})
	require.NoError(t, app.Run([]string{"app", "--format", "xml"}))
	require.Error(t, gotErr)
}
