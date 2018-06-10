package versions

import (
	"errors"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"strings"
)

func CreatePath(versionStr string) (*Path, error) {
	parts := strings.Split(versionStr, "/")
	size := len(parts)
	if size < 1 || size > 4 {
		err := errorutils.CheckError(errors.New("Unexpected format for argument: " + versionStr))
		if err != nil {
			return nil, err
		}
	}
	var subject, repo, pkg, version string
	if size >= 2 {
		subject = parts[0]
		repo = parts[1]
	}
	if size >= 3 {
		pkg = parts[2]
	}
	if size == 4 {
		version = parts[3]
	}
	return &Path{
		Subject: subject,
		Repo:    repo,
		Package: pkg,
		Version: version}, nil
}
