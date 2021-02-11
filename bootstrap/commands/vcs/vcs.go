package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/codegangsta/cli"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/utils"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

const (
	// Basic vcs questions keys
	JfrogUrl      = "jfrogUrl"
	JfrogUsername = "jfrogUsername"
	JfrogPassword = "jfrogPassword"
	VcsUrl        = "vcsUrl"
	VcsUsername   = "vcsUsername"
	VcsPassword   = "vcsPassword"
	VcsBranch     = "vcsBranch"
	LocalDir      = "localDir"

	// Technologies, repositories & build questions keys

	// Output questions keys
)

type vcsData struct {
	ProjectName             string
	LocalDirPath            string
	VcsBranch               string
	BuildCommand            string
	ArtifactoryVirtualRepos map[technology]services.RepositoryDetails
	// A collection of technologies that was found with a list of theirs indications
	Technologies     map[technology][]string
	VcsCredentials   auth.ServiceDetails
	JfrogCredentials auth.ServiceDetails
}

func VcsCmd(c *cli.Context) error {
	// Basic VCS questionnaire (URLs, Credentials, etc'...)
	err := runBasicAuthenticationQuestionnaire()
	if err != nil {
		return err
	}
	// Interactively create Artifactory repository based on the detected technologies and on going user input
	err = runBuildQuestionnaire()
	if err != nil {
		return err
	}
	// Ask user for the desired output (Pipliens.yaml, Jenkins file or Docker Image)

	return err
}

func runBuildQuestionnaire() (err error) {
	// First create repositories for each technology in Artifactory according to user input
	for tech, descriptor := range data.Technologies {
		if coreutils.AskYesNo(fmt.Sprintf("The %q project descriptor was found, would you like %q repositories will be created in your Artifactory?", descriptor, tech)) {
			switch tech {
			case Maven:

			case Gradle:

			case Npm:

			}
		}
	}
	// Ask for working build command

	return nil
}

func runBasicAuthenticationQuestionnaire() (err error) {
	basicAuthenticationQuestionnaire := &utils.InteractiveQuestionnaire{
		MandatoryQuestionsKeys: []string{JfrogUrl, JfrogUsername, JfrogPassword, VcsUrl, VcsUsername, VcsPassword, VcsBranch, LocalDir},
		QuestionsMap:           basicAuthenticationQuestionMap,
	}
	err = basicAuthenticationQuestionnaire.Perform()
	if err != nil {
		return err
	}
	resBytes, err := json.Marshal(basicAuthenticationQuestionnaire.AnswersMap)
	if err != nil {
		return errorutils.CheckError(err)
	}
	if err = ioutil.WriteFile("./", resBytes, 0644); err != nil {
		return errorutils.CheckError(err)
	}
	log.Info(fmt.Sprintf("Basic VCS Authentication config template successfully created at curent directory"))
	return nil
}

func setJfrogCredentials(iq *utils.InteractiveQuestionnaire, ans string) (value string, err error) {
	data.JfrogCredentials.SetUrl(iq.AnswersMap[JfrogUrl].(string))
	data.JfrogCredentials.SetUser(iq.AnswersMap[JfrogUsername].(string))
	data.JfrogCredentials.SetPassword(iq.AnswersMap[JfrogPassword].(string))
	return
}

func setVcsCredentials(iq *utils.InteractiveQuestionnaire, ans string) (value string, err error) {
	data.VcsCredentials.SetUrl(iq.AnswersMap[VcsUrl].(string))
	data.VcsCredentials.SetUser(iq.AnswersMap[VcsUsername].(string))
	data.VcsCredentials.SetPassword(iq.AnswersMap[VcsPassword].(string))
	return
}

func setAndPreformeClone(iq *utils.InteractiveQuestionnaire, ans string) (value string, err error) {
	data.VcsBranch = iq.AnswersMap[VcsBranch].(string)
	data.LocalDirPath = iq.AnswersMap[LocalDir].(string)
	err = cloneProject()
	if err != nil {
		return
	}
	err = detectTechnologies()
	return
}

func cloneProject() (err error) {
	// Create the desired path if necessary
	err = os.MkdirAll(data.LocalDirPath, os.ModePerm)
	if err != nil {
		return err
	}
	cloneOption := &git.CloneOptions{
		URL:           data.VcsCredentials.GetUrl(),
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", data.VcsBranch)),
		Progress:      os.Stdout,
		Auth:          createCredentials(data.VcsCredentials),
		// Enable git submodules clone if there any.
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	}
	setProjectName()
	// Clone the given repository to the given directory from the given branch
	log.Info("git clone project %q from: %q to: %q", data.ProjectName, data.VcsCredentials.GetUrl(), data.LocalDirPath)
	_, err = git.PlainClone(data.LocalDirPath, false, cloneOption)
	return
}

func setProjectName() {
	vcsUrl := data.VcsCredentials.GetUrl()
	// Trim trailing "/" if one exists
	vcsUrl = strings.TrimSuffix(vcsUrl, "/")
	data.VcsCredentials.SetUrl(vcsUrl)
	projectName := vcsUrl[strings.LastIndex(vcsUrl, "/")+1:]
	if strings.Contains(projectName, ".") {
		projectName = vcsUrl[:strings.LastIndex(vcsUrl, "/")]
	}
	data.ProjectName = projectName
}

func detectTechnologies() (err error) {
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

var data = &vcsData{}

var basicAuthenticationQuestionMap = map[string]utils.QuestionInfo{

	JfrogUrl: {
		Msg:          "",
		PromptPrefix: "Enter JFrog Platform URL >",
		AllowVars:    true,
		Writer:       utils.WriteStringAnswer,
		MapKey:       JfrogUrl,
		Callback:     nil,
	},
	JfrogUsername: {
		Msg:          "",
		PromptPrefix: "Enter JFrog admin user name >",
		AllowVars:    true,
		Writer:       utils.WriteStringAnswer,
		MapKey:       JfrogUsername,
		Callback:     nil,
	},
	JfrogPassword: {
		Msg:          "",
		PromptPrefix: "Enter JFrog password >",
		AllowVars:    true,
		Writer:       utils.WriteStringAnswer,
		MapKey:       JfrogPassword,
		Callback:     setJfrogCredentials,
	},
	VcsUrl: {
		Msg:          "",
		PromptPrefix: "Enter VCS URL >",
		AllowVars:    true,
		Writer:       utils.WriteStringAnswer,
		MapKey:       VcsUrl,
		Callback:     nil,
	},
	VcsUsername: {
		Msg:          "",
		PromptPrefix: "Enter VCS admin user name >",
		AllowVars:    true,
		Writer:       utils.WriteStringAnswer,
		MapKey:       VcsUsername,
		Callback:     nil,
	},
	VcsPassword: {
		Msg:          "",
		PromptPrefix: "Enter VCS password >",
		AllowVars:    true,
		Writer:       utils.WriteStringAnswer,
		MapKey:       VcsPassword,
		Callback:     setVcsCredentials,
	},
	VcsBranch: {
		Msg:          "",
		PromptPrefix: "Enter VCS desired branch >",
		AllowVars:    true,
		Writer:       utils.WriteStringAnswer,
		MapKey:       VcsBranch,
		Callback:     setVcsCredentials,
	},
	LocalDir: {
		Options: []prompt.Suggest{
			{Text: "./JFrogVcsBootstrap", Description: "a temp dir that will include a clone of the VCS project."},
		},
		Msg:          "",
		PromptPrefix: "Enter target directory for projet clone >",
		AllowVars:    true,
		Writer:       utils.WriteStringAnswer,
		MapKey:       LocalDir,
		Callback:     setAndPreformeClone,
	},
}
