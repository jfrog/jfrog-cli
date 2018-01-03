package fileutils

import (
	"archive/zip"
	"os"
	"path/filepath"
	"io"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"strings"
)

func IsZip(filename string) bool {
	return strings.HasSuffix(strings.ToLower(filename), ".zip")
}

func Unzip(at io.ReaderAt, size int64, dest string) error {
	r, err := zip.NewReader(at, size)
	if err != nil {
		return err
	}

	os.MkdirAll(dest, 0755)
	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func ZipFolderFiles(source, target string) (err error) {
	zipFile, err := os.Create(target)
	if err != nil {
		errorutils.CheckError(err)
		return
	}
	defer func() {
		if cerr := zipFile.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	archive := zip.NewWriter(zipFile)
	defer func() {
		if cerr := archive.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	filepath.Walk(source, func(path string, info os.FileInfo, err error) (currentErr error) {
		if info.IsDir() {
			return
		}

		if err != nil {
			currentErr = err
			return
		}

		header, currentErr := zip.FileInfoHeader(info)
		if currentErr != nil {
			errorutils.CheckError(currentErr)
			return
		}

		header.Method = zip.Deflate
		writer, currentErr := archive.CreateHeader(header)
		if currentErr != nil {
			errorutils.CheckError(currentErr)
			return
		}

		file, currentErr := os.Open(path)
		if currentErr != nil {
			errorutils.CheckError(currentErr)
			return
		}
		defer func() {
			if cerr := file.Close(); cerr != nil && currentErr == nil {
				currentErr = cerr
			}
		}()
		_, currentErr = io.Copy(writer, file)
		return
	})
	return
}
