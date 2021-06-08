package auditnpm

const Description = "Execute an audit npm command, using the configured Xray details."

var Usage = []string{`jfrog xr audit-npm  <path> [command options]`}

const Arguments string = `	path
path to npm project source code. path is optional. If not provided, current dir is assumed.`
