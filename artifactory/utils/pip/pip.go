package pip

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

// Get executable path.
// If run inside a virtual-env, this should return the path for the correct executable.
func GetExecutablePath(executableName string) (string, error) {
	executablePath, err := exec.LookPath(executableName)
	if err != nil {
		return "", errorutils.CheckError(err)
	}

	if executablePath == "" {
		return "", errorutils.CheckError(errors.New(fmt.Sprintf("Could not find '%s' executable", executableName)))
	}

	log.Debug(fmt.Sprintf("Found %s executable at: %s", executableName, executablePath))
	return executablePath, nil
}

func (pc *PipCmd) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, pc.Executable)
	cmd = append(cmd, pc.Command)
	cmd = append(cmd, pc.CommandArgs...)
	return exec.Command(cmd[0], cmd[1:]...)
}

func (pc *PipCmd) GetEnv() map[string]string {
	return pc.EnvVars
}

func (pc *PipCmd) GetStdWriter() io.WriteCloser {
	return pc.StrWriter
}

func (pc *PipCmd) GetErrWriter() io.WriteCloser {
	return pc.ErrWriter
}

// Return path to the dependency-tree script, if not exists -> create the file.
func GetDepTreeScriptPath() (string, error) {
	pipDependenciesPath, err := config.GetJfrogDependenciesPath()
	if err != nil {
		return "", err
	}
	pipDependenciesPath = filepath.Join(pipDependenciesPath, "pip", pipDepTreeVersion)
	depTreeScriptName := "pipdeptree.py"
	depTreeScriptPath := path.Join(pipDependenciesPath, depTreeScriptName)
	err = writeScriptIfNeeded(pipDependenciesPath, depTreeScriptName)

	return depTreeScriptPath, err
}

func writeScriptIfNeeded(targetDirPath, scriptName string) error {
	scriptPath := path.Join(targetDirPath, scriptName)
	exists, err := fileutils.IsFileExists(scriptPath, false)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if !exists {
		err = os.MkdirAll(targetDirPath, os.ModeDir|os.ModePerm)
		if errorutils.CheckError(err) != nil {
			return err
		}
		err = ioutil.WriteFile(scriptPath, pipDepTreeContent, os.ModePerm)
		if errorutils.CheckError(err) != nil {
			return err
		}
	}
	return nil
}

type PipCmd struct {
	Executable  string
	Command     string
	CommandArgs []string
	EnvVars     map[string]string
	StrWriter   io.WriteCloser
	ErrWriter   io.WriteCloser
}

func GetDependencyChecksumFromArtifactory(servicesManager *artifactory.ArtifactoryServicesManager, repository, dependencyFile string) (*buildinfo.Checksum, error) {
	log.Debug(fmt.Sprintf("Fetching checksums for: %s", dependencyFile))
	result, err := servicesManager.Aql(createAqlQueryForPypi(repository, dependencyFile))
	if err != nil {
		return nil, err
	}

	parsedResult := new(aqlResult)
	err = json.Unmarshal(result, parsedResult)
	if err = errorutils.CheckError(err); err != nil {
		return nil, err
	}
	if len(parsedResult.Results) == 0 {
		log.Debug(fmt.Sprintf("File: %s could not be found in repository: %s", dependencyFile, repository))
		return nil, nil
	}

	// Verify checksum exist.
	sha1 := parsedResult.Results[0].Actual_sha1
	md5 := parsedResult.Results[0].Actual_md5
	if sha1 == "" || md5 == "" {
		// Missing checksum.
		log.Debug(fmt.Sprintf("Missing checksums for file: %s, sha1: '%s', md5: '%s'", dependencyFile, sha1, md5))
		return nil, nil
	}

	// Update checksum.
	checksum := &buildinfo.Checksum{Sha1: sha1, Md5: md5}
	log.Debug(fmt.Sprintf("Found checksums for file: %s, sha1: '%s', md5: '%s'", dependencyFile, sha1, md5))

	return checksum, nil
}

// TODO: Move this function to jfrog-client-go/artifactory/services/utils/aqlquerybuilder.go
func createAqlQueryForPypi(repo, file string) string {
	itemsPart :=
		`items.find({` +
			`"repo": "%s",` +
			`"$or": [{` +
			`"$and":[{` +
			`"path": {"$match": "*"},` +
			`"name": {"$match": "%s"}` +
			`}]` +
			`}]` +
			`}).include("actual_md5","actual_sha1")`
	return fmt.Sprintf(itemsPart, repo, file)
}

type aqlResult struct {
	Results []*results `json:"results,omitempty"`
}

type results struct {
	Actual_md5  string `json:"actual_md5,omitempty"`
	Actual_sha1 string `json:"actual_sha1,omitempty"`
}
