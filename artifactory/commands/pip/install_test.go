package pip

import (
	"bytes"
	"fmt"
	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/pip"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	logUtils "github.com/jfrog/jfrog-cli-go/utils/log"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"os"
	"testing"
)

func TestRunStam(t *testing.T) {
	newLog := log.NewLogger(logUtils.GetCliLogLevel(), nil)
	buffer := &bytes.Buffer{}
	newLog.SetOutputWriter(buffer)
	log.SetLogger(newLog)

	pic := NewPipInstallCommand()
	pic.Run()
}

func TestRunCmd(t *testing.T) {
	newLog := log.NewLogger(logUtils.GetCliLogLevel(), nil)
	buffer := &bytes.Buffer{}
	newLog.SetOutputWriter(buffer)
	log.SetLogger(newLog)


	pipInstallCmd := &pip.PipCmd{
		Executable:  "sh",
		Command:     "-c",
		CommandArgs: []string{"env"},
		EnvVars:     nil,
		StrWriter:   nil,
		ErrWriter:   nil,
	}
	gofrogcmd.RunCmd(pipInstallCmd)
}

func TestExecutePipInstallWithLogParsing(t *testing.T) {
	newLog := log.NewLogger(logUtils.GetCliLogLevel(), nil)
	buffer := &bytes.Buffer{}
	newLog.SetOutputWriter(buffer)
	log.SetLogger(newLog)
	os.Setenv("PATH", "/Users/barb/trash/venv-test3/bin:/usr/local/opt/node@8/bin:/Users/barb/.sdkman/candidates/maven/current/bin:/Users/barb/.sdkman/candidates/java/current/bin:/Users/barb/.sdkman/candidates/groovy/current/bin:/usr/local/bin:/usr/local/sbin:/usr/local/opt/coreutils/libexec/gnubin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:/usr/local/go/bin:/Applications/Wireshark.app/Contents/MacOS:/Users/barb/install/gradle-4.6/bin:/Users/barb/go/bin")

	// Create pip cli command
	cmd := &PipInstallCommand{}

	// Create executable command
	pipInstallCmd := &pip.PipCmd{
		Executable:  "pip",
		Command:     "install",
		CommandArgs: []string{"macholib", "-v"},
		EnvVars:     nil,
		StrWriter:   nil,
		ErrWriter:   nil,
	}

	cmd.executePipInstallWithLogParsing(pipInstallCmd)
	log.Info(fmt.Sprintf("Result:\n%v", cmd.dependencyToFileMap))
}

func TestGetDependencyChecksumFromArtifactory(t *testing.T) {
	newLog := log.NewLogger(logUtils.GetCliLogLevel(), nil)
	buffer := &bytes.Buffer{}
	newLog.SetOutputWriter(buffer)
	log.SetLogger(newLog)

	// Create CLI pip install command.
	artifactoryDetails := &config.ArtifactoryDetails{Url: "http://localhost:8081/artifactory/", User: "admin", Password: "password"}
	cmd := &PipInstallCommand{rtDetails: artifactoryDetails, dependencyToFileMap: map[string]string{"macholib": "macholib-1.11-py2.py3-none-any.whl", "altgraph": "altgraph-0.16.1-py2.py3-none-any.whl"}, pypiRepo: "pypi"}

	// Init dependency map.
	depMap := map[string]*buildinfo.Dependency{"macholib": {Id:"macholib:1.11", Scopes: []string{}, Checksum: &buildinfo.Checksum{}}, "altgraph": {Id:"altgraph:0.16.1", Scopes: []string{}, Checksum: &buildinfo.Checksum{}}, "non-existing-pkg": {Id:"non-existing-pkg:1.0.0", Scopes: []string{}, Checksum: &buildinfo.Checksum{}}}

	// Get artifacts details.
	cmd.populateDependenciesInfoAndPromptMissingDependencies(depMap)

	log.Info(fmt.Sprintf("Result:\n%v", depMap))
}
