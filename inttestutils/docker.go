package inttestutils

import (
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jfrog/jfrog-cli-core/v2/common/spec"

	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils/container"
	commonTests "github.com/jfrog/jfrog-cli-core/v2/common/tests"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/assert"
)

func BuildTestImage(imageName, dockerfileName, repo string, containerManagerType container.ContainerManagerType) (string, error) {
	log.Info("Building image", imageName, "with", containerManagerType.String())
	imageName = path.Join(*tests.ContainerRegistry, repo, imageName)
	dockerFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "docker")
	imageBuilder := commonTests.NewBuildDockerImage(imageName, dockerFilePath, containerManagerType).SetDockerFileName(dockerfileName)
	return imageName, gofrogcmd.RunCmd(imageBuilder)
}

func ContainerTestCleanup(t *testing.T, serverDetails *config.ServerDetails, artHttpDetails httputils.HttpClientDetails, imageName, buildName, repo string) {
	// Remove build from Artifactory
	DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
	tests.CleanFileSystem()
	// Remove image from Artifactory
	deleteSpec := spec.NewBuilder().Pattern(path.Join(repo, imageName)).BuildSpec()
	successCount, failCount, err := tests.DeleteFiles(deleteSpec, serverDetails)
	assert.Greater(t, successCount, 0)
	assert.Equal(t, failCount, 0)
	assert.NoError(t, err)
}

func getAllImagesNames(serverDetails *config.ServerDetails) ([]string, error) {
	var imageNames []string
	for _, repo := range []string{tests.DockerLocalRepo, tests.DockerLocalPromoteRepo} {
		prefix := repo + "/"
		specFile := spec.NewBuilder().Pattern(prefix + tests.DockerImageName + "*").IncludeDirs(true).BuildSpec()
		searchCmd := generic.NewSearchCommand()
		searchCmd.SetServerDetails(serverDetails).SetSpec(specFile)
		reader, err := searchCmd.Search()
		if err != nil {
			return nil, err
		}
		for searchResult := new(utils.SearchResult); reader.NextRecord(searchResult) == nil; searchResult = new(utils.SearchResult) {
			imageNames = append(imageNames, strings.TrimPrefix(searchResult.Path, prefix))
		}
		err = reader.Close()
		if err != nil {
			return nil, err
		}
	}
	return imageNames, nil
}

func CleanUpOldImages(serverDetails *config.ServerDetails) {
	getActualItems := func() ([]string, error) { return getAllImagesNames(serverDetails) }
	deleteItem := func(imageName string) {
		deleteSpec := spec.NewBuilder().Pattern(path.Join(tests.DockerLocalRepo, imageName)).BuildSpec()
		_, _, err := tests.DeleteFiles(deleteSpec, serverDetails)
		if err != nil {
			log.Error("Couldn't delete image", imageName, ":", imageName)
		} else {
			log.Info("Image", imageName, "deleted.")
		}
	}
	tests.CleanUpOldItems([]string{tests.DockerImageName}, getActualItems, deleteItem)
}
