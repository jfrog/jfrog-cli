package common

import "strings"

func CreateEnvVars(envVars ...string) string {
	var s []string
	for _, envVar := range envVars {
		s = append(s, envVar)
	}
	return strings.Join(s[:], "\n\n")
}
