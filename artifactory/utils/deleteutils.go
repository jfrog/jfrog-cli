package utils

import (
	"regexp"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"strings"
	"errors"
)

func WildcardToDirsPath(deletePattern, searchResult string) (string, error) {
	if !strings.HasSuffix(deletePattern, "/") {
		return "", errors.New("Delete pattern must end with \"/\"")
	}

	regexpPattern := "^" + strings.Replace(deletePattern, "*", "([^/]*|.*)", -1)
	r, err := regexp.Compile(regexpPattern)
	cliutils.CheckError(err)
	if err != nil {
		return "", err
	}

	groups := r.FindStringSubmatch(searchResult)
	if len(groups) > 0 {
		return groups[0], nil
	}
	return "", nil
}