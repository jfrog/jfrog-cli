package utils

import (
	"path/filepath"
	"strings"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	btutils "github.com/jfrogdev/jfrog-cli-go/jfrog-cli/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/bintray/commands"
)

// Download file from Bintray.
// downloadPath: download path in the following format: subject/repo/path/version/filename.
// filename: the file full name, will replace ${filename} in downloadPath argument.
// version: version of the desired file, will replace ${version} in downloadPath argument.
// targetPath: local download target path.
func DownloadFromBintray(downloadPath, filename, version, targetPath string) error {
	filename = strings.Replace(filename, "${version}", version, -1)
	targetFile := filepath.Join(targetPath, filename)
	if fileutils.IsPathExists(targetFile) {
		return nil
	}

	bintrayConfig := &config.BintrayDetails{ApiUrl: btutils.BINTRAY_API_URL, DownloadServerUrl: btutils.BINTRAY_DOWNLOAD_SERVER_URL}
	downloadFlags := &btutils.DownloadFlags{
		BintrayDetails:     bintrayConfig,
		Threads:            3,
		MinSplitSize:       5120,
		SplitCount:         3,
		IncludeUnpublished: false,
		Flat:               true}

	downloadPath = strings.Replace(downloadPath, "${version}", version, -1)
	downloadPath = strings.Replace(downloadPath, "${filename}", filename, -1)
	pathDetails, err := btutils.CreatePathDetails(downloadPath)
	if err != nil {
		return err
	}
	return commands.DownloadFile(pathDetails, targetFile, downloadFlags)
}
