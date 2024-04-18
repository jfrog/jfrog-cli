package cliutils

import (
	"encoding/json"
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/utils"
	artifactoryUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"regexp"
	"strconv"
)

type Result struct {
	SourcePath string `json:"sourcePath"`
	TargetPath string `json:"targetPath"`
	RtUrl      string `json:"rtUrl"`
}

type ResultsWrapper struct {
	Results []Result `json:"results"`
}

type runtimeInfo struct {
	CurrentStepCount int `json:"CurrentStepCount"`
	TotalStepCount   int `json:"TotalStepCount"`
}

type GitHubActionSummary struct {
	dirPath     string
	rawDataFile string
	treeFile    string
	uploadTree  *artifactoryUtils.FileTree
	runtimeInfo *runtimeInfo
}

var (
	// TODO change this when stop developing on self hosted
	//homeDir = "/home/runner/work/_temp/jfrog-github-summary"
	homeDir = "/Users/eyalde/IdeaProjects/githubRunner/_work/_temp/jfrog-github-summary"
)

func GenerateGitHubActionSummary(result *utils.Result, command string) (err error) {
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		// TODO change to to return nothing
	}
	gh, err := initGithubActionSummary()
	if err != nil {
		return fmt.Errorf("failed while creating github summary object: %w", err)
	}
	// Append current command results to a temp file.
	err = gh.AppendResult(result, command)
	if err != nil {
		return fmt.Errorf("failed while appending results: %s", err)
	}
	err = gh.generateFileTree()
	if err != nil {
		return fmt.Errorf("failed while creating file tree: %w", err)
	}

	if gh.isLastWorkflowStep() {
		err = gh.generateMarkdown()
	}
	return
}

func (gh *GitHubActionSummary) generateFileTree() (err error) {
	// Create tree
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

func (gh *GitHubActionSummary) getRuntimeInfoFilePath() string {
	return path.Join(gh.dirPath, "runtime-info.json")
}

func (gh *GitHubActionSummary) getDataFilePath() string {
	return path.Join(gh.dirPath, gh.rawDataFile)
}

func (gh *GitHubActionSummary) AppendResult(result *utils.Result, command string) error {
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
	err = os.WriteFile(gh.getDataFilePath(), targetBytes, 0644)
	if err != nil {
		return err
	}

	return nil
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
	githubMarkdownPath := path.Join(os.Getenv("GITHUB_STEP_SUMMARY"))
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		githubMarkdownPath = path.Join(gh.dirPath, "github-action-summary.md")
	}
	log.Debug("writing markdown to: ", githubMarkdownPath)

	file, err := os.OpenFile(githubMarkdownPath, os.O_CREATE|os.O_WRONLY, 0644)
	defer func() {
		err = file.Close()
	}()
	if err != nil {
		return
	}
	// TODO handle errors better
	_, err = file.WriteString("# 🐸 JFrog CLI Github Action Summary 🐸\n step id: " + githubMarkdownPath + "\n")
	_, err = file.WriteString("## Uploaded artifacts:\n")
	_, err = file.WriteString(gh.uploadTree.String())
	return

}

// Updates the runtime info file with the current step id
func (gh *GitHubActionSummary) updateRuntimeInfo() error {
	currentStepId := os.Getenv("GITHUB_STEP_SUMMARY")
	// TODO remove this code block
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		currentStepId = path.Join(gh.dirPath, "github-action-summary.md")
	}
	log.Debug("current step summary path: ", currentStepId)
	//gh.runtimeInfo.PreviousStepId = currentStepId
	// Marshal the runtimeInfo object into JSON
	content, err := json.Marshal(gh.runtimeInfo)
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

func initGithubActionSummary() (gh *GitHubActionSummary, err error) {
	gh, err = tryLoadRuntimeInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to load runtime info: %w", err)
	}
	if gh != nil {
		log.Debug("successfully loaded GitHubActionSummary from previous runs")
		return
	}
	log.Debug("creating new GitHubActionSummary...")
	gh, err = createNewGithubSummary()
	if err != nil {
		return nil, fmt.Errorf("failed to create temp files: %w", err)
	}
	return
}

// Loads previous steps information
func tryLoadRuntimeInfo() (gh *GitHubActionSummary, err error) {
	gh = &GitHubActionSummary{
		dirPath:     homeDir,
		rawDataFile: "data.json",
		treeFile:    "tree.json",
		runtimeInfo: &runtimeInfo{},
	}
	err = fileutils.CreateDirIfNotExist(homeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create dir %s: %w", homeDir, err)
	}

	runtimeFilePath := gh.getRuntimeInfoFilePath()
	// Check previous runs files exists
	exists, err := fileutils.IsFileExists(runtimeFilePath, true)
	if err != nil || !exists {
		return nil, err
	}
	// The Previous file exists, read it and load details
	file, err := os.Open(runtimeFilePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = file.Close()
	}()

	// Read the file content
	content, err := os.ReadFile(runtimeFilePath)
	if err != nil || content != nil && len(content) <= 0 {
		return nil, err
	}
	// Unmarshal the JSON content into the runtimeInfo object
	err = json.Unmarshal(content, gh.runtimeInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal runtime info: %w", err)
	}
	//log.Debug("Deleting previous markdown steps:", gh.runtimeInfo.PreviousStepId)
	//err = os.Remove(gh.runtimeInfo.PreviousStepId)
	//if err != nil {
	//	log.Warn("failed to delete previous markdown steps:", err)
	//}
	return
}

// Initializes a new GitHubActionSummary
func createNewGithubSummary() (gh *GitHubActionSummary, err error) {
	gh = &GitHubActionSummary{
		dirPath:     homeDir,
		rawDataFile: "data.json",
		treeFile:    "tree.json",
		runtimeInfo: &runtimeInfo{},
	}
	err = gh.createTempFile(gh.getDataFilePath(), ResultsWrapper{Results: []Result{}})
	if err != nil {
		return nil, fmt.Errorf("failed to create data file: %w", err)
	}
	rtInfo, err := gh.calculateWorkflowSteps()
	if err != nil {
		return
	}
	err = gh.createTempFile(gh.getRuntimeInfoFilePath(), rtInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to create runtime info file: %w", err)
	}
	return
}

func (gh *GitHubActionSummary) createTempFile(filePath string, content any) (err error) {
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

func (gh *GitHubActionSummary) isLastWorkflowStep() bool {
	currentStepCount := os.Getenv("GITHUB_ACTION")
	log.Info("current step count: ", currentStepCount)
	currentStepInt := extractNumber(currentStepCount)
	log.Debug("compare steps: ", gh.runtimeInfo.TotalStepCount-1, currentStepInt)
	return gh.runtimeInfo.TotalStepCount == currentStepInt
}

type Workflow struct {
	Jobs map[string]struct {
		Steps []map[string]interface{} `yaml:"steps"`
	} `yaml:"jobs"`
}

func (gh *GitHubActionSummary) calculateWorkflowSteps() (rt *runtimeInfo, err error) {
	content, err := os.ReadFile(".github/workflows/jobSummary.yml")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	var wf Workflow
	err = yaml.Unmarshal(content, &wf)
	if err != nil {
		fmt.Println("Error parsing YAML:", err)
		return
	}

	stepCount := 0
	for _, job := range wf.Jobs {
		stepCount += len(job.Steps)
	}

	fmt.Println("Step count:", stepCount)
	currentCount := os.Getenv("GITHUB_ACTION")

	return &runtimeInfo{
		CurrentStepCount: extractNumber(currentCount),
		TotalStepCount:   stepCount,
	}, err
}

func extractNumber(s string) int {
	re := regexp.MustCompile("[0-9]+")
	match := re.FindString(s)
	if match == "" {
		return -1
	}
	number, _ := strconv.Atoi(match)
	return number
}
