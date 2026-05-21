package exportcmd

import "github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"

var Usage = []string{"config export [server ID]"}

func GetDescription() string {
	return `Creates a server configuration token. The generated Config Token can be imported by the "` + coreutils.GetCliExecutableName() + ` config import <Config Token>" command.`
}

func GetAIDescription() string {
	return `Emit a server configuration as an opaque Config Token string. Pipe the output to 'jf c import' on another machine to transfer the configuration in one step. If no server ID is provided, the default server is exported.

When to use:
- Provisioning identical configurations across multiple machines or CI runners.
- Backing up a configuration before a rotation or migration.

Prerequisites:
- At least one configured server (see 'jf c show').

Common patterns:
  $ jf c export my-server
  $ jf c export  # exports default server

Gotchas:
- The token embeds credentials; redact or treat as secret. Do not paste into logs or commit to version control.
- The token is not human-readable but is reversible; security is "obfuscated", not encrypted.

Related: jf c import, jf c show`
}
