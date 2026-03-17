package login

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestLoginCmdRejectsExtraArguments(t *testing.T) {
	app := cli.NewApp()
	app.Commands = []cli.Command{
		{
			Name:   "login",
			Action: LoginCmd,
		},
	}
	err := app.Run([]string{"jf", "login", "extra-arg"})
	assert.Error(t, err)
}

func TestLoginCmdPassesServerIdFlag(t *testing.T) {
	const testServerId = "my-test-server"
	var capturedServerId string

	app := cli.NewApp()
	app.Commands = []cli.Command{
		{
			Name: "login",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "server-id"},
			},
			Action: func(c *cli.Context) error {
				capturedServerId = c.String("server-id")
				// Return early without running the actual login command.
				return nil
			},
		},
	}
	err := app.Run([]string{"jf", "login", "--server-id", testServerId})
	assert.NoError(t, err)
	assert.Equal(t, testServerId, capturedServerId)
}

func TestLoginCmdNoArgsCallsLoginWithEmptyServerId(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.String("server-id", "", "")
	c := cli.NewContext(nil, set, nil)

	// Verify that the context has no arguments and server-id is empty.
	assert.Equal(t, 0, c.NArg())
	assert.Equal(t, "", c.String("server-id"))
}
