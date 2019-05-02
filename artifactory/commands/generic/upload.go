package generic

import (
	"errors"
	"github.com/jfrog/jfrog-cli-go/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	logUtils "github.com/jfrog/jfrog-cli-go/utils/log"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"os"
	"strconv"
	"strings"
)

// Uploads the artifacts in the specified local path pattern to the specified target path.
// Returns the total number of artifacts successfully uploaded.
func Upload(uploadSpec *spec.SpecFiles, configuration *utils.UploadConfiguration) (successCount, failCount int, logFile *os.File, err error) {
	// Initialize Progress bar, set logger to a log file
	progressBar, logFile, err := logUtils.InitProgressBarIfPossible()
	if err != nil {
		return 0, 0, logFile, err
	}
	if progressBar != nil {
		defer progressBar.Quit()
	}

	// Create Service Manager:
	certPath, err := utils.GetJfrogSecurityDir()
	if err != nil {
		return 0, 0, logFile, err
	}
	configuration.MinChecksumDeploySize, err = getMinChecksumDeploySize()
	if err != nil {
		return 0, 0, logFile, err
	}
	servicesManager, err := utils.CreateUploadServiceManager(configuration.ArtDetails, configuration, certPath, progressBar)
	if err != nil {
		return 0, 0, logFile, err
	}

	// Build Info Collection:
	isCollectBuildInfo := len(configuration.BuildName) > 0 && len(configuration.BuildNumber) > 0
	if isCollectBuildInfo && !configuration.DryRun {
		if err := utils.SaveBuildGeneralDetails(configuration.BuildName, configuration.BuildNumber); err != nil {
			return 0, 0, logFile, err
		}
		for i := 0; i < len(uploadSpec.Files); i++ {
			addBuildProps(&uploadSpec.Get(i).Props, configuration.BuildName, configuration.BuildNumber)
		}
	}

	var errorOccurred = false
	var uploadParamsArray []services.UploadParams
	// Create UploadParams for all File-Spec groups.
	for i := 0; i < len(uploadSpec.Files); i++ {
		uploadParams, err := getUploadParams(uploadSpec.Get(i), configuration)
		if err != nil {
			errorOccurred = true
			log.Error(err)
			continue
		}
		uploadParamsArray = append(uploadParamsArray, uploadParams)
	}

	// Perform upload.
	filesInfo, successCount, failCount, err := servicesManager.UploadFiles(uploadParamsArray...)
	if err != nil {
		errorOccurred = true
		log.Error(err)
	}

	if errorOccurred {
		err = errors.New("Upload finished with errors, Please review the logs.")
		return
	}
	if failCount > 0 {
		return
	}

	// Build Info
	if isCollectBuildInfo && !configuration.DryRun {
		buildArtifacts := convertFileInfoToBuildArtifacts(filesInfo)
		populateFunc := func(partial *buildinfo.Partial) {
			partial.Artifacts = buildArtifacts
		}
		err = utils.SavePartialBuildInfo(configuration.BuildName, configuration.BuildNumber, populateFunc)
	}
	return
}

func convertFileInfoToBuildArtifacts(filesInfo []clientutils.FileInfo) []buildinfo.Artifact {
	buildArtifacts := make([]buildinfo.Artifact, len(filesInfo))
	for i, fileInfo := range filesInfo {
		buildArtifacts[i] = fileInfo.ToBuildArtifacts()
	}
	return buildArtifacts
}

func getMinChecksumDeploySize() (int64, error) {
	minChecksumDeploySize := os.Getenv("JFROG_CLI_MIN_CHECKSUM_DEPLOY_SIZE_KB")
	if minChecksumDeploySize == "" {
		return 10240, nil
	}
	minSize, err := strconv.ParseInt(minChecksumDeploySize, 10, 64)
	err = errorutils.CheckError(err)
	if err != nil {
		return 0, err
	}
	return minSize * 1000, nil
}

func addBuildProps(props *string, buildName, buildNumber string) error {
	if buildName == "" || buildNumber == "" {
		return nil
	}
	buildProps, err := utils.CreateBuildProperties(buildName, buildNumber)
	if err != nil {
		return err
	}

	if len(*props) > 0 && !strings.HasSuffix(*props, ";") && len(buildProps) > 0 {
		*props += ";"
	}
	*props += buildProps
	return nil
}

func getUploadParams(f *spec.File, configuration *utils.UploadConfiguration) (uploadParams services.UploadParams, err error) {
	uploadParams = services.NewUploadParams()
	uploadParams.ArtifactoryCommonParams = f.ToArtifactoryCommonParams()
	uploadParams.Deb = configuration.Deb
	uploadParams.Symlink = configuration.Symlink
	uploadParams.MinChecksumDeploy = configuration.MinChecksumDeploySize

	uploadParams.Recursive, err = f.IsRecursive(true)
	if err != nil {
		return
	}

	uploadParams.Regexp, err = f.IsRegexp(false)
	if err != nil {
		return
	}

	uploadParams.IncludeDirs, err = f.IsIncludeDirs(false)
	if err != nil {
		return
	}

	uploadParams.Flat, err = f.IsFlat(true)
	if err != nil {
		return
	}

	uploadParams.ExplodeArchive, err = f.IsExplode(false)
	if err != nil {
		return
	}

	return
}
