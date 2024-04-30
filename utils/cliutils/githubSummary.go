package cliutils

import (
	"encoding/json"
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/utils"
	artifactoryUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"os"
	"path"
	"path/filepath"
)

type Result struct {
	SourcePath string `json:"sourcePath"`
	TargetPath string `json:"targetPath"`
	RtUrl      string `json:"rtUrl"`
}

type ResultsWrapper struct {
	Results []Result `json:"results"`
}

type GitHubActionSummary struct {
	homeDirPath string                     // Directory path for the GitHubActionSummary data
	rawDataFile string                     // File which contains all the results of the commands
	uploadTree  *artifactoryUtils.FileTree // Upload a tree object to generate markdown
}

const (
	githubActionsEnv = "GITHUB_ACTIONS"
)

func GenerateGitHubActionSummary(result *utils.Result) (err error) {
	if os.Getenv(githubActionsEnv) != "true" {
		return
	}
	// Initiate the GitHubActionSummary, will check for previous runs and aggregate results if needed.
	gh, err := createNewGithubSummary()
	if err != nil {
		return fmt.Errorf("failed while initiating Github job summaries: %w", err)
	}
	// Appends the current command results to the result file.
	log.Debug("append results to file")
	if err = gh.AppendResult(result); err != nil {
		return fmt.Errorf("failed while appending results: %s", err)
	}
	// Generate upload tree
	log.Debug("generate uploaded files tree")
	if err = gh.generateUploadedFilesTree(); err != nil {
		return fmt.Errorf("failed while creating file tree: %w", err)
	}
	// Generate the whole markdown
	log.Debug("generating markdown")
	return gh.generateMarkdown()
}

func (gh *GitHubActionSummary) generateUploadedFilesTree() (err error) {
	object, _, err := gh.loadAndMarshalResultsFile()
	if err != nil {
		return
	}
	gh.uploadTree = artifactoryUtils.NewFileTree()
	for _, b := range object.Results {
		gh.uploadTree.AddFile(b.TargetPath)
	}
	return
}

func (gh *GitHubActionSummary) getDataFilePath() string {
	return path.Join(gh.homeDirPath, gh.rawDataFile)
}

func (gh *GitHubActionSummary) AppendResult(result *utils.Result) error {
	// Read all the current command result files.
	var readContent []Result
	if result != nil && result.Reader() != nil {
		for _, file := range result.Reader().GetFilesPaths() {
			// Read source file
			sourceBytes, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			// Unmarshal source file content
			var sourceWrapper ResultsWrapper
			err = json.Unmarshal(sourceBytes, &sourceWrapper)
			if err != nil {
				return err
			}
			readContent = append(readContent, sourceWrapper.Results...)
		}
	}
	targetWrapper, targetBytes, err := gh.loadAndMarshalResultsFile()
	// Append source results to target results
	targetWrapper.Results = append(targetWrapper.Results, readContent...)
	// Marshal target results
	targetBytes, err = json.MarshalIndent(targetWrapper, "", "  ")
	if err != nil {
		return err
	}
	// Write target results to target file
	return os.WriteFile(gh.getDataFilePath(), targetBytes, 0644)
}

func (gh *GitHubActionSummary) loadAndMarshalResultsFile() (targetWrapper ResultsWrapper, targetBytes []byte, err error) {
	// Load target file
	targetBytes, err = os.ReadFile(gh.getDataFilePath())
	if err != nil && !os.IsNotExist(err) {
		log.Warn("data file not found ", gh.getDataFilePath())
		return ResultsWrapper{}, nil, err
	}
	if len(targetBytes) <= 0 {
		log.Warn("empty data file: ", gh.getDataFilePath())
		return
	}
	// Unmarshal target file content, if it exists
	if err = json.Unmarshal(targetBytes, &targetWrapper); err != nil {
		return
	}
	return
}

func (gh *GitHubActionSummary) generateMarkdown() (err error) {
	tempMarkdownPath := path.Join(gh.homeDirPath, "summary.md")
	// Remove the file if it exists
	if err = os.Remove(tempMarkdownPath); err != nil {
		log.Debug("failed to remove old markdown file: ", err)
	}
	file, err := os.OpenFile(tempMarkdownPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer func() {
		err = file.Close()
	}()
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	WriteStringToFile(file, "# 🐸 JFrog CLI Github Action Summary 🐸\n")
	WriteStringToFile(file, "## Uploaded artifacts:\n")
	WriteStringToFile(file, "```\n"+gh.uploadTree.String()+"```")
	return
}

func (gh *GitHubActionSummary) createTempFileIfNeeded(filePath string, content any) (err error) {
	exists, err := fileutils.IsFileExists(filePath, true)
	if err != nil || exists {
		return
	}
	file, err := os.Create(filePath)
	defer func() {
		err = file.Close()
	}()
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	bytes, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to marshal content: %w", err)
	}
	_, err = file.Write(bytes)
	if err != nil {
		return fmt.Errorf("failed to write content: %w", err)
	}
	log.Info("created file:", file.Name())
	return
}

func (gh *GitHubActionSummary) ensureHomeDirExists() error {
	if _, err := os.Stat(gh.homeDirPath); os.IsNotExist(err) {
		err = os.MkdirAll(gh.homeDirPath, 0755)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// Initializes a new GitHubActionSummary
func createNewGithubSummary() (gh *GitHubActionSummary, err error) {
	gh = newGithubActionSummary(gh)
	if err = gh.ensureHomeDirExists(); err != nil {
		return nil, err
	}
	if err = gh.createTempFileIfNeeded(gh.getDataFilePath(), ResultsWrapper{Results: []Result{}}); err != nil {
		return nil, err
	}
	return
}

func newGithubActionSummary(gh *GitHubActionSummary) *GitHubActionSummary {
	homedir := GetHomeDirByOs()
	log.Info("home is is:", homedir)
	gh = &GitHubActionSummary{
		homeDirPath: homedir,
		rawDataFile: "data.json",
	}
	return gh
}

func WriteStringToFile(file *os.File, str string) {
	_, err := file.WriteString(str)
	if err != nil {
		log.Error(fmt.Errorf("failed to write string to file: %w", err))
	}
}

func GetHomeDirByOs() string {
	switch osString := os.Getenv("RUNNER_OS"); osString {
	case "Windows":
		return filepath.Join(os.Getenv("USERPROFILE"), ".jfrog", "jfrog-github-summary")
	case "Linux", "macOS":
		return filepath.Join(os.Getenv("HOME"), ".jfrog", "jfrog-github-summary")
	default:
		return ""
	}
}
