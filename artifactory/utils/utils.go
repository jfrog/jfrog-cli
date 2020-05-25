package utils

import (
	"errors"
	"github.com/jfrog/jfrog-client-go/utils/io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/auth"
	clientConfig "github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/distribution"
	"github.com/jfrog/jfrog-client-go/httpclient"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

const repoDetailsUrl = "api/repositories/"

func GetProjectDir(global bool) (string, error) {
	configDir, err := getConfigDir(global)
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	return filepath.Join(configDir, "projects"), nil
}

func getConfigDir(global bool) (string, error) {
	if !global {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return filepath.Join(wd, ".jfrog"), nil
	}
	return cliutils.GetJfrogHomeDir()
}

func CreateServiceManager(artDetails *config.ArtifactoryDetails, isDryRun bool) (*artifactory.ArtifactoryServicesManager, error) {
	return CreateServiceManagerWithThreads(artDetails, isDryRun, 0)
}

func CreateServiceManagerWithThreads(artDetails *config.ArtifactoryDetails, isDryRun bool, threads int) (*artifactory.ArtifactoryServicesManager, error) {
	certPath, err := cliutils.GetJfrogSecurityDir()
	if err != nil {
		return nil, err
	}
	artAuth, err := artDetails.CreateArtAuthConfig()
	if err != nil {
		return nil, err
	}
	config := clientConfig.NewConfigBuilder().
		SetServiceDetails(artAuth).
		SetCertificatesPath(certPath).
		SetInsecureTls(artDetails.InsecureTls).
		SetDryRun(isDryRun)
	if threads > 0 {
		config.SetThreads(threads)
	}
	serviceConfig, err := config.Build()
	if err != nil {
		return nil, err
	}
	return artifactory.New(&artAuth, serviceConfig)
}

func CreateServiceManagerWithProgressBar(artDetails *config.ArtifactoryDetails, threads int, dryRun bool, progressBar io.Progress) (*artifactory.ArtifactoryServicesManager, error) {
	certPath, err := cliutils.GetJfrogSecurityDir()
	if err != nil {
		return nil, err
	}
	artAuth, err := artDetails.CreateArtAuthConfig()
	if err != nil {
		return nil, err
	}
	servicesConfig, err := clientConfig.NewConfigBuilder().
		SetServiceDetails(artAuth).
		SetDryRun(dryRun).
		SetCertificatesPath(certPath).
		SetInsecureTls(artDetails.InsecureTls).
		SetThreads(threads).
		Build()

	if err != nil {
		return nil, err
	}
	return artifactory.NewWithProgress(&artAuth, servicesConfig, progressBar)
}

func CreateDistributionServiceManager(artDetails *config.ArtifactoryDetails, isDryRun bool) (*distribution.DistributionServicesManager, error) {
	certPath, err := cliutils.GetJfrogSecurityDir()
	if err != nil {
		return nil, err
	}
	distAuth, err := artDetails.CreateDistAuthConfig()
	if err != nil {
		return nil, err
	}
	serviceConfig, err := clientConfig.NewConfigBuilder().
		SetServiceDetails(distAuth).
		SetCertificatesPath(certPath).
		SetInsecureTls(artDetails.InsecureTls).
		SetDryRun(isDryRun).
		Build()
	if err != nil {
		return nil, err
	}
	return distribution.New(&distAuth, serviceConfig)
}

func isRepoExists(repository string, artDetails auth.ServiceDetails) (bool, error) {
	artHttpDetails := artDetails.CreateHttpClientDetails()
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return false, err
	}
	resp, _, _, err := client.SendGet(artDetails.GetUrl()+repoDetailsUrl+repository, true, artHttpDetails)
	if err != nil {
		return false, errorutils.CheckError(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		return true, nil
	}
	return false, nil
}

func CheckIfRepoExists(repository string, artDetails auth.ServiceDetails) error {
	repoExists, err := isRepoExists(repository, artDetails)
	if err != nil {
		return err
	}

	if !repoExists {
		return errorutils.CheckError(errors.New("The repository '" + repository + "' does not exist."))
	}
	return nil
}

// Get build name and number from env, only if both missing
func GetBuildNameAndNumber(buildName, buildNumber string) (string, string) {
	if buildName != "" || buildNumber != "" {
		return buildName, buildNumber
	}
	return os.Getenv(cliutils.BuildName), os.Getenv(cliutils.BuildNumber)
}

func GetBuildName(buildName string) string {
	return getOrDefaultEnv(buildName, cliutils.BuildName)
}

func GetBuildUrl(buildUrl string) string {
	return getOrDefaultEnv(buildUrl, cliutils.BuildUrl)
}

func GetEnvExclude(envExclude string) string {
	return getOrDefaultEnv(envExclude, cliutils.EnvExclude)
}

// Return argument if not empty or retrieve from environment variable
func getOrDefaultEnv(arg, envKey string) string {
	if arg != "" {
		return arg
	}
	return os.Getenv(envKey)
}

// This error indicates that the build was scanned by Xray, but Xray found issues with the build.
// If Xray failed to scan the build, for example due to a networking issue, a regular error should be returned.
var buildScanError = errors.New("issues found during xray build scan")

func GetBuildScanError() error {
	return buildScanError
}
