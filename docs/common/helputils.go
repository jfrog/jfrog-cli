package common

import "strings"

func CreateEnvVars(envVars ...string) string {
	s := append([]string{}, envVars...)
	return strings.Join(s[:], "\n\n")
}
