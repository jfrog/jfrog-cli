package completion

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-go/docs/common"
	"github.com/jfrog/jfrog-cli-go/docs/completion/bash"
	"github.com/jfrog/jfrog-cli-go/docs/completion/zsh"
)

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
	zshAutocomplete = `
_jfrog() {
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
			Name:     "bash",
			Usage:    bash.Description,
			HelpName: common.CreateUsage("completion bash", bash.UsageDescription, bash.Usage),
			Action: func(*cli.Context) {
				fmt.Print(bashAutocomplete)
			},
		},
		{
			Name:     "zsh",
			Usage:    zsh.Description,
			HelpName: common.CreateUsage("completion zsh", zsh.UsageDescription, zsh.Usage),
			Action: func(*cli.Context) {
				fmt.Print(zshAutocomplete)
			},
		},
	}
}
