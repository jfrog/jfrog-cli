package invite

import (
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/general/invite"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/urfave/cli"
)

func RunInviteCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	serverDetails, err := cliutils.CreateArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}
	inviteCmd := invite.NewInviteCommand().SetInvitedEmail(c.Args().Get(0)).SetServerDetails(serverDetails)
	return commands.Exec(inviteCmd)
}
