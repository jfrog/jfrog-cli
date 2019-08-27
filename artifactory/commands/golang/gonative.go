package golang

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type GoNativeCommand struct {
	configFilePath string
	GoCommand
}

func NewGoNativeCommand() *GoNativeCommand {
	return &GoNativeCommand{GoCommand: *new(GoCommand)}
}

func (gnc *GoNativeCommand) SetConfigFilePath(configFilePath string) *GoNativeCommand {
	gnc.configFilePath = configFilePath
	return gnc
}

func (gnc *GoNativeCommand) SetArgs(args []string) *GoNativeCommand {
	gnc.goArg = args
	return gnc
}

func (gnc *GoNativeCommand) Run() error {
	// Read config file.
	log.Debug("Preparing to read the config file", gnc.configFilePath)
	vConfig, err := utils.ReadConfigFile(gnc.configFilePath, utils.YAML)
	if err != nil {
		return err
	}

	// Extract resolution params.
	gnc.resolverParams, err = utils.GetRepoConfigByPrefix(gnc.configFilePath, utils.ProjectConfigResolverPrefix, vConfig)
	if err != nil {
		return err
	}

	if vConfig.IsSet(utils.ProjectConfigDeployerPrefix) {
		// Extract deployer params.
		gnc.deployerParams, err = utils.GetRepoConfigByPrefix(gnc.configFilePath, utils.ProjectConfigDeployerPrefix, vConfig)
		if err != nil {
			return err
		}
		// Set to true for publishing dependencies.
		gnc.SetPublishDeps(true)
	}

	// Extract build info information from the args.
	flagIndex, valueIndex, buildName, err := utils.FindFlag("--build-name", gnc.goArg)
	if err != nil {
		return err
	}
	utils.RemoveFlagFromCommand(&gnc.goArg, flagIndex, valueIndex)

	flagIndex, valueIndex, buildNumber, err := utils.FindFlag("--build-number", gnc.goArg)
	if err != nil {
		return err
	}
	utils.RemoveFlagFromCommand(&gnc.goArg, flagIndex, valueIndex)

	flagIndex, valueIndex, module, err := utils.FindFlag("--module", gnc.goArg)
	if err != nil {
		return err
	}
	utils.RemoveFlagFromCommand(&gnc.goArg, flagIndex, valueIndex)

	gnc.buildConfiguration = &utils.BuildConfiguration{BuildName: buildName, BuildNumber: buildNumber, Module: module}
	return gnc.GoCommand.Run()
}

func (gnc *GoNativeCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	// If deployer Artifactory details exists, returs it.
	if gnc.deployerParams != nil && !gnc.deployerParams.IsRtDetailsEmpty() {
		return gnc.deployerParams.RtDetails()
	}

	// If resolver Artifactory details exists, returs it.
	if gnc.resolverParams != nil && !gnc.resolverParams.IsRtDetailsEmpty() {
		return gnc.resolverParams.RtDetails()
	}

	// If conf file exists, return the server configured in the conf file.
	if gnc.configFilePath != "" {
		vConfig, err := utils.ReadConfigFile(gnc.configFilePath, utils.YAML)
		if err != nil {
			return nil, err
		}
		return utils.GetRtDetails(vConfig)
	}
	return nil, nil
}
