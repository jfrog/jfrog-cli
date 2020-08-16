package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/jfrog/jfrog-cli/utils/cliutils"

	"github.com/c-bata/go-prompt"
)

const (
	InsertValuePromptMsg = "Insert the value for "
	DummyDefaultAnswer   = "-"
)

// The interactive questionnaire works as follows:
//	We have to provide a map of QuestionInfo which include all possible questions may be asked.
//	1. Mandatory Questions:
//		* We will ask all the questions in MandatoryQuestionsKeys list one after the other.
//	2. Optional questions:
//		* We have to provide a slice of prompt.Suggest, in which each suggest.Text is a key of a question in the map.
//		* After a suggest was chosen from the list, the corresponding question from the map will be asked.
//		* Each answer is written to to the configMap using its writer, under the MapKey specified in the questionInfo.
//		* We will execute the previous step until the SaveAndExit string was inserted.
type InteractiveQuestionnaire struct {
	QuestionsMap           map[string]QuestionInfo
	MandatoryQuestionsKeys []string
	OptionalKeysSuggests   []prompt.Suggest
	AnswersMap             map[string]interface{}
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
	EmptyValueMsg    = "The value cannot be empty. Please enter a valid value."
	OptionalKey      = "OptionalKey"
	SaveAndExit      = ":x"

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

// Bind ctrl+c key to interrupt the command
func interruptKeyBind() prompt.Option {
	interrupt := prompt.KeyBind{
		Key: prompt.ControlC,
		Fn: func(buf *prompt.Buffer) {
			panic("Interrupted")
		},
	}
	return prompt.OptionAddKeyBind(interrupt)
}

// Ask question with free string answer.
// If answer is empty and defaultValue isn't, return defaultValue.
// Otherwise, answer cannot be empty.
// Variable aren't checked and can be part of the answer.
func AskStringWithDefault(msg, promptPrefix, defaultValue string) string {
	return askString(msg, promptPrefix, defaultValue, false, false)
}

// Ask question with free string answer, allow an empty string as an answer
func AskString(msg, promptPrefix string, allowEmpty bool, allowVars bool) string {
	return askString(msg, promptPrefix, "", allowEmpty, allowVars)
}

// Ask question with free string answer.
// If an empty answer is allowed, the answer returned as is,
// if not and a default value was provided, the default value is returned.
func askString(msg, promptPrefix, defaultValue string, allowEmpty bool, allowVars bool) string {
	if msg != "" {
		fmt.Println(msg + ":")
	}
	errMsg := EmptyValueMsg
	if allowVars {
		errMsg += VariableUseMsg
	}
	promptPrefix = addDefaultValueToPrompt(promptPrefix, defaultValue)
	for {
		answer := prompt.Input(promptPrefix, prefixCompleter(nil), interruptKeyBind())
		answer = strings.TrimSpace(answer)
		if allowEmpty || answer != "" {
			return answer
		}
		// An empty answer wan given, default value wad provided.
		if defaultValue != "" {
			return defaultValue
		}
		fmt.Println(errMsg)
	}
}

// Ask question with list of possible answers.
// If answer is empty and defaultValue isn't, return defaultValue.
// Otherwise, the answer must be chosen from the list, but can be a variable if allowVars set to true.
func AskFromList(msg, promptPrefix string, allowVars bool, options []prompt.Suggest, defaultValue string) string {
	if msg != "" {
		fmt.Println(msg + PressTabMsg)
	}
	errMsg := InvalidAnswerMsg
	if allowVars {
		errMsg += VariableUseMsg
	}
	promptPrefix = addDefaultValueToPrompt(promptPrefix, defaultValue)
	for {
		answer := prompt.Input(promptPrefix, prefixCompleter(options), interruptKeyBind())
		answer = strings.TrimSpace(answer)
		if answer == "" && defaultValue != "" {
			return defaultValue
		}
		if validateAnswer(answer, options, allowVars) {
			return answer
		}
		fmt.Println(errMsg)
	}
}

func addDefaultValueToPrompt(promptPrefix, defaultValue string) string {
	if defaultValue != "" && defaultValue != DummyDefaultAnswer {
		return promptPrefix + " [" + defaultValue + "]: "
	}
	return promptPrefix + " "
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

// Ask question with list of possible answers.
// If the provided answer does not appear in list, confirm the choice.
func AskFromListWithMismatchConfirmation(promptPrefix, misMatchMsg string, options []prompt.Suggest) string {
	for {
		answer := prompt.Input(promptPrefix+" ", prefixCompleter(options), interruptKeyBind())
		if answer == "" {
			fmt.Println(EmptyValueMsg)
		}
		for _, option := range options {
			if answer == option.Text {
				return answer
			}
		}
		if cliutils.AskYesNo(misMatchMsg+" continue anyway?", false) {
			return answer
		}
	}
}

// Ask question steps:
// 		1. Ask for string/from list
//		2. Write the answer to answersMap (if writer provided)
// 		3. Run callback (if provided)
func (iq *InteractiveQuestionnaire) AskQuestion(question QuestionInfo) (value string, err error) {
	var answer string
	if question.Options != nil {
		answer = AskFromList(question.Msg, question.PromptPrefix, question.AllowVars, question.Options, "")
	} else {
		answer = AskString(question.Msg, question.PromptPrefix, false, question.AllowVars)
	}
	if question.Writer != nil {
		err = question.Writer(&iq.AnswersMap, question.MapKey, answer)
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
	iq.AnswersMap = make(map[string]interface{})
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
		if key == SaveAndExit {
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

func ConvertToSuggests(options []string) []prompt.Suggest {
	var suggests []prompt.Suggest
	for _, opt := range options {
		suggests = append(suggests, prompt.Suggest{Text: opt})
	}
	return suggests
}

// After an optional value was chosen we'll ask for its value.
func OptionalKeyCallback(iq *InteractiveQuestionnaire, key string) (value string, err error) {
	if key != SaveAndExit {
		valueQuestion := iq.QuestionsMap[key]
		// Since we are using default question in most of the cases we set the map key here.
		valueQuestion.MapKey = key
		valueQuestion.PromptPrefix = InsertValuePromptMsg + key
		if valueQuestion.Options != nil {
			valueQuestion.PromptPrefix += PressTabMsg
		}
		valueQuestion.PromptPrefix += " >"
		value, err = iq.AskQuestion(valueQuestion)
	}
	return value, err
}
