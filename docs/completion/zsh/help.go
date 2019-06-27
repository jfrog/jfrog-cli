package zsh

const Description = `Generate zsh completion script.
To use it on your current zsh session: 'source <(jfrog completion zsh)'.

To make jfrog zsh completion permanent:
1. Generate completion zsh script: 'jfrog completion zsh > ~/.jfrog_completion_zsh'
2. Add the following line to ~/.zshrc file: 'source ~/.jfrog_completion_zsh'.
3. Source your zshrc: 'source ~/.zshrc'`

var Usage = []string{"jfrog completion zsh"}
