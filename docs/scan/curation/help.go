package curation

var Usage = []string{"curation-audit [command options]"}

func GetDescription() string {
	return "curation audit your local project's dependencies by generating a dependency tree and check curation status."
}
