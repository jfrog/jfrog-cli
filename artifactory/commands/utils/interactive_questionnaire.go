package utils

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"regexp"
	"strconv"
	"strings"
)

// The interactive questionnaire works as follows:
//	We have to provide a map of QuestionInfo which include all possible questions may be asked.
//	1. Mandatory Questions:
//		* We will ask all the questions in MandatoryQuestionsKeys list one after the other.
//	2. Optional questions:
//		* we have to provide a slice of prompt.Suggest, in which each suggest.Text is a key of a question in the map.
//		* after a suggest was chosen from the list, the corresponding question from the map will be asked.
//		* aach answer is written to to the configMap using its writer, under the MapKey specified in the questionInfo.
//		* we will execute the previous step until the WriteAndExit string was inserted.
type InteractiveQuestionnaire struct {
	QuestionsMap           map[string]QuestionInfo
	MandatoryQuestionsKeys []string
	OptionalKeysSuggests   []prompt.Suggest
	AnswersMap             map[string]string
}

// Each question can have the following properties:
// 		* Msg - will be printed in separate line
// 		* PromptPrefix - will be printed before the input cursor in the answer line
// 		* Options - In case the answer must be selected from a predefined list
// 		* AllowVars - a flag indicates whether a variable (in form of ${var}) is an acceptable answer despite the predefined list
// 		* Writer - how to write the answer to the final config map
// 		* MapKey - the key under which the answer will be written to the configMap
// 		* Callback - optional function can be executed after the answer was inserted. Can be used to implement some dependencies between questions.
type AnswerWriter func(resultMap *map[string]interface{}, key, value string) error
type questionCallback func(*InteractiveQuestionnaire, string) (string, error)

type QuestionInfo struct {
	Msg          string
	PromptPrefix string
	Options      []prompt.Suggest
	AllowVars    bool
	Writer       AnswerWriter
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

	// Boolean answers
	True  = "true"
	False = "false"

	CommaSeparatedListMsg = "The value should be a comma separated list"
)

// Var can be inserted in the form of ${key}
var VarPattern = regexp.MustCompile(`^\$\{\w+\}+$`)

func prefixCompleter(options []prompt.Suggest) prompt.Completer {
	return func(document prompt.Document) []prompt.Suggest {
		return prompt.FilterHasPrefix(options, document.GetWordBeforeCursor(), true)
	}
}

// Ask question with free string answer, answer cannot be empty.
// Variable aren't check and can be part of the answer
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

// Ask question with list of possible answers.
// The answer must be chosen from the list, but can be a variable if allowVars set to true.
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

// Ask question steps:
// 		1. Ask for string/from list
//		2. Write the answer to answersMap (if writer provided)
// 		3. Run callback (if provided)q
func (iq *InteractiveQuestionnaire) AskQuestion(question QuestionInfo) (value string, err error) {

	var answer string
	if question.Options != nil {
		answer = AskFromList(question.Msg, question.PromptPrefix, question.AllowVars, question.Options)
	} else {
		answer = AskString(question.Msg, question.PromptPrefix)
	}
	if question.Writer != nil {
		//err = question.Writer(&iq.AnswersMap, question.MapKey, answer)
		iq.AnswersMap[question.MapKey] = answer
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

// The main function to perform the questionnaire
func (iq *InteractiveQuestionnaire) Perform() error {
	iq.AnswersMap = make(map[string]string)
	for i := 0; i < len(iq.MandatoryQuestionsKeys); i++ {
		iq.AskQuestion(iq.QuestionsMap[iq.MandatoryQuestionsKeys[i]])
	}
	fmt.Println("You can type \":x\" at any time to save and exit.")
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

// Common questions
var FreeStringQuestionInfo = QuestionInfo{
	Options:   nil,
	AllowVars: false,
	Writer:    WriteStringAnswer,
}

func GetBoolSuggests() []prompt.Suggest {
	return []prompt.Suggest{
		{Text: True},
		{Text: False},
	}
}

var BoolQuestionInfo = QuestionInfo{
	Options:   GetBoolSuggests(),
	AllowVars: true,
	Writer:    WriteBoolAnswer,
}

var IntQuestionInfo = QuestionInfo{
	Options:   nil,
	AllowVars: true,
	Writer:    WriteIntAnswer,
}

var StringListQuestionInfo = QuestionInfo{
	Msg:       CommaSeparatedListMsg,
	Options:   nil,
	AllowVars: true,
	Writer:    WriteStringArrayAnswer,
}

// Common writers
func WriteStringAnswer(resultMap *map[string]interface{}, key, value string) error {
	(*resultMap)[key] = value
	return nil
}

func WriteBoolAnswer(resultMap *map[string]interface{}, key, value string) error {
	if regexMatch := VarPattern.FindStringSubmatch(value); regexMatch != nil {
		return WriteStringAnswer(resultMap, key, value)
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}
	(*resultMap)[key] = boolValue
	return nil
}

func WriteIntAnswer(resultMap *map[string]interface{}, key, value string) error {
	if regexMatch := VarPattern.FindStringSubmatch(value); regexMatch != nil {
		return WriteStringAnswer(resultMap, key, value)
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	(*resultMap)[key] = intValue
	return nil
}

func WriteStringArrayAnswer(resultMap *map[string]interface{}, key, value string) error {
	if regexMatch := VarPattern.FindStringSubmatch(value); regexMatch != nil {
		return WriteStringAnswer(resultMap, key, value)
	}
	arrValue := strings.Split(value, ",")
	(*resultMap)[key] = arrValue
	return nil
}

func GetSuggestsFromKeys(keys []string, SuggestionMap map[string]prompt.Suggest) []prompt.Suggest {
	var suggests []prompt.Suggest
	for _, key := range keys {
		suggests = append(suggests, SuggestionMap[key])
	}
	return suggests
}
