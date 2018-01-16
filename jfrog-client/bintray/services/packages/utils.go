package packages

import (
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"strings"
)

func CreatePath(packageStr string) (*Path, error) {
	parts := strings.Split(packageStr, "/")
	size := len(parts)
	if size != 3 {
		err := errorutils.CheckError(errors.New("Expecting an argument in the form of subject/repository/package"))
		if err != nil {
			return nil, err
		}
	}
	return &Path{
		Subject: parts[0],
		Repo:    parts[1],
		Package: parts[2]}, nil
}
