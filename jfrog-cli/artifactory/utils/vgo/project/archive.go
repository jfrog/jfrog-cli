package project

import (
	"archive/zip"
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/sabhiram/go-gitignore"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Archive project files according to the vgo project standard
func archiveProject(writer io.Writer, sourcePath, module, version string) error {
	// Ignoring unnecessary file
	ignoreParser, err := getIgnoreParser(sourcePath)
	if err != nil {
		return nil
	}
	zipWriter := zip.NewWriter(writer)
	defer zipWriter.Close()

	return filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return err
		}

		fileName := getFileName(sourcePath, path, module, version)
		if ignoreParser.MatchesPath(fileName) {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		zipFile, err := zipWriter.Create(fileName)
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

// Use .gitignore if available to filter out unnecessary files while archiving a vgo project.
// Ignores vendor and .gitignore file by default.
func getIgnoreParser(sourcePath string) (ignore.IgnoreParser, error) {
	exists, err := fileutils.IsFileExists(filepath.Join(sourcePath, ".gitignore"))
	if err != nil {
		return nil, err
	}

	if exists {
		return ignore.CompileIgnoreFileAndLines(filepath.Join(sourcePath, ".gitignore"), "vendor", ".gitignore")
	}
	return ignore.CompileIgnoreLines("vendor", ".gitignore")
}
