package api

import (
	"bytes"
	"encoding/json"
	"testing"

	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

// newSearchApp builds a minimal cli.App exercising runSearchCmd exactly like
// the real "search" subcommand's flag set, without going through main.go's
// full command tree.
func newSearchApp(stdOut *bytes.Buffer, capturedErr *error) *cli.App {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: flagTag},
		cli.StringFlag{Name: flagMethod},
		cli.IntFlag{Name: flagLimit, Value: defaultLimit},
		cli.StringFlag{Name: "format"},
	}
	app.Action = func(c *cli.Context) error {
		*capturedErr = runSearchCmd(c, stdOut)
		return nil
	}
	return app
}

// newDescribeApp builds a minimal cli.App exercising runDescribeCmd exactly
// like the real "describe" subcommand's flag set.
func newDescribeApp(stdOut *bytes.Buffer, capturedErr *error) *cli.App {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "format"},
	}
	app.Action = func(c *cli.Context) error {
		*capturedErr = runDescribeCmd(c, stdOut)
		return nil
	}
	return app
}

// runSearchJSON runs the search app with JSON output (the default) and returns
// the parsed result body plus whatever landed on the logger's Warn/Info/Error
// channel.
func runSearchJSON(t *testing.T, args ...string) (result map[string]any, logged string) {
	t.Helper()
	var jsonOut, logOut bytes.Buffer
	logger := clientlog.NewLoggerWithFlags(clientlog.INFO, &logOut, 0)
	logger.SetOutputWriter(&jsonOut)
	prevLogger := clientlog.GetLogger()
	t.Cleanup(func() { clientlog.SetLogger(prevLogger) })
	clientlog.SetLogger(logger)

	var stdOut bytes.Buffer
	var runErr error
	app := newSearchApp(&stdOut, &runErr)

	require.NoError(t, app.Run(append([]string{"cmd"}, args...)))
	require.NoError(t, runErr)

	require.NoError(t, json.Unmarshal(jsonOut.Bytes(), &result), "output should be parseable JSON")
	return result, logOut.String()
}

// runDescribeJSON runs the describe app with JSON output (the default) and
// returns the parsed result body.
func runDescribeJSON(t *testing.T, method, path string) map[string]any {
	t.Helper()
	var jsonOut, logOut bytes.Buffer
	logger := clientlog.NewLoggerWithFlags(clientlog.INFO, &logOut, 0)
	logger.SetOutputWriter(&jsonOut)
	prevLogger := clientlog.GetLogger()
	t.Cleanup(func() { clientlog.SetLogger(prevLogger) })
	clientlog.SetLogger(logger)

	var stdOut bytes.Buffer
	var runErr error
	app := newDescribeApp(&stdOut, &runErr)

	require.NoError(t, app.Run([]string{"cmd", method, path}))
	require.NoError(t, runErr)

	var result map[string]any
	require.NoError(t, json.Unmarshal(jsonOut.Bytes(), &result), "output should be parseable JSON")
	return result
}
