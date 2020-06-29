package utils

import (
	"errors"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"os"
	"strings"

	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

const PathErrorSuffixMsg = " please enter a path, in which the new template file will be created"

func ValidateMapEntry(key string, value interface{}, writersMap map[string]AnswerWriter) error {
	if _, ok := writersMap[key]; !ok {
		return errorutils.CheckError(errors.New("template syntax error: unknown key: \"" + key + "\"."))
	}
	if _, ok := value.(string); !ok {
		return errorutils.CheckError(errors.New("template syntax error: the value for the  key: \"" + key + "\" is not a string type."))
	}
	return nil
}

func ValidateTemlatePath(templatePath string) error {
	exists, err := fileutils.IsDirExists(templatePath, false)
	if err != nil {
		return errorutils.CheckError(err)
	}
	if exists || strings.HasSuffix(templatePath, string(os.PathSeparator)) {
		return errorutils.CheckError(errors.New("path cannot be a directory," + PathErrorSuffixMsg))
	}
	exists, err = fileutils.IsFileExists(templatePath, false)
	if err != nil {
		return errorutils.CheckError(err)
	}
	if exists {
		return errorutils.CheckError(errors.New("file already exists," + PathErrorSuffixMsg))
	}
	return nil
}
