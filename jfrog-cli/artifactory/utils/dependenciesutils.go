package utils

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/bintray/commands"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray/auth"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray/services"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray/services/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"path/filepath"
	"strings"
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

	bintrayConfig := auth.NewBintrayDetails()
	config := bintray.NewConfigBuilder().
		SetBintrayDetails(bintrayConfig).
		Build()

	downloadPath = strings.Replace(downloadPath, "${version}", version, -1)
	downloadPath = strings.Replace(downloadPath, "${filename}", filename, -1)
	pathDetails, err := utils.CreatePathDetails(downloadPath)
	if err != nil {
		return err
	}

	params := &services.DownloadFileParams{}
	params.PathDetails = pathDetails
	params.TargetPath = targetFile
	params.Flat = true

	_, _, err = commands.DownloadFile(config, params)
	return err
}
