package generic

import (
	commandsutils "github.com/jfrog/jfrog-cli-go/artifactory/commands/utils"
	"github.com/jfrog/jfrog-cli-go/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/utils/config"
)

type GenericCommand struct {
	rtDetails *config.ArtifactoryDetails
	spec      *spec.SpecFiles
	result    *commandsutils.Result
	dryRun    bool
}

func NewGenericCommand() *GenericCommand {
	return &GenericCommand{result: new(commandsutils.Result)}
}

func (gc *GenericCommand) DryRun() bool {
	return gc.dryRun
}

func (gc *GenericCommand) SetDryRun(dryRun bool) *GenericCommand {
	gc.dryRun = dryRun
	return gc
}

func (gc *GenericCommand) Result() *commandsutils.Result {
	return gc.result
}

func (gc *GenericCommand) Spec() *spec.SpecFiles {
	return gc.spec
}

func (gc *GenericCommand) SetSpec(spec *spec.SpecFiles) *GenericCommand {
	gc.spec = spec
	return gc
}

func (gc *GenericCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return gc.rtDetails, nil
}

func (gc *GenericCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *GenericCommand {
	gc.rtDetails = rtDetails
	return gc
}
