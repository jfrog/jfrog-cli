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
	runtimeInfo *runtimeInfo
}

func GenerateGitHubActionSummary(result *utils.Result, command string) (err error) {
	// TODO remove this after
	//if os.Getenv("GITHUB_ACTIONS") != "true" {
	//	// Do nothing if not running in GitHub Actions
	//	log.Warn("Not running in GitHub Actions, skipping GitHub Action summary generation")
	//	return
	//}

	gh, err := initGithubActionSummary()
	if err != nil {
		return
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

	// Set current step markdown as the final markdown

	return
}

type runtimeInfo struct {
	PreviousStepId string
}

func initGithubActionSummary() (gh *GitHubActionSummary, err error) {

	// First create a directory to store all the results. across the entire workflow.

	// TODO replace this when moving from self hosted.
	//dirPath := "/home/runner/work/_temp/jfrog-github-summary/"
	dirPath := "/Users/eyalde/IdeaProjects/githubRunner/_work/_temp/jfrog-github-summary"

	err = fileutils.CreateDirIfNotExist(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create dir %s: %w", dirPath, err)
	}
	log.Debug("successfully created dir ", dirPath)

	gh = &GitHubActionSummary{
		dirPath:     dirPath,
		rawDataFile: "text.txt",
		treeFile:    "tree.txt",
		uploadTree:  nil,
	}

	err = gh.loadRuntimeInfo()
	if err != nil {
		return nil, err
	}
	return
}

func (gh *GitHubActionSummary) getRuntimeInfoFilePath() string {
	return path.Join(gh.dirPath, "runtime-info.json")
}

// Loads previous steps information
func (gh *GitHubActionSummary) loadRuntimeInfo() error {
	runtimeFilePath := gh.getRuntimeInfoFilePath()
	// Check if the file exists
	_, err := os.Stat(runtimeFilePath)
	if os.IsNotExist(err) {
		// If the file does not exist, create it
		log.Debug("file doesn't exists, creating it...")
		file, err := os.Create(runtimeFilePath)
		if err != nil {
			return err
		}
		defer file.Close()
	} else if err != nil {
		// If there was an error checking the file, return it
		return err
	}

	file, err := os.Open(runtimeFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read the file content
	content, err := os.ReadFile(runtimeFilePath)
	if err != nil {
		return err
	}

	if len(content) > 0 {
		// Unmarshal the JSON content into the runtimeInfo object
		err = json.Unmarshal(content, gh.runtimeInfo)
		if err != nil {
			return err
		}
	} else {
		gh.runtimeInfo = &runtimeInfo{}
	}

	err = os.Remove(gh.runtimeInfo.PreviousStepId)
	if err != nil {
		log.Warn("failed trying to remove previous step id ", gh.runtimeInfo.PreviousStepId)
	}

	return nil
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

	//finalMarkdownPath := path.Join(gh.dirPath, "github-action-summary.md")
	finalMarkdownPath := path.Join(os.Getenv("GITHUB_STEP_SUMMARY"))

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

	_ = gh.updateRuntimeInfo()
}

// Updates the runtime info file with the current step id
func (gh *GitHubActionSummary) updateRuntimeInfo() error {
	rt := runtimeInfo{PreviousStepId: os.Getenv("GITHUB_STEP_SUMMARY")}
	// Marshal the runtimeInfo object into JSON
	content, err := json.Marshal(rt)
	if err != nil {
		return err
	}
	// Write the JSON content to the runtime-info.json file
	err = os.WriteFile(gh.getRuntimeInfoFilePath(), content, 0644)
	if err != nil {
		return err
	}
	return nil
}
