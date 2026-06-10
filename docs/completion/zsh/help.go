package zsh

var Usage = []string{"completion zsh"}

func GetDescription() string {
	return "Generate zsh completion script."
}

func GetAIDescription() string {
	return `Emit a zsh completion script for jf to standard output. Drop it into a directory on $fpath (usually ~/.zsh/completions/) and add 'autoload -U compinit && compinit' to ~/.zshrc.

When to use:
- One-time shell setup on a developer machine using zsh.

Prerequisites:
- zsh with the completion system enabled.

Common patterns:
  $ jf completion zsh > "${fpath[1]}/_jf"
  $ jf completion zsh --install

Gotchas:
- The completion file must be named '_jf' on $fpath for zsh to find it.
- After installing, run 'compinit' or reopen the shell.

Related: jf completion bash, jf completion fish`
}
