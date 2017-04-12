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

func GitLfsClean(gitpath string, flags *GitLfsCleanFlags) error {
	var err error
	repo := flags.Repo
	if len(repo) <= 0 {
		repo, err = detectRepo(gitpath, flags.ArtDetails.Url)
		if err != nil {
			ex := "Cannot detect Git LFS repository from Git config"
			qu := "Try passing the --repo option manually?"
			return fmt.Errorf("%s: %s. %s", ex, err, qu)
		}
	}
	log.Info("Gathering artifacts in repository", repo, "...")
	regex := flags.Refs
	if !flags.Regexp {
		regex = getRefsRegex(flags.Refs)
	}
	sflags := new(SearchFlags)
	sflags.ArtDetails = flags.ArtDetails
	artfiles, err := getArtFiles(repo, sflags)
	if err != nil {
		return err
	}
	log.Info("Gathering artifacts to preserve from Git repository ...")
	lfsfiles, err := getLfsFiles(gitpath, regex)
	if err != nil {
		return err
	}
	delfiles := findFilesToDelete(artfiles, lfsfiles)
	log.Info("Found", len(lfsfiles), "files to keep, and",
		len(delfiles), "to clean")
	if confirmDelete(delfiles, flags.Quiet) {
		log.Info("Cleaning", len(delfiles), "files from", repo, "...")
		dflags := new(DeleteFlags)
		dflags.ArtDetails = flags.ArtDetails
		dflags.DryRun = flags.DryRun
		err = deleteArtFiles(repo, delfiles, dflags)
		if err != nil {
			return err
		}
	}
	return nil
}

func detectRepo(gitpath, arturl string) (string, error) {
	lfsconf, err := os.Open(path.Join(gitpath, ".lfsconfig"))
	if err != nil {
		return "", err
	}
	conf := gitconfig.New()
	err = gitconfig.NewDecoder(lfsconf).Decode(conf)
	if err != nil {
		return "", err
	}
	url1, err := url.Parse(arturl)
	if err != nil {
		return "", err
	}
	url2, err := url.Parse(conf.Section("lfs").Option("url"))
	if err != nil {
		return "", err
	}
	if url1.Scheme != url2.Scheme || url1.Host != url2.Host {
		ex := "Configured Git LFS URL %q does not match provided URL %q"
		return "", fmt.Errorf(ex, url2.String(), url1.String())
	}
	url1path := path.Clean("/"+url1.Path+"/api/lfs") + "/"
	url2path := path.Clean(url2.Path)
	if strings.HasPrefix(url2path, url1path) {
		return url2path[len(url1path):], nil
	}
	ex := "Configured Git LFS URL %q does not match provided URL %q"
	return "", fmt.Errorf(ex, url2.String(), url1.String())
}

func getRefsRegex(refs string) string {
	replacer := strings.NewReplacer(",", "|", "\\*", ".*")
	return replacer.Replace(regexp.QuoteMeta(refs))
}

func getArtFiles(repo string, flags *SearchFlags) ([]string, error) {
	regexs := "^" + regexp.QuoteMeta(repo) + "/objects/"
	regexs += "([A-Fa-f0-9]{2})/([A-Fa-f0-9]{2})/([A-Fa-f0-9]{64})$"
	regex := regexp.MustCompile(regexs)
	spec := utils.CreateSpec(repo, "", "", "", true, false, false, false)
	results, err := Search(spec, flags)
	if err != nil {
		return nil, err
	}
	res := make([]string, 0, len(results))
	for _, val := range results {
		path := regex.FindStringSubmatch(val.Path)
		if path == nil || path[1]+path[2] != path[3][0:4] {
			continue
		}
		res = append(res, path[3])
	}
	return res, nil
}

func deleteArtFiles(repo string, files []string, flags *DeleteFlags) error {
	fs := make([]utils.File, 0, len(files))
	for _, val := range files {
		var file utils.File
		path := val[0:2] + "/" + val[2:4] + "/" + val
		file.Pattern = repo + "/objects/" + path
		fs = append(fs, file)
	}
	spec := new(utils.SpecFiles)
	spec.Files = fs
	return Delete(spec, flags)
}

func findFilesToDelete(existing []string, lfs map[string]struct{}) []string {
	results := make([]string, 0, len(existing))
	for _, file := range existing {
		if _, inlfs := lfs[file]; !inlfs {
			results = append(results, file)
		}
	}
	return results
}

func getLfsFiles(path, refmatch string) (map[string]struct{}, error) {
	regex := "^oid sha256:[A-Fa-f0-9]{64}$"
	results := make(map[string]struct{}, 0)
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	refs, err := repo.References()
	if err != nil {
		return nil, err
	}
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Type() != plumbing.HashReference {
			return nil
		}
		match, err := regexp.MatchString(refmatch, ref.Name().String())
		if err != nil {
			return err
		}
		if !match {
			return nil
		}
		commit, err := repo.Commit(ref.Hash())
		if err != nil {
			return err
		}
		files, err := commit.Files()
		if err != nil {
			return err
		}
		err = files.ForEach(func(file *object.File) error {
			reader, err := file.Reader()
			if err != nil {
				return err
			}
			defer reader.Close()
			buf := make([]byte, 8)
			size, err := reader.Read(buf)
			if err != nil {
				return err
			}
			if string(buf[:size]) != "version " {
				return nil
			}
			lines, err := file.Lines()
			if err != nil {
				return err
			}
			for _, line := range lines {
				if !strings.HasPrefix(line, "oid ") {
					continue
				}
				if len(line) != 75 {
					return nil
				}
				matched, err := regexp.MatchString(regex, line)
				if err != nil {
					return err
				}
				if !matched {
					return nil
				}
				result := line[11:]
				results[result] = struct{}{}
				break
			}
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func confirmDelete(files []string, quiet bool) bool {
	if len(files) < 1 {
		return false
	}
	if quiet {
		return true
	}
	for _, v := range files {
		fmt.Println("  " + v)
	}
	var confirm string
	fmt.Print("Are you sure you want to delete the above files? (y/n): ")
	fmt.Scanln(&confirm)
	return cliutils.ConfirmAnswer(confirm)
}
