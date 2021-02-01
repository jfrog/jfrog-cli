package vcs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type vcsData struct {
	ProjectName      string
	LocalDirPath     string
	VcsBranch        string
	BuildCommand     string
	ArtifactoryRepos []services.RepositoryDetails
	// A collection of technologies that was found with a list of theirs indications
	Technologies     map[technology][]string
	VcsCredentials   auth.ServiceDetails
	JfrogCredentials auth.ServiceDetails
}

func VcsCmd(c *cli.Context) error {
	var data vcsData

	return fmt.Errorf("Not Impelemanted...")
}

func cloneProject(data vcsData) (err error) {
	// Create the desired path if necessary
	err = os.MkdirAll(data.LocalDirPath, os.ModePerm)
	if err != nil {
		return err
	}

	cloneOption := &git.CloneOptions{
		URL:           data.VcsCredentials.GetUrl(),
		ReferenceName: plumbing.ReferenceName(data.VcsBranch),
		Progress:      os.Stdout,
		Auth:          createCredentials(data.VcsCredentials),
		// Enable git submodules clone if there any.
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	}
	// Clone the given repository to the given directory from the given branch
	log.Info("git clone %s %s", data.VcsCredentials.GetUrl(), data.LocalDirPath)
	_, err = git.PlainClone(data.LocalDirPath, false, cloneOption)
	return
}

func detectTechnologies(data vcsData) (err error) {
	indicators := GetTechIndicators()
	filesList, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(filepath.Join(data.LocalDirPath, data.ProjectName), false)
	if err != nil {
		return err
	}
	for _, file := range filesList {
		for _, indicator := range indicators {
			if indicator.Indicates(file) {
				data.Technologies[indicator.GetTechnology()] = append(data.Technologies[indicator.GetTechnology()], file)
				// Same file can't indicate on more than one technology.
				break
			}
		}
	}
	return
}

func createCredentials(serviceDetails auth.ServiceDetails) (auth transport.AuthMethod) {
	var password string
	if serviceDetails.GetApiKey() != "" {
		password = serviceDetails.GetApiKey()
	} else if serviceDetails.GetAccessToken() != "" {
		password = serviceDetails.GetAccessToken()
	} else {
		password = serviceDetails.GetPassword()
	}
	return &http.BasicAuth{Username: serviceDetails.GetUser(), Password: password}
}

func creatRepo() (*services.RepositoryDetails, error) {
	return nil, fmt.Errorf("Not Impelemanted...")
}
