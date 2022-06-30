package common

import "strings"

func CreateEnvVars(envVars ...string) string {
	var s []string
	s = append(s, envVars...)
	return strings.Join(s[:], "\n\n")
}
