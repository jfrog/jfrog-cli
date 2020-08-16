package utils

import (
	"encoding/json"
	"errors"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"os"
	"strings"

	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

const PathErrorSuffixMsg = " please enter a path, in which the new template file will be created"

type TemplateUserCommand interface {
	// Returns the file path.
	TemplatePath() string
	// Returns vars to replace in the template content.
	Vars() string
}

func ConvertTemplateToMap(tuc TemplateUserCommand) (map[string]interface{}, error) {
	// Read the template file
	content, err := fileutils.ReadFile(tuc.TemplatePath())
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	// Replace vars string-by-string if needed
	if len(tuc.Vars()) > 0 {
		templateVars := cliutils.SpecVarsStringToMap(tuc.Vars())
		content = cliutils.ReplaceVars(content, templateVars)
	}
	// Unmarshal template to a map
	var configMap map[string]interface{}
	err = json.Unmarshal(content, &configMap)
	return configMap, errorutils.CheckError(err)
}

func ValidateMapEntry(key string, value interface{}, writersMap map[string]AnswerWriter) error {
	if _, ok := writersMap[key]; !ok {
		return errorutils.CheckError(errors.New("template syntax error: unknown key: \"" + key + "\"."))
	}
	if _, ok := value.(string); !ok {
		return errorutils.CheckError(errors.New("template syntax error: the value for the  key: \"" + key + "\" is not a string type."))
	}
	return nil
}

func ValidateTemplatePath(templatePath string) error {
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
