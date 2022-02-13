package zsh

//go:generate go run ../generate_scripts.go

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io/ioutil"
	"path/filepath"
)

const ZshAutocomplete = `#compdef _jf jf _jfrog jfrog

_jfrog() {
	local -a opts
	opts=("${(@f)$(_CLI_ZSH_AUTOCOMPLETE_HACK=1 ${words[@]:0:#words[@]-1} --generate-bash-completion)}")
	_describe 'values' opts
	if [[ $compstate[nmatches] -eq 0 && $words[$CURRENT] != -* ]]; then
		_files
	fi
}

compdef _jfrog jfrog
compdef _jfrog jf`

func WriteZshCompletionScript(install bool) {
	if !install {
		fmt.Print(ZshAutocomplete)
		return
	}
	homeDir, err := coreutils.GetJfrogHomeDir()
	if err != nil {
		log.Error(err)
		return
	}
	completionPath := filepath.Join(homeDir, "jfrog_zsh_completion")
	if err = ioutil.WriteFile(completionPath, []byte(ZshAutocomplete), 0600); err != nil {
		log.Error(err)
		return
	}
	sourceCommand := "source " + completionPath + ""
	fmt.Printf(`Generated zsh completion script at %s.
To activate auto-completion on this shell only, source the completion script by running the following three commands:

autoload -Uz compinit
compinit
%s

To activate auto-completion permanently, put the above three commands in ~/.zshrc.

`,
		completionPath, sourceCommand)
}
