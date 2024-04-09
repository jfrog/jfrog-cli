package summary

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/utils"
	artifactoryUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"os"
	"path"
)

type MarkdownGenerator struct {
	file   *os.File
	result *utils.Result
}

func NewGithubMarkdownGenerator(result *utils.Result) (markdownGenerator *MarkdownGenerator, cleanUp func() error, err error) {
	filename := os.Getenv("GITHUB_STEP_SUMMARY")
	if filename == "" {
		wd, _ := os.Getwd()
		filename = path.Join(wd, "github-action-summary.md")
		// TODO change this to return nil, nil, nil
	}
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	cleanUp = func() error {
		return file.Close()
	}
	markdownGenerator = &MarkdownGenerator{file: file, result: result}

	// Handle empty file
	info, err := file.Stat()
	if err != nil {
		return
	}
	if info.Size() == 0 {
		// First time writing to the file, insert JFrog CLI header
		err = markdownGenerator.writeHeader("🐸 JFrog CLI Github Action Summary 🐸" + filename)
		if err != nil {
			return
		}
	}

	return
}

func (m *MarkdownGenerator) WriteGithubJobSummary(operationTitle string) (err error) {

	if m.result.SuccessCount() > 0 {
		tree := artifactoryUtils.NewFileTree()
		var transferDetailsArray []*clientutils.FileTransferDetails
		for transferDetails := new(clientutils.FileTransferDetails); m.result.Reader().NextRecord(transferDetails) == nil; transferDetails = new(clientutils.FileTransferDetails) {
			tree.AddFile(transferDetails.TargetPath)
			transferDetailsArray = append(transferDetailsArray, transferDetails)
		}
		err = m.writeTable(operationTitle+" summary", transferDetailsArray)
		if err != nil {
			return
		}
		_ = m.writeHeader("Files Structure")
		_ = m.writeString(tree.String())
	}

	if m.result.FailCount() > 0 {
		a := "[🚨Error] Failed uploading %d artifacts.\n"
		err = m.writeHeader(fmt.Sprintf(a, m.result.FailCount()))
	}

	return
}

func (m *MarkdownGenerator) writeHeader(header string) error {
	_, err := m.file.WriteString(fmt.Sprintf("## %s\n", header))
	return err
}

func (m *MarkdownGenerator) writeString(content string) error {
	_, err := m.file.WriteString(content)
	return err
}

func (m *MarkdownGenerator) writeTable(header string, details []*clientutils.FileTransferDetails) error {
	// Write the table header
	_, err := m.file.WriteString(fmt.Sprintf("## %s\n", header))
	if err != nil {
		return err
	}

	// Write the table column names
	_, err = m.file.WriteString("| Source Path 📁   | Target Path 🎯  | Artifactory Url 🔗   | Sha256 🔢  |\n| --- | --- | --- |--- |\n")
	if err != nil {
		return err
	}

	// Write the table rows
	for _, detail := range details {
		line := fmt.Sprintf("| %s | %s | %s | %s |\n", detail.SourcePath, detail.TargetPath, detail.RtUrl, detail.Sha256)
		_, err = m.file.WriteString(line)
		if err != nil {
			return err
		}
	}

	return nil
}
