package tests

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func CreateFileWithContent(fileName, relativePath string) (string, string, error) {
	var err error
	tempDirPath, err := ioutil.TempDir("", "tests")
	if err != nil {
		return tempDirPath, "", err
	}

	fullPath := ""
	if relativePath != "" {
		fullPath = filepath.Join(tempDirPath, relativePath)
		err = os.MkdirAll(fullPath, 0777)
		if err != nil {
			return tempDirPath, "", err
		}
	}
	fullPath = filepath.Join(fullPath, fileName)
	file, err := os.Create(fullPath)
	if err != nil {
		return tempDirPath, "", err
	}
	defer file.Close()
	_, err = file.Write([]byte(fullPath))
	return tempDirPath, fullPath, err
}
