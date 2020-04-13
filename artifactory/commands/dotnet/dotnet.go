package dotnet

import (
	"github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli/utils/config"
)

type DotnetCommand struct {
	rtDetails          *config.ArtifactoryDetails
	args               []string
	repository         string
	buildConfiguration *utils.BuildConfiguration
	//shouldCollectBuildInfo	bool
}

func NewDotnetCommand() *DotnetCommand {
	return &DotnetCommand{}
}

func (dc *DotnetCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *DotnetCommand {
	dc.rtDetails = rtDetails
	return dc
}

func (dc *DotnetCommand) SetArgs(args []string) *DotnetCommand {
	dc.args = args
	return dc
}

func (dc *DotnetCommand) SetRepo(repo string) *DotnetCommand {
	dc.repository = repo
	return dc
}

func (dc *DotnetCommand) SetBuildConfiguration(buildConfiguration *utils.BuildConfiguration) *DotnetCommand {
	dc.buildConfiguration = buildConfiguration
	return dc
}

func (dc *DotnetCommand) CommandName() string {
	return "rt_dotnet"
}

func (dc *DotnetCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return dc.rtDetails, nil
}

func (dc *DotnetCommand) Run() error {
	return nil
}
