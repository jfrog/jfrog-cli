package pip

import (
	"github.com/jfrog/jfrog-cli-go/utils/config"
)

type PipCommand struct {
	rtDetails  *config.ArtifactoryDetails
	args       []string
	repository string
}

func (pc *PipCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *PipCommand {
	pc.rtDetails = rtDetails
	return pc
}

func (pc *PipCommand) SetRepo(repo string) *PipCommand {
	pc.repository = repo
	return pc
}

func (pc *PipCommand) SetArgs(arguments []string) *PipCommand {
	pc.args = arguments
	return pc
}

type PipCommandInterface interface {
	SetRtDetails(rtDetails *config.ArtifactoryDetails) *PipCommand
	SetRepo(repo string) *PipCommand
	SetArgs(arguments []string) *PipCommand
	RtDetails() (*config.ArtifactoryDetails, error)
	CommandName() string
	Run() error
}
