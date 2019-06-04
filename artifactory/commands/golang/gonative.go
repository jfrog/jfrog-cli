package golang

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/spf13/viper"
)

type GoNativeCommand struct {
	configFilePath string
	GoCommand
}

const (
	resolverPrefix = "resolver"
	deployerPrefix = "deployer"
)

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
	if !vConfig.IsSet(resolverPrefix) {
		return errorutils.CheckError(fmt.Errorf("Resolver information is missing"))
	}
	log.Debug("Found resolver in the config file")
	gnc.resolverParams, err = gnc.extractInfo(resolverPrefix, vConfig)
	if err != nil {
		return err
	}

	if vConfig.IsSet(deployerPrefix) {
		// Extract deployer params.
		log.Debug("Found deployer information in the config file")
		gnc.deployerParams, err = gnc.extractInfo(deployerPrefix, vConfig)
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

func (gnc *GoNativeCommand) extractInfo(prefix string, vConfig *viper.Viper) (*GoParamsCommand, error) {
	repo := vConfig.GetString(prefix + ".repo")
	if repo == "" {
		return nil, fmt.Errorf("Missing repository for %s within %s", prefix, gnc.configFilePath)
	}
	serverId := vConfig.GetString(prefix + ".serverID")
	if serverId == "" {
		return nil, fmt.Errorf("Missing server ID for %s within %s", prefix, gnc.configFilePath)
	}
	rtDetails, err := config.GetArtifactoryConf(serverId)
	if err != nil {
		return nil, err
	}
	return &GoParamsCommand{targetRepo: repo, rtDetails: rtDetails}, nil
}

func (gnc *GoNativeCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	// If deployer Artifactory details exists, returs it.
	if gnc.deployerParams != nil && !gnc.deployerParams.isRtDetailsEmpty() {
		return gnc.deployerParams.RtDetails()
	}

	// If resolver Artifactory details exists, returs it.
	if gnc.resolverParams != nil && !gnc.resolverParams.isRtDetailsEmpty() {
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
