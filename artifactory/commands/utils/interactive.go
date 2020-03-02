package utils

import (
	"fmt"
	goPrompt "github.com/c-bata/go-prompt"
	"regexp"
)

const (
	PressTabMsg      = " (press Tab for options):"
	InvalidAnswerMsg = "Invalid answer. Please select value from the suggestions list."
	VariableUseMsg   = " You may use dynamic variable in the form of ${key}."
	EmptyValueMsg    = "The value can not be empty. Please enter a valid value:"
)

var VarPattern = regexp.MustCompile(`^\$\{\w+\}+$`)

func prefixCompleter(options []goPrompt.Suggest) goPrompt.Completer {
	return func(document goPrompt.Document) []goPrompt.Suggest {
		return goPrompt.FilterHasPrefix(options, document.GetWordBeforeCursor(), true)
	}
}

func AskString(msg, promptPrefix string) string {
	if msg != "" {
		fmt.Println(msg + ":")
	}
	for {
		answer := goPrompt.Input(promptPrefix+" ", prefixCompleter(nil))
		if answer != "" {
			return answer
		}
		fmt.Println(EmptyValueMsg)
	}
}

func AskFromList(msg, promptPrefix string, allowVars bool, options []goPrompt.Suggest) string {
	if msg != "" {
		fmt.Println(msg + PressTabMsg)
	}
	errMsg := InvalidAnswerMsg
	if allowVars {
		errMsg += VariableUseMsg
	}
	for {
		answer := goPrompt.Input(promptPrefix+" ", prefixCompleter(options))
		if validateAnswer(answer, options, allowVars) {
			return answer
		}
		fmt.Println(errMsg)
	}
}

func validateAnswer(answer string, options []goPrompt.Suggest, allowVars bool) bool {
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
