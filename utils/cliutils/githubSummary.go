package cliutils

import (
	"encoding/json"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"os"
)

const (
	aggregatedResultsPath = "/private/var/folders/t6/sprgv27970x8zw7h165d47vm0000gq/T/githubsummary/text.txt"
)

type Result struct {
	SourcePath string `json:"sourcePath"`
	TargetPath string `json:"targetPath"`
	RtUrl      string `json:"rtUrl"`
}

type ResultsWrapper struct {
	Results []Result `json:"results"`
}

func AppendResults(sourceFile string) error {

	exists, err := fileutils.IsFileExists(aggregatedResultsPath, true)
	if err != nil {
		return err
	}
	if !exists {
		_, err = fileutils.CreateFilePath("/private/var/folders/t6/sprgv27970x8zw7h165d47vm0000gq/T/githubsummary/", "text.txt")
	}

	// Read source file
	sourceBytes, err := os.ReadFile(sourceFile)
	if err != nil {
		return err
	}

	// Unmarshal source file content
	var sourceWrapper ResultsWrapper
	err = json.Unmarshal(sourceBytes, &sourceWrapper)
	if err != nil {
		return err
	}

	// Read target file, if it exists
	targetBytes, err := os.ReadFile(aggregatedResultsPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Unmarshal target file content, if it exists
	var targetWrapper ResultsWrapper
	if len(targetBytes) > 0 {
		err = json.Unmarshal(targetBytes, &targetWrapper)
		if err != nil {
			return err
		}
	}

	// Append source results to target results
	targetWrapper.Results = append(targetWrapper.Results, sourceWrapper.Results...)

	// Marshal target results
	targetBytes, err = json.MarshalIndent(targetWrapper, "", "  ")
	if err != nil {
		return err
	}

	// Write target results to target file
	err = os.WriteFile(aggregatedResultsPath, targetBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}
