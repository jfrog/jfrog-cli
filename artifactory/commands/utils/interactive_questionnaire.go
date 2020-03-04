package utils

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"regexp"
)

type InteractiveQuestionnaire struct {
	MandatoryQuestionsKeys []string
	OptionalKeysSuggests   []prompt.Suggest
	QuestionsMap           map[string]QuestionInfo
	ConfigMap              map[string]string
}

type answerWriter func(resultMap *map[string]interface{}, key, value string) error
type questionCallback func(*InteractiveQuestionnaire, string) (string, error)

type QuestionInfo struct {
	Msg          string
	PromptPrefix string
	Options      []prompt.Suggest
	AllowVars    bool
	Writer       answerWriter
	MapKey       string
	Callback     questionCallback
}

const (
	PressTabMsg      = " (press Tab for options):"
	InvalidAnswerMsg = "Invalid answer. Please select value from the suggestions list."
	VariableUseMsg   = " You may use dynamic variable in the form of ${key}."
	EmptyValueMsg    = "The value can not be empty. Please enter a valid value:"
	OptionalKey      = "OptionalKey"
	WriteAndExist    = ":x"
)

var VarPattern = regexp.MustCompile(`^\$\{\w+\}+$`)

func prefixCompleter(options []prompt.Suggest) prompt.Completer {
	return func(document prompt.Document) []prompt.Suggest {
		return prompt.FilterHasPrefix(options, document.GetWordBeforeCursor(), true)
	}
}

func AskString(msg, promptPrefix string) string {
	if msg != "" {
		fmt.Println(msg + ":")
	}
	for {
		answer := prompt.Input(promptPrefix+" ", prefixCompleter(nil))
		if answer != "" {
			return answer
		}
		fmt.Println(EmptyValueMsg)
	}
}

func AskFromList(msg, promptPrefix string, allowVars bool, options []prompt.Suggest) string {
	if msg != "" {
		fmt.Println(msg + PressTabMsg)
	}
	errMsg := InvalidAnswerMsg
	if allowVars {
		errMsg += VariableUseMsg
	}
	for {
		answer := prompt.Input(promptPrefix+" ", prefixCompleter(options))
		if validateAnswer(answer, options, allowVars) {
			return answer
		}
		fmt.Println(errMsg)
	}
}

func validateAnswer(answer string, options []prompt.Suggest, allowVars bool) bool {
	if allowVars {
		if regexMatch := VarPattern.FindStringSubmatch(answer); regexMatch != nil {
			return true
		}
	}
	for _, option := range options {
		if answer == option.Text {
			return true
		}
	}
	return false
}

func (iq *InteractiveQuestionnaire) AskQuestion(question QuestionInfo) (value string, err error) {

	var answer string
	if question.Options != nil {
		answer = AskFromList(question.Msg, question.PromptPrefix, question.AllowVars, question.Options)
	} else {
		answer = AskString(question.Msg, question.PromptPrefix)
	}
	if question.Writer != nil {
		//err = question.Writer(&iq.ConfigMap, question.MapKey, answer)
		iq.ConfigMap[question.MapKey] = answer
		if err != nil {
			return "", err
		}
	}
	if question.Callback != nil {
		_, err = question.Callback(iq, answer)
		if err != nil {
			return "", err
		}
	}
	return answer, nil
}

func (iq *InteractiveQuestionnaire) Perform() error {
	iq.ConfigMap = make(map[string]string)
	for i := 0; i < len(iq.MandatoryQuestionsKeys); i++ {
		iq.AskQuestion(iq.QuestionsMap[iq.MandatoryQuestionsKeys[i]])
	}
	OptionalKeyQuestion := iq.QuestionsMap[OptionalKey]
	OptionalKeyQuestion.Options = iq.OptionalKeysSuggests
	for {
		key, err := iq.AskQuestion(OptionalKeyQuestion)
		if err != nil {
			return err
		}
		if key == WriteAndExist {
			break
		}
	}
	return nil
}
