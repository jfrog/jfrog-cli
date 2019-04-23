package project

import (
	"archive/zip"
	"fmt"
	"github.com/jfrog/jfrog-cli-go/utils/ioutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Archive project files according to the go project standard
func archiveProject(writer io.Writer, sourcePath, module, version string, excludePathsRegExp *regexp.Regexp) error {
	zipWriter := zip.NewWriter(writer)
	defer zipWriter.Close()

	return filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() || fileutils.IsPathSymlink(path) {
			return err
		}

		if excludePathsRegExp.FindString(path) != "" {
			log.Debug(fmt.Sprintf("Excluding path '%s' from zip archive.", path))
			return nil
		}
		fileName := getFileName(sourcePath, path, module, version)
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		zipFile, err := zipWriter.Create(ioutils.PrepareFilePathForUnix(fileName))
		if err != nil {
			return err
		}

		_, err = io.CopyN(zipFile, file, info.Size())
		return err
	})
}

// getFileName composes filename for zip to match standard specified as
// module@version/{filename}
func getFileName(sourcePath, filePath, moduleName, version string) string {
	filename := strings.TrimPrefix(filePath, sourcePath)
	filename = strings.TrimLeft(filename, string(os.PathSeparator))
	moduleID := fmt.Sprintf("%s@%s", moduleName, version)

	return filepath.Join(moduleID, filename)
}
