package utils

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/jfrog/jfrog-client-go/utils/io"

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

func GetEncryptedPasswordFromArtifactory(artifactoryAuth auth.ServiceDetails, insecureTls bool) (string, error) {
	u, err := url.Parse(artifactoryAuth.GetUrl())
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, "api/security/encryptedPassword")
	httpClientsDetails := artifactoryAuth.CreateHttpClientDetails()
	certsPath, err := cliutils.GetJfrogCertsDir()
	if err != nil {
		return "", err
	}
	client, err := httpclient.ClientBuilder().
		SetCertificatesPath(certsPath).
		SetInsecureTls(insecureTls).
		SetClientCertPath(artifactoryAuth.GetClientCertPath()).
		SetClientCertKeyPath(artifactoryAuth.GetClientCertKeyPath()).
		Build()
	if err != nil {
		return "", err
	}
	resp, body, _, err := client.SendGet(u.String(), true, httpClientsDetails)
	if err != nil {
		return "", err
	}

	if resp.StatusCode == http.StatusOK {
		return string(body), nil
	}

	if resp.StatusCode == http.StatusConflict {
		message := "\nYour Artifactory server is not configured to encrypt passwords.\n" +
			"You may use \"art config --enc-password=false\""
		return "", errorutils.CheckError(errors.New(message))
	}

	return "", errorutils.CheckError(errors.New("Artifactory response: " + resp.Status))
}

func CreateServiceManager(artDetails *config.ArtifactoryDetails, isDryRun bool) (artifactory.ArtifactoryServicesManager, error) {
	return CreateServiceManagerWithThreads(artDetails, isDryRun, 0)
}

func CreateServiceManagerWithThreads(artDetails *config.ArtifactoryDetails, isDryRun bool, threads int) (artifactory.ArtifactoryServicesManager, error) {
	certsPath, err := cliutils.GetJfrogCertsDir()
	if err != nil {
		return nil, err
	}
	artAuth, err := artDetails.CreateArtAuthConfig()
	if err != nil {
		return nil, err
	}
	config := clientConfig.NewConfigBuilder().
		SetServiceDetails(artAuth).
		SetCertificatesPath(certsPath).
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

func CreateServiceManagerWithProgressBar(artDetails *config.ArtifactoryDetails, threads int, dryRun bool, progressBar io.Progress) (artifactory.ArtifactoryServicesManager, error) {
	certsPath, err := cliutils.GetJfrogCertsDir()
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
		SetCertificatesPath(certsPath).
		SetInsecureTls(artDetails.InsecureTls).
		SetThreads(threads).
		Build()

	if err != nil {
		return nil, err
	}
	return artifactory.NewWithProgress(&artAuth, servicesConfig, progressBar)
}

func CreateDistributionServiceManager(artDetails *config.ArtifactoryDetails, isDryRun bool) (*distribution.DistributionServicesManager, error) {
	certsPath, err := cliutils.GetJfrogCertsDir()
	if err != nil {
		return nil, err
	}
	distAuth, err := artDetails.CreateDistAuthConfig()
	if err != nil {
		return nil, err
	}
	serviceConfig, err := clientConfig.NewConfigBuilder().
		SetServiceDetails(distAuth).
		SetCertificatesPath(certsPath).
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
