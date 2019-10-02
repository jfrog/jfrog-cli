package pip

import (
	"fmt"
	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/auth"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"net/url"
	"strings"
)

type PipInstaller struct {
	RtDetails           *config.ArtifactoryDetails
	Args                []string
	Repository          string
	ShouldParseLogs     bool
	DependencyToFileMap map[string]string
}

func (pi *PipInstaller) Install() error {
	// Prepare for running.
	pipExecutablePath, pipIndexUrl, err := pi.prepare()
	if err != nil {
		return err
	}

	// Run pip install.
	err = pi.runPipInstall(pipExecutablePath, pipIndexUrl)
	if err != nil {
		return err
	}

	return nil
}

func (pi *PipInstaller) prepare() (pipExecutablePath, pipIndexUrl string, err error) {
	log.Debug("Preparing prerequisites.")

	pipExecutablePath, err = GetExecutablePath("pip")
	if err != nil {
		return
	}

	pipIndexUrl, err = getArtifactoryUrlWithCredentials(pi.RtDetails, pi.Repository)
	if err != nil {
		return
	}

	return
}

func getArtifactoryUrlWithCredentials(rtDetails *config.ArtifactoryDetails, repository string) (string, error) {
	rtUrl, err := url.Parse(rtDetails.GetUrl())
	if err != nil {
		return "", errorutils.CheckError(err)
	}

	username := rtDetails.GetUser()
	password := rtDetails.GetPassword()

	// Get credentials from access-token if exists.
	if rtDetails.GetAccessToken() != "" {
		username, err = auth.ExtractUsernameFromAccessToken(rtDetails.GetAccessToken())
		if err != nil {
			return "", err
		}
		password = rtDetails.GetAccessToken()
	}

	if username != "" && password != "" {
		rtUrl.User = url.UserPassword(username, password)
	}
	rtUrl.Path += "api/pypi/" + repository + "/simple"

	return rtUrl.String(), nil
}

func (pi *PipInstaller) runPipInstall(pipExecutablePath, pipIndexUrl string) error {
	pipInstallCmd := &PipCmd{
		Executable:  pipExecutablePath,
		Command:     "install",
		CommandArgs: append(pi.Args, "-i", pipIndexUrl),
	}

	// Check if need to run with log parsing.
	if pi.ShouldParseLogs {
		return pi.runPipInstallWithLogParsing(pipInstallCmd)
	}

	// Run without log parsing.
	return gofrogcmd.RunCmd(pipInstallCmd)
}

// Run pip-install command while parsing the logs for downloaded packages.
// Supports running pip either in non-verbose and verbose mode.
// Populates 'dependencyToFileMap' with downloaded package-name and its actual downloaded file (wheel/egg/zip...).
func (pi *PipInstaller) runPipInstallWithLogParsing(pipInstallCmd *PipCmd) error {
	// Create regular expressions for log parsing.
	collectingPackageRegexp, err := clientutils.GetRegExp(`^Collecting\s(\w[\w-\.]+)`)
	if err != nil {
		return err
	}
	downloadFileRegexp, err := clientutils.GetRegExp(`^\s\sDownloading\s[^\s]*\/packages\/[^\s]*\/([^\s]*)`)
	if err != nil {
		return err
	}
	installedPackagesRegexp, err := clientutils.GetRegExp(`^Requirement\salready\ssatisfied\:\s(\w[\w-\.]+)`)
	if err != nil {
		return err
	}

	downloadedDependencies := make(map[string]string)
	var packageName string
	expectingPackageFilePath := false

	// Extract downloaded package name.
	dependencyNameParser := gofrogcmd.CmdOutputPattern{
		RegExp: collectingPackageRegexp,
		ExecFunc: func(pattern *gofrogcmd.CmdOutputPattern) (string, error) {
			// If this pattern matched a second time before downloaded-file-name was found, prompt a message.
			if expectingPackageFilePath {
				// This may occur when a package-installation file is saved in pip-cache-dir, thus not being downloaded during the installation.
				// Re-running pip-install with 'no-cache-dir' fixes this issue.
				log.Debug(fmt.Sprintf("Could not resolve download path for package: %s, continuing...", packageName))

				// Save package with empty file path.
				downloadedDependencies[strings.ToLower(packageName)] = ""
			}

			// Check for out of bound results.
			if len(pattern.MatchedResults)-1 < 0 {
				log.Debug(fmt.Sprintf("Failed extracting package name from line: %s", pattern.Line))
				return pattern.Line, nil
			}

			// Save dependency information.
			expectingPackageFilePath = true
			packageName = pattern.MatchedResults[1]

			return pattern.Line, nil
		},
	}

	// Extract downloaded file, stored in Artifactory.
	dependencyFileParser := gofrogcmd.CmdOutputPattern{
		RegExp: downloadFileRegexp,
		ExecFunc: func(pattern *gofrogcmd.CmdOutputPattern) (string, error) {
			// Check for out of bound results.
			if len(pattern.MatchedResults)-1 < 0 {
				log.Debug(fmt.Sprintf("Failed extracting download path from line: %s", pattern.Line))
				return pattern.Line, nil
			}

			// If this pattern matched before package-name was found, do not collect this path.
			if !expectingPackageFilePath {
				log.Debug(fmt.Sprintf("Could not resolve package name for download path: %s , continuing...", packageName))
				return pattern.Line, nil
			}

			// Save dependency information.
			filePath := pattern.MatchedResults[1]
			downloadedDependencies[strings.ToLower(packageName)] = filePath
			expectingPackageFilePath = false

			log.Debug(fmt.Sprintf("Found package: %s installed with: %s", packageName, filePath))
			return pattern.Line, nil
		},
	}

	// Extract already installed packages names.
	installedPackagesParser := gofrogcmd.CmdOutputPattern{
		RegExp: installedPackagesRegexp,
		ExecFunc: func(pattern *gofrogcmd.CmdOutputPattern) (string, error) {
			// Check for out of bound results.
			if len(pattern.MatchedResults)-1 < 0 {
				log.Debug(fmt.Sprintf("Failed extracting package name from line: %s", pattern.Line))
				return pattern.Line, nil
			}

			// Save dependency with empty file name.
			downloadedDependencies[strings.ToLower(pattern.MatchedResults[1])] = ""

			log.Debug(fmt.Sprintf("Found package: %s already installed", pattern.MatchedResults[1]))
			return pattern.Line, nil
		},
	}

	// Execute command.
	_, _, _, err = gofrogcmd.RunCmdWithOutputParser(pipInstallCmd, true, &dependencyNameParser, &dependencyFileParser, &installedPackagesParser)
	if errorutils.CheckError(err) != nil {
		return err
	}

	// Update dependencyToFileMap.
	pi.DependencyToFileMap = downloadedDependencies

	return nil
}
