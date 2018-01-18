package commands

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils/buildinfo"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils/npm"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services"
	specutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"github.com/mattn/go-shellwords"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func NpmPublish(repo string, cliFlags *npm.CliFlags) (err error) {
	log.Info("Running npm Publish")
	npmp := npmPublish{cliFlags: cliFlags, repo: repo}
	if err = npmp.preparePrerequisites(); err != nil {
		return err
	}

	if !npmp.tarballProvided {
		if err = npmp.pack(); err != nil {
			return err
		}
	}

	if err = npmp.deploy(); err != nil {
		if npmp.tarballProvided {
			return err
		}
		// We should delete the tarball we created
		return deleteCreatedTarballAndError(npmp.packedFilePath, err)
	}

	if !npmp.tarballProvided {
		if err = deleteCreatedTarball(npmp.packedFilePath); err != nil {
			return err
		}
	}

	if !npmp.collectBuildInfo {
		log.Info("npm publish finished successfully.")
		return nil
	}

	if err = npmp.saveArtifactData(); err != nil {
		return err
	}

	log.Info("npm publish finished successfully.")
	return nil
}

func (npmp *npmPublish) preparePrerequisites() error {
	log.Debug("Preparing prerequisites.")
	npmExecPath, err := exec.LookPath("npm")
	if err != nil {
		return errorutils.CheckError(err)
	}

	if npmExecPath == "" {
		return errorutils.CheckError(errors.New("Could not find 'npm' executable"))
	}

	npmp.executablePath = npmExecPath
	log.Debug("Using npm executable:", npmp.executablePath)
	currentDir, err := os.Getwd()
	if err != nil {
		return errorutils.CheckError(err)
	}

	currentDir, err = filepath.Abs(currentDir)
	if err != nil {
		return errorutils.CheckError(err)
	}

	npmp.workingDirectory = currentDir
	log.Debug("Working directory set to:", npmp.workingDirectory)
	npmp.collectBuildInfo = len(npmp.cliFlags.BuildName) > 0 && len(npmp.cliFlags.BuildNumber) > 0
	if err = npmp.setPublishPath(); err != nil {
		return err
	}

	artDetails, err := npmp.cliFlags.ArtDetails.CreateArtAuthConfig()
	if err != nil {
		return err
	}

	if err = utils.CheckIfRepoExists(npmp.repo, artDetails); err != nil {
		return err
	}

	return npmp.setPackageInfo()
}

func (npmp *npmPublish) pack() error {
	log.Debug("Creating npm package.")
	if err := npm.Pack(npmp.cliFlags.NpmArgs, npmp.executablePath); err != nil {
		return err
	}

	npmp.packedFilePath = filepath.Join(npmp.workingDirectory, npmp.packageInfo.GetExpectedPackedFileName())
	log.Debug("Created npm package at", npmp.packedFilePath)
	return nil
}

func (npmp *npmPublish) deploy() (err error) {
	log.Debug("Deploying npm package.")
	if err = npmp.readPackageInfoFromTarball(); err != nil {
		return err
	}

	target := fmt.Sprintf("%s/%s", npmp.repo, npmp.packageInfo.GetDeployPath())
	artifactsFileInfo, err := npmp.doDeploy(target, npmp.cliFlags.ArtDetails)
	if err != nil {
		return err
	}

	npmp.artifactData = artifactsFileInfo
	return nil
}

func (npmp *npmPublish) doDeploy(target string, artDetails *config.ArtifactoryDetails) (artifactsFileInfo []specutils.FileInfo, err error) {
	servicesManager, err := utils.CreateServiceManager(artDetails, false)
	if err != nil {
		return nil, err
	}
	up := &services.UploadParamsImp{}
	up.ArtifactoryCommonParams = &specutils.ArtifactoryCommonParams{Pattern: npmp.packedFilePath, Target: target}
	if npmp.collectBuildInfo {
		utils.SaveBuildGeneralDetails(npmp.cliFlags.BuildName, npmp.cliFlags.BuildNumber)
		props, err := utils.CreateBuildProperties(npmp.cliFlags.BuildName, npmp.cliFlags.BuildNumber)
		if err != nil {
			return nil, err
		}
		up.ArtifactoryCommonParams.Props = props
	}
	artifactsFileInfo, _, failed, err := servicesManager.UploadFiles(up)
	if err != nil {
		return nil, err
	}

	// We deploying only one Artifact which have to be deployed, in case of failure we should fail
	if failed > 0 {
		return nil, errorutils.CheckError(errors.New("Failed to upload the npm package to Artifactory. See Artifactory logs for more details."))
	}
	return artifactsFileInfo, nil
}

func (npmp *npmPublish) saveArtifactData() error {
	log.Debug("Saving npm package artifact build info data.")
	buildArtifacts := convertFileInfoToBuildArtifacts(npmp.artifactData)
	populateFunc := func(partial *buildinfo.Partial) {
		partial.Artifacts = buildArtifacts
		partial.ModuleId = npmp.packageInfo.BuildInfoModuleId()
	}
	return utils.SavePartialBuildInfo(npmp.cliFlags.BuildName, npmp.cliFlags.BuildNumber, populateFunc)
}

func (npmp *npmPublish) setPublishPath() error {
	log.Debug("Reading Package Json.")
	splitFlags, err := shellwords.Parse(npmp.cliFlags.NpmArgs)
	if err != nil {
		return errorutils.CheckError(err)
	}

	npmp.publishPath = npmp.workingDirectory
	if len(splitFlags) > 0 && !strings.HasPrefix(strings.TrimSpace(splitFlags[0]), "-") {
		path := strings.TrimSpace(splitFlags[0])
		path = clientutils.ReplaceTildeWithUserHome(path)
		if err != nil {
			return errorutils.CheckError(err)
		}

		if filepath.IsAbs(path) {
			npmp.publishPath = path
		} else {
			npmp.publishPath = filepath.Join(npmp.workingDirectory, path)
		}
	}
	return nil
}

func (npmp *npmPublish) setPackageInfo() error {
	log.Debug("Setting Package Info.")
	fileInfo, err := os.Stat(npmp.publishPath)
	if err != nil {
		return errorutils.CheckError(err)
	}

	if fileInfo.IsDir() {
		npmp.packageInfo, err = npm.ReadPackageInfoFromPackageJson(npmp.publishPath)
		return err
	}
	log.Debug("The provided path is not a directory, we assume this is a compressed npm package")
	npmp.tarballProvided = true
	npmp.packedFilePath = npmp.publishPath
	return npmp.readPackageInfoFromTarball()
}

func (npmp *npmPublish) readPackageInfoFromTarball() error {
	log.Debug("Extracting info from npm package:", npmp.packedFilePath)
	tarball, err := os.Open(npmp.packedFilePath)
	if err != nil {
		return errorutils.CheckError(err)
	}
	defer tarball.Close()
	gZipReader, err := gzip.NewReader(tarball)
	if err != nil {
		return errorutils.CheckError(err)
	}

	tarReader := tar.NewReader(gZipReader)
	for {
		hdr, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				return errorutils.CheckError(errors.New("Could not find 'package.json' in the compressed npm package: " + npmp.packedFilePath))
			}
			return errorutils.CheckError(err)
		}
		if strings.HasSuffix(hdr.Name, "package.json") {
			packageJson, err := ioutil.ReadAll(tarReader)
			if err != nil {
				return errorutils.CheckError(err)
			}

			npmp.packageInfo, err = npm.ReadPackageInfo(packageJson)
			return err
		}
	}
}

func deleteCreatedTarballAndError(packedFilePath string, currentError error) error {
	if err := deleteCreatedTarball(packedFilePath); err != nil {
		errorText := fmt.Sprintf("Two errors occurred: \n%s \n%s", currentError, err)
		return errorutils.CheckError(errors.New(errorText))
	}
	return currentError
}

func deleteCreatedTarball(packedFilePath string) error {
	if err := os.Remove(packedFilePath); err != nil {
		return errorutils.CheckError(err)
	}
	log.Debug("Successfully deleted the created npm package:", packedFilePath)
	return nil
}

type npmPublish struct {
	executablePath   string
	cliFlags         *npm.CliFlags
	workingDirectory string
	collectBuildInfo bool
	packedFilePath   string
	packageInfo      *npm.PackageInfo
	publishPath      string
	tarballProvided  bool
	artifactData     []specutils.FileInfo
	repo             string
}
