package utils

import (
	"errors"

	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

func ValidateMapEntry(key string, value interface{}, writersMap map[string]AnswerWriter) error {
	if _, ok := writersMap[key]; !ok {
		return errorutils.CheckError(errors.New("template syntax error: unknown key: \"" + key + "\"."))
	}
	if _, ok := value.(string); !ok {
		return errorutils.CheckError(errors.New("template syntax error: the value for the  key: \"" + key + "\" is not a string type."))
	}
	return nil
}
