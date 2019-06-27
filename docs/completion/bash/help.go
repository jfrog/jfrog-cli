package bash

const Description = "Generate bash completion script."

const UsageDescription = Description + `
To use it on your current bash session: 'source <(jfrog completion bash)'.

To make jfrog bash completion permanent:
  1. Generate completion bash script: 'jfrog completion bash > ~/.jfrog_completion_bash'
  2. Depending on your system configuration, add the following line to '~/.bashrc' or '~/.bash_profile': 'source ~/.jfrog_completion_bash'.
  3. Source bashrc or bash_profile respectively: 'source ~/.bashrc' or 'source ~/.bash_profile'`

var Usage = []string{"jfrog completion bash"}
