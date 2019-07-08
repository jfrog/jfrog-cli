package completion

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-go/docs/common"
	"github.com/jfrog/jfrog-cli-go/docs/completion/bash"
	"github.com/jfrog/jfrog-cli-go/docs/completion/zsh"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io/ioutil"
	"path/filepath"
)

// Shell completion types
const (
	bashCompletionType = "bash"
	zshCompletionType  = "zsh"
)

// Shell completion scripts
const (
	bashAutocomplete = `#!/bin/bash
_jfrog() {
    local cur opts base
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    opts=$( ${COMP_WORDS[@]:0:$COMP_CWORD} --generate-bash-completion )
    COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
}

complete -F _jfrog -o default jfrog
`
	zshAutocomplete = `_jfrog() {
	local -a opts
	opts=("${(@f)$(_CLI_ZSH_AUTOCOMPLETE_HACK=1 ${words[@]:0:#words[@]-1} --generate-bash-completion)}")
	_describe 'values' opts
	if [[ $compstate[nmatches] -eq 0 && $words[$CURRENT] != -* ]]; then
		_files
	fi
}

compdef _jfrog jfrog`
)

func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:         "bash",
			Usage:        bash.Description,
			HelpName:     common.CreateUsage("completion bash", bash.Description, bash.Usage),
			BashComplete: common.CreateBashCompletionFunc(),
			Action: func(*cli.Context) {
				writeCompletionFile(bashCompletionType, bashAutocomplete)
			},
		},
		{
			Name:         "zsh",
			Usage:        zsh.Description,
			HelpName:     common.CreateUsage("completion zsh", zsh.Description, zsh.Usage),
			BashComplete: common.CreateBashCompletionFunc(),
			Action: func(*cli.Context) {
				writeCompletionFile(zshCompletionType, zshAutocomplete)
			},
		},
	}
}

func writeCompletionFile(completionType, completionScript string) {
	homeDir, err := config.GetJfrogHomeDir()
	if err != nil {
		log.Error(err)
		return
	}
	completionPath := filepath.Join(homeDir, "jfrog_"+completionType+"_completion")
	if err = ioutil.WriteFile(completionPath, []byte(completionScript), 0600); err != nil {
		log.Error(err)
		return
	}
	sourceCommand := "'source " + completionPath + "'"
	fmt.Printf(`Generated %s completion script at %s.
To activate auto-completion on this shell only, source the completion script: %s
To activate auto-completion permanently, put this line in ~/.bashrc, ~/.bash_profile or ~/.zshrc, depend on your system.

`,
		completionType, completionPath, sourceCommand)
}
