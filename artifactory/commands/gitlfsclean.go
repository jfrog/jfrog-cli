package commands

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	gitconfig "gopkg.in/src-d/go-git.v4/plumbing/format/config"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type GitLfsCleanFlags struct {
	ArtDetails *config.ArtifactoryDetails
	Refs       string
	Regexp     bool
	Repo       string
	Quiet      bool
	DryRun     bool
}

func (flags *GitLfsCleanFlags) GetArtifactoryDetails() *config.ArtifactoryDetails {
	return flags.ArtDetails
}

func (flags *GitLfsCleanFlags) IsDryRun() bool {
	return flags.DryRun
}

func GitLfsClean(gitPath string, flags *GitLfsCleanFlags) error {
	var err error
	repo := flags.Repo
	if len(repo) <= 0 {
		repo, err = detectRepo(gitPath, flags.ArtDetails.Url)
		if err != nil {
			return err
		}
	}
	log.Info("Gathering artifacts in repository", repo, "...")
	regex := flags.Refs
	if !flags.Regexp {
		regex = getRefsRegex(flags.Refs)
	}
	artifactoryLfsFiles, err := searchLfsFilesInArtifactory(repo, flags)
	if err != nil {
		return err
	}
	log.Info("Gathering artifacts to preserve from Git references matching the pattern", flags.Refs, "...")
	gitLfsFiles, err := getLfsFilesFromGit(gitPath, regex)
	if err != nil {
		return err
	}
	filesToDelete := findFilesToDelete(artifactoryLfsFiles, gitLfsFiles)
	log.Info("Found", len(gitLfsFiles), "files to keep, and", len(filesToDelete), "to clean")
	if confirmDelete(filesToDelete, flags.Quiet) {
		err = deleteLfsFilesFromArtifactory(repo, filesToDelete, flags)
		if err != nil {
			return err
		}
	}
	return nil
}

func detectRepo(gitPath, rtUrl string) (string, error) {
	repo, err := extractRepo(gitPath, ".lfsconfig", rtUrl)
	if err == nil {
		return repo, nil
	}
	errMsg1 := fmt.Sprintln("Cannot detect Git LFS repository from .lfsconfig: %s", err)
	repo, err = extractRepo(gitPath, ".git/config", rtUrl)
	if err == nil {
		return repo, nil
	}
	errMsg2 := fmt.Sprintln("Cannot detect Git LFS repository from .git/config: %s", err)
	suggestedSolution := "You may want to try passing the --repo option manually"
	return "", fmt.Errorf("%s%s%s", errMsg1, errMsg2, suggestedSolution)
}

func extractRepo(gitPath, configFile, rtUrl string) (string, error) {
	lfsConf, err := os.Open(path.Join(gitPath, configFile))
	if err != nil {
		return "", err
	}
	defer lfsConf.Close()
	conf := gitconfig.New()
	err = gitconfig.NewDecoder(lfsConf).Decode(conf)
	if err != nil {
		return "", err
	}
	artifactoryConfiguredUrl, err := url.Parse(rtUrl)
	if err != nil {
		return "", err
	}
	lfsUrl, err := url.Parse(conf.Section("lfs").Option("url"))
	if err != nil {
		return "", err
	}
	if artifactoryConfiguredUrl.Scheme != lfsUrl.Scheme || artifactoryConfiguredUrl.Host != lfsUrl.Host {
		return "", fmt.Errorf("Configured Git LFS URL %q does not match provided URL %q", lfsUrl.String(), artifactoryConfiguredUrl.String())
	}
	artifactoryConfiguredUrlPath := path.Clean("/" + artifactoryConfiguredUrl.Path + "/api/lfs") + "/"
	lfsUrlPath := path.Clean(lfsUrl.Path)
	if strings.HasPrefix(lfsUrlPath, artifactoryConfiguredUrlPath) {
		return lfsUrlPath[len(artifactoryConfiguredUrlPath):], nil
	}
	return "", fmt.Errorf("Configured Git LFS URL %q does not match provided URL %q", lfsUrl.String(), artifactoryConfiguredUrl.String())
}

func getRefsRegex(refs string) string {
	replacer := strings.NewReplacer(",", "|", "\\*", ".*")
	return replacer.Replace(regexp.QuoteMeta(refs))
}

func searchLfsFilesInArtifactory(repo string, flags utils.CommonFlag) ([]utils.AqlSearchResultItem, error) {
	err := utils.PreCommandSetup(flags)
	if err != nil {
		return nil, err
	}
	spec := utils.CreateSpec(repo, "", "", "", true, false, false, false)
	return utils.AqlSearchDefaultReturnFields(spec.Get(0), flags)
}

func deleteLfsFilesFromArtifactory(repo string, files []utils.AqlSearchResultItem, flags utils.CommonFlag) error {
	log.Info("Cleaning", len(files), "files from", repo, "...")
	return DeleteFiles(files, flags)
}

func findFilesToDelete(artifactoryLfsFiles []utils.AqlSearchResultItem, gitLfsFiles map[string]struct{}) []utils.AqlSearchResultItem {
	results := make([]utils.AqlSearchResultItem, 0, len(artifactoryLfsFiles))
	for _, file := range artifactoryLfsFiles {
		if _, keepFile := gitLfsFiles[file.Name]; !keepFile {
			results = append(results, file)
		}
	}
	return results
}

func getLfsFilesFromGit(path, refMatch string) (map[string]struct{}, error) {
	// a hash set of sha2 sums, to make lookup faster later
	results := make(map[string]struct{}, 0)
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	log.Debug("Opened Git repo at", path, "for reading")
	refs, err := repo.References()
	if err != nil {
		return nil, err
	}
	// look for every Git LFS pointer file that exists in any ref (branch,
	// remote branch, tag, etc.) who's name matches the regex refMatch
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		// go-git recognizes three types of refs: regular hash refs,
		// symbolic refs (e.g. HEAD), and invalid refs. We only care
		// about the first type here.
		if ref.Type() != plumbing.HashReference {
			return nil
		}
		log.Debug("Checking ref", ref.Name().String())
		match, err := regexp.MatchString(refMatch, ref.Name().String())
		if err != nil || !match {
			return err
		}
		commit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			return err
		}
		files, err := commit.Files()
		if err != nil {
			return err
		}
		err = files.ForEach(func(file *object.File) error {
			return collectLfsFileFromGit(results, file)
		})
		return err
	})
	return results, err
}

func collectLfsFileFromGit(results map[string]struct{}, file *object.File) error {
	// A Git LFS pointer is a small file containing a sha2. Any file bigger
	// than a kilobyte is extremely unlikely to be such a pointer.
	if file.Size > 1024 {
		return nil
	}
	lines, err := file.Lines()
	if err != nil {
		return err
	}
	// the line containing the sha2 we're looking for will match this regex
	regex := "^oid sha256:[[:alnum:]]{64}$"
	for _, line := range lines {
		if !strings.HasPrefix(line, "oid ") {
			continue
		}
		match, err := regexp.MatchString(regex, line)
		if err != nil || !match {
			return err
		}
		result := line[strings.Index(line, ":") + 1:]
		log.Debug("Found file", result)
		results[result] = struct{}{}
		break
	}
	return nil
}

func confirmDelete(files []utils.AqlSearchResultItem, quiet bool) bool {
	if len(files) < 1 {
		return false
	}
	if quiet {
		return true
	}
	for _, v := range files {
		fmt.Println("  " + v.Name)
	}
	var confirm string
	fmt.Print("Are you sure you want to delete the above files? (y/n): ")
	fmt.Scanln(&confirm)
	return cliutils.ConfirmAnswer(confirm)
}
