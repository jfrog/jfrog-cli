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
	dirPath     string
	rawDataFile string
	treeFile    string
	uploadTree  *artifactoryUtils.FileTree
	githubPath  string
}

func GenerateGitHubActionSummary(result *utils.Result, command string) (err error) {

	if os.Getenv("GITHUB_ACTIONS") != "true" {
		// Do nothing if not running in GitHub Actions
		log.Warn("Not running in GitHub Actions, skipping GitHub Action summary generation")
		return
	}

	gh, err := initGithubActionSummary()
	if err != nil {
		return
	}

	fullGithubPath := path.Join(gh.githubPath, "workflow-summary")
	err = fileutils.CreateDirIfNotExist(fullGithubPath)
	if err != nil {
		return fmt.Errorf("failed to create dir %s: %w", fullGithubPath, err)
	}

	// Append current command results to a temp file.
	err = gh.AppendResult(result, command)

	// Create tree
	object, _, err := gh.loadAndMarshalResultsFile()
	tree := artifactoryUtils.NewFileTree()
	for _, b := range object.Results {
		tree.AddFile(b.TargetPath)
	}

	gh.uploadTree = tree

	// Write markdown to current step
	gh.generateFinalMarkdown()

	// Clear all previous steps markdowns to avoid duplication

	return
}

func initGithubActionSummary() (gh *GitHubActionSummary, err error) {
	githubPath := "/home/runner/work/_temp/jfrog-github-summary"

	exists, err := fileutils.IsDirExists(githubPath, true)
	if err != nil {
		return
	}
	log.Error("dir exists: ", exists)

	err = fileutils.CreateDirIfNotExist(githubPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create dir %s: %w", githubPath, err)
	}
	return &GitHubActionSummary{}, nil
}

func (gh *GitHubActionSummary) getFilePath() string {
	return path.Join(gh.dirPath, gh.rawDataFile)
}

func (gh *GitHubActionSummary) AppendResult(result *utils.Result, command string) error {
	// Create temp file if don't exists
	exists, err := fileutils.IsFileExists(gh.getFilePath(), true)
	if err != nil {
		return err
	}
	if !exists {

		_, err = fileutils.CreateFilePath(gh.dirPath, gh.rawDataFile)
	}
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
	err = os.WriteFile(gh.getFilePath(), targetBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (gh *GitHubActionSummary) loadAndMarshalResultsFile() (targetWrapper ResultsWrapper, targetBytes []byte, err error) {
	// Load target file
	targetBytes, err = os.ReadFile(gh.getFilePath())
	if err != nil && !os.IsNotExist(err) {
		return ResultsWrapper{}, nil, err
	}
	// Unmarshal target file content, if it exists
	if len(targetBytes) > 0 {
		err = json.Unmarshal(targetBytes, &targetWrapper)
		if err != nil {
			return
		}
	}
	return
}

func (gh *GitHubActionSummary) generateFinalMarkdown() {

	wd, _ := os.Getwd()
	finalMarkdownPath := path.Join(wd, "github-action-summary.md")

	// Delete preexisting file
	exists, err := fileutils.IsFileExists(finalMarkdownPath, true)
	if exists {
		err = os.Remove(finalMarkdownPath)
	}

	file, err := os.OpenFile(finalMarkdownPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()
	if err != nil {
		return
	}

	_, _ = file.WriteString("# 🐸 JFrog CLI Github Action Summary 🐸\n ")

	_, _ = file.WriteString("## Uploaded artifacts:\n")
	_, _ = file.WriteString(gh.uploadTree.String())

}
