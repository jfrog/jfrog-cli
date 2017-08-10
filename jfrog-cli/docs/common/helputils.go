package common

import (
	"strings"
)

func CreateUsage(command string, name string, commands []string) string {
	return "\nName:\n\t" + "jfrog " + command + " - " + name + "\n\nUsage:\n\t" + strings.Join(commands[:], "\n\t") + "\n"
}

func CreateEnvVars(envVars ...string) string {
	var s []string
	for _, envVar := range envVars {
		s = append(s, envVar)
	}
	s = append(s, GlobalEnvVars)
	return strings.Join(s[:], "\n\n")
}