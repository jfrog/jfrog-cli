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
	"strings"
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
	CurrentStepCount        int `json:"CurrentStepCount"`        // The current step of the workflow
	LastJFrogCliCommandStep int `json:"LastJFrogCliCommandStep"` // The last step that uses JFrog CLI
}

type GitHubActionSummary struct {
	dirPath     string                     // Directory path for the GitHubActionSummary data
	rawDataFile string                     // File which contains all the results of the commands
	uploadTree  *artifactoryUtils.FileTree // Upload tree object to generate markdown
	runtimeInfo *runtimeInfo               // Information needed to determine the current step and the last JFrog CLI command step
}

type Workflow struct {
	Name string `yaml:"name"`
	Jobs map[string]struct {
		Steps []map[string]interface{} `yaml:"steps"`
	} `yaml:"jobs"`
}

const (
	// TODO change this when stop developing on self hosted
	homeDir = "/home/runner/work/_temp/jfrog-github-summary"
	//homeDir = "/Users/eyalde/IdeaProjects/githubRunner/_work/_temp/jfrog-github-summary"
)

func GenerateGitHubActionSummary(result *utils.Result) (err error) {
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		// TODO change to to return nothing
	}
	// Initiate the GitHubActionSummary, will check for previous runs and manage the runtime info.
	gh, err := initGithubActionSummary()
	if err != nil {
		return fmt.Errorf("failed while initiating Github job summaries: %w", err)
	}
	// Appends the current command results to the results file.
	err = gh.AppendResult(result)
	if err != nil {
		return fmt.Errorf("failed while appending results: %s", err)
	}
	// On the final step, generate the file tree and markdown file.
	if gh.isLastWorkflowStep() {
		if err = gh.generateFileTree(); err != nil {
			return fmt.Errorf("failed while creating file tree: %w", err)
		}
		err = gh.generateMarkdown()
	}
	return
}

func (gh *GitHubActionSummary) generateFileTree() (err error) {
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

	file, err := os.OpenFile(githubMarkdownPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer func() {
		err = file.Close()
	}()
	if err != nil {
		return
	}
	// TODO handle errors better
	_, err = file.WriteString("# 🐸 JFrog CLI Github Action Summary 🐸\n")
	_, err = file.WriteString("## Uploaded artifacts:\n")
	_, err = file.WriteString("```\n" + gh.uploadTree.String() + "```")
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
	//gh.runtimeInfo.MarkdownPath = currentStepId
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
	// TODO for some reasons in cloud we need to subtract 2.
	log.Debug("compare steps: last step: ", gh.runtimeInfo.LastJFrogCliCommandStep, "current step:", currentStepInt)
	return gh.runtimeInfo.LastJFrogCliCommandStep == currentStepInt
}

func (gh *GitHubActionSummary) calculateWorkflowSteps() (rt *runtimeInfo, err error) {
	executedWorkFlow := findCurrentlyExecuteWorkflowFile()
	content, err := os.ReadFile(path.Join(".github/workflows/", executedWorkFlow))
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

	lastStepAppearance := 0
	totalSteps := 0
	for _, job := range wf.Jobs {
		for i, step := range job.Steps {
			for key, v := range step {
				if key == "uses" || key == "run" {
					if str, ok := v.(string); ok {
						if strings.Contains(str, "jf") {
							lastStepAppearance = i
						}
					}
				}
			}
			totalSteps++
		}
	}

	log.Debug("last JFrog CLI command step: ", lastStepAppearance, "out of ", totalSteps)
	currentCount := os.Getenv("GITHUB_ACTION")

	return &runtimeInfo{
		CurrentStepCount:        extractNumber(currentCount),
		LastJFrogCliCommandStep: lastStepAppearance,
	}, err
}

func initGithubActionSummary() (gh *GitHubActionSummary, err error) {
	gh, err = tryLoadPreviousRuntimeInfo()
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

// Loads previous steps information if exists
func tryLoadPreviousRuntimeInfo() (gh *GitHubActionSummary, err error) {
	gh = &GitHubActionSummary{
		dirPath:     homeDir,
		rawDataFile: "data.json",
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
	// Deletes the current markdown steps file, to avoid duplication.
	err = os.Remove(os.Getenv("GITHUB_STEP_SUMMARY"))
	if err != nil {
		log.Warn("failed to delete previous markdown steps:", err)
	}
	return
}

// Initializes a new GitHubActionSummary
func createNewGithubSummary() (gh *GitHubActionSummary, err error) {
	gh = &GitHubActionSummary{
		dirPath:     homeDir,
		rawDataFile: "data.json",
		runtimeInfo: &runtimeInfo{},
	}
	err = gh.createTempFile(gh.getDataFilePath(), ResultsWrapper{Results: []Result{}})
	if err != nil {
		return nil, fmt.Errorf("failed to create data file: %w", err)
	}
	gh.runtimeInfo, err = gh.calculateWorkflowSteps()
	if err != nil {
		return
	}
	err = gh.createTempFile(gh.getRuntimeInfoFilePath(), gh.runtimeInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to create runtime info file: %w", err)
	}
	return
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

func findCurrentlyExecuteWorkflowFile() string {
	files, err := os.ReadDir(".github/workflows")
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return ""
	}
	envWorkflowName := os.Getenv("GITHUB_WORKFLOW")
	//envWorkflowName := "Print Job Summary"
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".yml") || strings.HasSuffix(file.Name(), ".yaml") {
			content, err := os.ReadFile(".github/workflows/" + file.Name())
			if err != nil {
				fmt.Println("Error reading file:", err)
				continue
			}
			var wf Workflow
			err = yaml.Unmarshal(content, &wf)
			if err != nil {
				fmt.Println("Error parsing YAML:", err)
				continue
			}
			if wf.Name == envWorkflowName {
				fmt.Println("Found matching workflow file:", file.Name())
				return file.Name()
			}
		}
	}
	fmt.Println("No matching workflow file found.")
	return ""
}
