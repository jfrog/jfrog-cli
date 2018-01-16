package utils

import (
	"encoding/json"
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"strings"
)

func ReadBintrayMessage(resp []byte) string {
	var response bintrayResponse
	err := json.Unmarshal(resp, &response)
	if err != nil {
		return string(resp)
	}
	return response.Message
}

func CreatePathDetails(str string) (*PathDetails, error) {
	parts := strings.Split(str, "/")
	size := len(parts)
	if size < 3 {
		err := errorutils.CheckError(errors.New("Expecting an argument in the form of subject/repository/file-path"))
		if err != nil {
			return nil, err
		}
	}
	path := strings.Join(parts[2:], "/")

	return &PathDetails{
		Subject: parts[0],
		Repo:    parts[1],
		Path:    path}, nil
}

type bintrayResponse struct {
	Message string
}

type FileDetails struct {
	Sha1 string
	Size int64
}

type PathDetails struct {
	Subject string
	Repo    string
	Path    string
}
