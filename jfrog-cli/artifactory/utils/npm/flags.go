package npm

import "github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"

type CliFlags struct {
	BuildName   string
	BuildNumber string
	NpmArgs     string
	ArtDetails  *config.ArtifactoryDetails
}