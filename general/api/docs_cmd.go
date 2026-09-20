package api

import (
	"fmt"

	"github.com/urfave/cli"
)

// DocsCommand is the fallback Action for the bare "docs" node (jf api docs).
// It has no behavior of its own beyond dispatching to "search" — same pattern
// as jf hf's bare-subcommand fallback in buildtools/cli.go.
func DocsCommand(c *cli.Context) error {
	if c.Args().Present() {
		return fmt.Errorf("'%s %s' is not a valid subcommand. Run 'jf api docs --help' for usage", c.App.Name, c.Args().First())
	}
	return cli.ShowSubcommandHelp(c)
}
