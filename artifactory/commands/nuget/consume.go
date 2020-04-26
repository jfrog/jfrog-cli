package nuget

import (
	"github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli/artifactory/utils/dotnet"
	dotnetCmd "github.com/jfrog/jfrog-cli/artifactory/utils/dotnet/commandargs"
	"github.com/jfrog/jfrog-cli/artifactory/utils/dotnet/solution"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"os"
)

type NugetCommand struct {
	configFilePath string
	*dotnetCmd.DotnetCommandArgs
}

func NewNugetCommand() *NugetCommand {
	nugetCmd := NugetCommand{"", &dotnetCmd.DotnetCommandArgs{}}
	nugetCmd.SetToolchainType(dotnet.Nuget)
	return &nugetCmd
}

func (nc *NugetCommand) SetConfigFilePath(configFilePath string) *NugetCommand {
	nc.configFilePath = configFilePath
	return nc
}

func (nc *NugetCommand) Run() error {
	// Read config file.
	log.Debug("Preparing to read the config file", nc.configFilePath)
	vConfig, err := utils.ReadConfigFile(nc.configFilePath, utils.YAML)
	if err != nil {
		return err
	}
	// Extract resolution params.
	resolveParams, err := utils.GetRepoConfigByPrefix(nc.configFilePath, utils.ProjectConfigResolverPrefix, vConfig)
	if err != nil {
		return err
	}
	RtDetails, err := resolveParams.RtDetails()
	if err != nil {
		return err
	}
	nc.SetRepoName(resolveParams.TargetRepo()).SetRtDetails(RtDetails)
	return nc.Exec()
}

func DependencyTreeCmd() error {
	workspace, err := os.Getwd()
	if err != nil {
		return errorutils.CheckError(err)
	}

	sol, err := solution.Load(workspace, "")
	if err != nil {
		return err
	}

	// Create the tree for each project
	for _, project := range sol.GetProjects() {
		err = project.CreateDependencyTree()
		if err != nil {
			return err
		}
	}
	// Build the tree.
	content, err := sol.Marshal()
	if err != nil {
		return errorutils.CheckError(err)
	}
	log.Output(clientutils.IndentJson(content))
	return nil
}
