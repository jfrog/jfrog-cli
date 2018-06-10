package prompt

import (
	"errors"
	"github.com/chzyer/readline"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
	"github.com/spf13/viper"
	"regexp"
	"strings"
)

type Simple struct {
	// Prompt message
	Msg string

	// Default value
	Default string

	// Mask user input, mainly for passwords input
	Mask bool

	// Return value will be saved under Label key
	Label string

	// private result
	results
}

type Array struct {
	// Prompt container for sequence of prompts
	Prompts []Prompt
}

type YesNo struct {
	// Prompt message
	Msg string

	// Default value
	Default string

	// Prompt container for positive outcome
	Yes Prompt

	// Prompt container for negative outcome
	No Prompt

	// Return value will be saved under Label key
	Label string

	// private result
	results
}

type Autocomplete struct {
	// Prompt message
	Msg string

	// Autocomplete selection options
	Options []string

	// Default option
	Default string

	// Prompt error massage for incorrect input
	ErrMsg string

	// Prompt confirmation massage for incorrect input
	ConfirmationMsg string

	// Default option for confirmation prompt
	ConfirmationDefault string

	// Return value will be saved under Label key
	Label string

	// private result
	results
}

type Prompt interface {
	// Read the input from the user
	Read() error
	// Get results
	GetResults() *viper.Viper
}

type results struct {
	// Results of user input, reading multiple times will override existing user input.
	Result *viper.Viper
}

func (results *results) set(key string, value interface{}) {
	if results.Result == nil {
		results.Result = viper.New()
	}
	if key != "" {
		results.Result.Set(key, value)
	}
}

func (autocomplete *Autocomplete) Read() error {
	msg := replaceDefault(autocomplete.Msg, autocomplete.Default)
	completer := createCompleter(autocomplete.Options)
	l, err := createAutocompletePrompt(msg, completer, false)
	if err != nil {
		return err
	}
	defer l.Close()

	for {
		line, err := l.Readline()
		line = strings.TrimSpace(line)
		if errorutils.CheckError(err) != nil {
			return err
		}
		if line == "" && autocomplete.Default != "" {
			line = autocomplete.Default
		}
		if line != "" && len(autocomplete.Options) == 0 {
			autocomplete.set(autocomplete.Label, line)
			return nil
		}
		for _, v := range autocomplete.Options {
			if v == line {
				autocomplete.set(autocomplete.Label, line)
				return nil
			}
		}
		if autocomplete.ErrMsg != "" {
			log.Error(autocomplete.ErrMsg)
			continue
		}
		if autocomplete.ConfirmationMsg != "" {
			confirm := &YesNo{Msg: autocomplete.ConfirmationMsg, Default: autocomplete.ConfirmationDefault, Label: "defaultConfirm"}
			confirm.Read()
			if confirm.GetResults().GetBool(confirm.Label) {
				autocomplete.set(autocomplete.Label, line)
				return nil
			}
		}
	}
	return nil
}

func (autocomplete *Autocomplete) GetResults() *viper.Viper {
	return autocomplete.Result
}

func (simple *Simple) Read() error {
	msg := replaceDefault(simple.Msg, simple.Default)
	l, err := createAutocompletePrompt(msg, nil, simple.Mask)
	if err != nil {
		return err
	}
	defer l.Close()

	line, err := l.Readline()
	if errorutils.CheckError(err) != nil {
		return err
	}

	line = strings.TrimSpace(line)
	if line == "" && simple.Default != "" {
		line = simple.Default
	}

	simple.set(simple.Label, line)
	return nil
}

func (simple *Simple) GetResults() *viper.Viper {
	return simple.Result
}

func (yesNo *YesNo) Read() error {
	result, err := readYesNoQuestion(yesNo.Msg, yesNo.Default)
	if errorutils.CheckError(err) != nil {
		return err
	}

	yesNo.set(yesNo.Label, result)
	if result && yesNo.Yes != nil {
		return yesNo.Yes.Read()
	}

	if yesNo.No != nil {
		return yesNo.No.Read()
	}

	return nil
}

func (yesNo *YesNo) GetResults() *viper.Viper {
	if yesNo.Result == nil {
		yesNo.Result = viper.New()
	}

	if yesNo.Yes != nil {
		appendViperConfig(yesNo.Result, yesNo.Yes.GetResults())
	}
	if yesNo.No != nil {
		appendViperConfig(yesNo.Result, yesNo.No.GetResults())
	}

	return yesNo.Result
}

func (array *Array) Read() error {
	for _, v := range array.Prompts {
		err := v.Read()
		if errorutils.CheckError(err) != nil {
			return err
		}
	}
	return nil
}

func (array *Array) GetResults() *viper.Viper {
	vConfig := viper.New()
	for _, v := range array.Prompts {
		appendViperConfig(vConfig, v.GetResults())
	}
	return vConfig
}

func appendViperConfig(dest, src *viper.Viper) {
	if dest == nil || src == nil {
		return
	}
	for _, v := range src.AllKeys() {
		dest.Set(v, src.Get(v))
	}
}

func readYesNoQuestion(prompt, defaultVal string) (bool, error) {
	errMsg := "Please enter a valid option."
	msg := replaceDefault(prompt, defaultVal)
	l, err := createAutocompletePrompt(msg, nil, false)
	if err != nil {
		return false, err
	}
	defer l.Close()

	for {
		line, err := l.Readline()
		if errorutils.CheckError(err) != nil {
			return false, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			line = defaultVal
		}
		if boolVal, err := parseYesNo(line); err == nil {
			return boolVal, nil
		}
		log.Error(errMsg)
	}
}

func createAutocompletePrompt(msg string, completer readline.AutoCompleter, mask bool) (*readline.Instance, error) {
	l, err := readline.NewEx(&readline.Config{
		Prompt:                 msg,
		DisableAutoSaveHistory: true,
		AutoComplete:           completer,
		InterruptPrompt:        "\n",
		EOFPrompt:              "exit",
		EnableMask:             mask,

		FuncFilterInputRune: filterInput,
	})
	return l, errorutils.CheckError(err)
}

func parseYesNo(s string) (bool, error) {
	matchedYes, err := regexp.MatchString("^yes$|^y$", strings.ToLower(s))
	if errorutils.CheckError(err) != nil {
		return matchedYes, err
	}
	if matchedYes {
		return true, nil
	}

	matchedNo, err := regexp.MatchString("^no$|^n$", strings.ToLower(s))
	if errorutils.CheckError(err) != nil {
		return matchedNo, err
	}
	if matchedNo {
		return false, nil
	}
	return false, errors.New(s + " is not yes or no pattern")
}

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}

	return r, true
}

func createCompleter(values []string) *readline.PrefixCompleter {
	pcItems := []readline.PrefixCompleterInterface{}

	for _, v := range values {
		pcItems = append(pcItems, readline.PcItem(v))
	}

	return readline.NewPrefixCompleter(pcItems...)
}

func replaceDefault(msg, defaultVal string) string {
	return strings.Replace(msg, "${default}", defaultVal, -1)
}
