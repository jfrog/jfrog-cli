package fish

var Usage = []string{"completion fish"}

func GetDescription() string {
	return "Generate fish completion script."
}

func GetAIDescription() string {
	return `Emit a fish completion script for jf. Pipe it to ~/.config/fish/completions/jf.fish so the fish shell picks up tab completion for subcommands and flags.

When to use:
- One-time shell setup on a developer machine using fish.

Prerequisites:
- fish 3.0+.

Common patterns:
  $ jf completion fish > ~/.config/fish/completions/jf.fish
  $ jf completion fish --install

Gotchas:
- fish reads completions from ~/.config/fish/completions/; the file must be named jf.fish.

Related: jf completion bash, jf completion zsh`
}
