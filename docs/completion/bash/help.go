package bash

var Usage = []string{"completion bash"}

func GetDescription() string {
	return "Generate bash completion script."
}

func GetAIDescription() string {
	return `Emit a bash completion script for jf to standard output. Source it from ~/.bashrc or pipe it into a system completions directory to get tab completion for subcommands and flags.

When to use:
- One-time shell setup on a developer machine.
- Installing into a Docker base image so interactive shells get completions.

Prerequisites:
- bash 4.0+ and bash-completion installed.

Common patterns:
  $ jf completion bash > ~/.jfrog-completion.bash && echo 'source ~/.jfrog-completion.bash' >> ~/.bashrc
  $ jf completion bash --install

Gotchas:
- --install writes to a system path; may require elevated privileges.
- The generated script must be re-sourced (or shell reopened) to take effect.

Related: jf completion zsh, jf completion fish`
}
