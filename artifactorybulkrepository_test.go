package main

import (
	"fmt"
	"testing"

	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryCreateAndUpdateIntegration(t *testing.T) {
	initArtifactoryTest(t, "")
	defer cleanArtifactoryTest()

	t.Run("SingleRepositoryCreate", func(t *testing.T) {
		testSingleRepositoryCreate(t)
	})

	t.Run("SingleRepositoryUpdate", func(t *testing.T) {
		testSingleRepositoryUpdate(t)
	})

	t.Run("MultipleRepositoryCreate", func(t *testing.T) {
		testMultipleRepositoryCreate(t)
	})

	t.Run("MultipleRepositoryUpdate", func(t *testing.T) {
		testMultipleRepositoryUpdate(t)
	})

	t.Run("DifferentPackageTypesAndRClass", func(t *testing.T) {
		testCreateWithDifferentRclass(t)
	})

	t.Run("DifferentPackageTypesAndRClassUpdate", func(t *testing.T) {
		testUpdateWithDifferentRclass(t)
	})
}

func testSingleRepositoryCreate(t *testing.T) {
	templatePath := tests.GetTestResourcesPath() + tests.MavenRepositoryConfig1

	runRt(t, "repo-create", templatePath, "--vars", fmt.Sprintf("MAVEN_REPO1=%s", tests.MvnRepo1))

	assert.True(t, isRepoExist(tests.MvnRepo1), "Repository should exist after creation")

	execDeleteRepo(tests.MvnRepo1)
}

func testSingleRepositoryUpdate(t *testing.T) {
	templatePath := tests.GetTestResourcesPath() + tests.MavenRepositoryConfig1

	runRt(t, "repo-create", templatePath, "--vars", fmt.Sprintf("MAVEN_REPO1=%s", tests.MvnRepo1))

	runRt(t, "repo-update", templatePath, "--vars", fmt.Sprintf("MAVEN_REPO1=%s", tests.MvnRepo1))

	assert.True(t, isRepoExist(tests.MvnRepo1), "Repository should still exist after update")

	execDeleteRepo(tests.MvnRepo1)
}

func testMultipleRepositoryCreate(t *testing.T) {
	repo1Name := "test-maven-repo1"
	repo2Name := "test-npm-repo2"
	repo3Name := "test-docker-repo3"

	templatePath := tests.GetTestResourcesPath() + tests.MultipleRepositoriesConfig

	runRt(t, "repo-create", templatePath, "--vars",
		fmt.Sprintf("REPO1_KEY=%s;REPO2_KEY=%s;REPO3_KEY=%s", repo1Name, repo2Name, repo3Name))

	assert.True(t, isRepoExist(repo1Name), "Repository 1 doesn't exist")
	assert.True(t, isRepoExist(repo2Name), "Repository 2 doesn't exist")
	assert.True(t, isRepoExist(repo3Name), "Repository 3 doesn't exist")

	execDeleteRepo(repo1Name)
	execDeleteRepo(repo2Name)
	execDeleteRepo(repo3Name)
}

func testMultipleRepositoryUpdate(t *testing.T) {
	repo1Name := "test-multi-repo1"
	repo2Name := "test-multi-repo2"
	repo3Name := "test-multi-repo3"

	templatePath := tests.GetTestResourcesPath() + tests.MultipleRepositoriesConfig

	runRt(t, "repo-create", templatePath, "--vars",
		fmt.Sprintf("REPO1_KEY=%s;REPO2_KEY=%s;REPO3_KEY=%s", repo1Name, repo2Name, repo3Name))

	runRt(t, "repo-update", templatePath, "--vars",
		fmt.Sprintf("REPO1_KEY=%s;REPO2_KEY=%s;REPO3_KEY=%s", repo1Name, repo2Name, repo3Name))

	assert.True(t, isRepoExist(repo1Name), "Repository 1 doesn't exist")
	assert.True(t, isRepoExist(repo2Name), "Repository 2 doesn't exist")
	assert.True(t, isRepoExist(repo3Name), "Repository 3 doesn't exist")

	execDeleteRepo(repo1Name)
	execDeleteRepo(repo2Name)
	execDeleteRepo(repo3Name)
}

func testCreateWithDifferentRclass(t *testing.T) {
	t.Skip("RTECO-525 - Skipping testCreateWithDifferentRclass")

	mvnRepoName := "test-mvn-local1"
	dockerLocalRepoName := "test-docker-local1"
	npmRepoName := "test-npm-local"
	dockerRemoteRepoName := "test-docker-remote1"
	dockerVirtualRepoName := "test-docker-virtual1"

	templatePathForLocalMvnRepo := tests.GetTestResourcesPath() + tests.MavenRepositoryConfig1
	templatePathForDockerRepo := tests.GetTestResourcesPath() + tests.DockerLocalRepositoryConfig

	runRt(t, "repo-create", templatePathForLocalMvnRepo, "--vars", fmt.Sprintf("MAVEN_REPO1=%s", mvnRepoName))
	runRt(t, "repo-create", templatePathForDockerRepo, "--vars", fmt.Sprintf("DOCKER_REPO=%s", dockerLocalRepoName))

	templatePath := tests.GetTestResourcesPath() + tests.MixedRepositoriesConfig
	runRt(t, "repo-create", templatePath, "--vars",
		fmt.Sprintf("REPO2_KEY=%s;DOCKER_REMOTE_REPO=%s;DOCKER_VIRTUAL_REPO=%s;REPO1=%s;DEBIAN_REPO=%s",
			npmRepoName, dockerRemoteRepoName, dockerVirtualRepoName, mvnRepoName, dockerLocalRepoName))

	assert.True(t, isRepoExist(mvnRepoName), "local Maven repository doesn't exist")
	assert.True(t, isRepoExist(dockerLocalRepoName), "local docker repository doesn't exist")
	assert.True(t, isRepoExist(npmRepoName), "local npm repository doesn't exist")
	assert.True(t, isRepoExist(dockerRemoteRepoName), "remote docker repository doesn't exist")
	assert.True(t, isRepoExist(dockerVirtualRepoName), "virtual docker repository doesn't exist")

	execDeleteRepo(dockerVirtualRepoName)
	execDeleteRepo(dockerRemoteRepoName)
	execDeleteRepo(npmRepoName)
	execDeleteRepo(dockerLocalRepoName)
	execDeleteRepo(mvnRepoName)
}

func testUpdateWithDifferentRclass(t *testing.T) {
	t.Skip("RTECO-525 - Skipping testUpdateWithDifferentRclass")

	mvnRepoName := "test-mvn-local-1"
	dockerLocalRepoName := "test-docker-local-1"
	npmRepoName := "test-npm-local"
	dockerRemoteRepoName := "test-docker-remote-1"
	dockerVirtualRepoName := "test-docker-virtual-1"

	templatePathForLocalMvnRepo := tests.GetTestResourcesPath() + tests.MavenRepositoryConfig1
	templatePathForDockerRepo := tests.GetTestResourcesPath() + tests.DockerLocalRepositoryConfig

	runRt(t, "repo-create", templatePathForLocalMvnRepo, "--vars", fmt.Sprintf("MAVEN_REPO1=%s", mvnRepoName))
	runRt(t, "repo-create", templatePathForDockerRepo, "--vars", fmt.Sprintf("DOCKER_REPO=%s", dockerLocalRepoName))

	createTemplatePath := tests.GetTestResourcesPath() + tests.MixedRepositoriesConfig
	runRt(t, "repo-create", createTemplatePath, "--vars",
		fmt.Sprintf("REPO2_KEY=%s;DOCKER_REMOTE_REPO=%s;DOCKER_VIRTUAL_REPO=%s;REPO1=%s;DEBIAN_REPO=%s",
			npmRepoName, dockerRemoteRepoName, dockerVirtualRepoName, mvnRepoName, dockerLocalRepoName))

	updateTemplatePath := tests.GetTestResourcesPath() + tests.MixedRepositoriesUpdateConfig
	runRt(t, "repo-update", templatePathForLocalMvnRepo, "--vars", fmt.Sprintf("MAVEN_REPO1=%s", mvnRepoName))
	runRt(t, "repo-update", templatePathForDockerRepo, "--vars", fmt.Sprintf("DOCKER_REPO=%s", dockerLocalRepoName))
	runRt(t, "repo-update", updateTemplatePath, "--vars",
		fmt.Sprintf("REPO2_KEY=%s;DOCKER_REMOTE_REPO=%s;DOCKER_VIRTUAL_REPO=%s;REPO1=%s;DEBIAN_REPO=%s",
			npmRepoName, dockerRemoteRepoName, dockerVirtualRepoName, mvnRepoName, dockerLocalRepoName))

	assert.True(t, isRepoExist(mvnRepoName), "local Maven repository doesn't exist after update")
	assert.True(t, isRepoExist(dockerLocalRepoName), "local docker repository doesn't exist after update")
	assert.True(t, isRepoExist(npmRepoName), "local npm repository doesn't exist after update")
	assert.True(t, isRepoExist(dockerRemoteRepoName), "remote docker repository doesn't exist after update")
	assert.True(t, isRepoExist(dockerVirtualRepoName), "virtual docker repository doesn't exist after update")

	execDeleteRepo(dockerVirtualRepoName)
	execDeleteRepo(dockerRemoteRepoName)
	execDeleteRepo(npmRepoName)
	execDeleteRepo(dockerLocalRepoName)
	execDeleteRepo(mvnRepoName)
}
