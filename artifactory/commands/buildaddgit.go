package commands

import (
	"os"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"strings"
	"bufio"
	"path/filepath"
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
)

func BuildAddGit(buildName, buildNumber, dotGitPath string) (err error) {
	if err = utils.SaveBuildGeneralDetails(buildName, buildNumber); err != nil {
		return
	}
	if dotGitPath == "" {
		dotGitPath, err = os.Getwd()
		if err != nil {
			return
		}
	}
	gitManager := NewGitManager(dotGitPath)
	err = gitManager.ReadGitConfig()
	if err != nil {
		return
	}

	populateFunc := func(tempWrapper *utils.ArtifactBuildInfoWrapper) {
		tempWrapper.Vcs = &utils.Vcs{
			VcsUrl: gitManager.GetUrl() + ".git",
			VcsRevision: gitManager.GetRevision(),
		}
	}
	err = utils.SavePartialBuildInfo(buildName, buildNumber, populateFunc)
	return
}

type gitManager struct {
	path     string
	err      error
	revision string
	url      string
}

func NewGitManager(path string) *gitManager {
	dotGitPath := filepath.Join(path, ".git")
	return &gitManager{path:dotGitPath}
}

func (m *gitManager) ReadGitConfig() error {
	if m.path == "" {
		return cliutils.CheckError(errors.New(".git path must be defined."))
	}
	m.readRevision()
	m.readUrl()
	return m.err
}

func (m *gitManager) GetUrl() string {
	return m.url
}

func (m *gitManager) GetRevision() string {
	return m.revision
}

func (m *gitManager) readUrl() {
	if m.err != nil {
		return
	}
	dotGitPath := filepath.Join(m.path, "config")
	file, err := os.Open(dotGitPath)
	if cliutils.CheckError(err) != nil {
		m.err = err
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var IsNextLineUrl bool
	var originUrl string
	for scanner.Scan() {
		if IsNextLineUrl {
			text := scanner.Text()
			strings.HasPrefix(text, "url")
			originUrl = strings.TrimSpace(strings.SplitAfter(text, "=")[1])
			break
		}
		if scanner.Text() == "[remote \"origin\"]" {
			IsNextLineUrl = true
		}
	}
	if err := scanner.Err(); err != nil {
		cliutils.CheckError(err)
		m.err = err
		return
	}
	m.url = originUrl
}

func (m *gitManager) getWorkingBranchFilePath() (refUrl string, err error) {
	dotGitPath := filepath.Join(m.path, "HEAD")
	file, e := os.Open(dotGitPath)
	if cliutils.CheckError(e) != nil {
		err = e
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, "ref") {
			refUrl = strings.TrimSpace(strings.SplitAfter(text, ":")[1])
			break
		}
	}
	if err = scanner.Err(); err != nil {
		cliutils.CheckError(err)
		return
	}
	return
}

func (m *gitManager) readRevision() {
	if m.err != nil {
		return
	}
	ref, err := m.getWorkingBranchFilePath()
	if err != nil {
		m.err = err
		return
	}
	dotGitPath := filepath.Join(m.path, ref)
	file, err := os.Open(dotGitPath)
	if err != nil {
		m.err = err
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var revision string
	for scanner.Scan() {
		text := scanner.Text()
		revision = strings.TrimSpace(text)
		break
	}
	if err := scanner.Err(); err != nil {
		cliutils.CheckError(err)
		m.err = err
		return
	}

	m.revision = revision
	return
}
