package versions

import (
	"encoding/json"
	"errors"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray/auth"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/httpclient"
	clientutils "github.com/jfrog/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
	"net/http"
	"path"
)

func NewService(client *httpclient.HttpClient) *VersionService {
	us := &VersionService{client: client}
	return us
}

func NewVersionParams() *Params {
	return &Params{Path: &Path{}}
}

type VersionService struct {
	client         *httpclient.HttpClient
	BintrayDetails auth.BintrayDetails
}

type Path struct {
	Subject string
	Repo    string
	Package string
	Version string
}

type Params struct {
	*Path
	Desc                     string
	VcsTag                   string
	Released                 string
	GithubReleaseNotesFile   string
	GithubUseTagReleaseNotes bool
}

func (vs *VersionService) Create(params *Params) error {
	log.Info("Creating version...")
	resp, body, err := vs.doCreateVersion(params)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}
	log.Debug("Bintray response:", resp.Status)
	log.Output(clientutils.IndentJson(body))
	return nil
}

func (vs *VersionService) Update(params *Params) error {
	if vs.BintrayDetails.GetUser() == "" {
		vs.BintrayDetails.SetUser(params.Subject)
	}

	content, err := createVersionContent(params)
	if err != nil {
		return err
	}

	url := vs.BintrayDetails.GetApiUrl() + path.Join("packages/", params.Subject, params.Repo, params.Package, "versions", params.Version)

	log.Info("Updating version...")
	httpClientsDetails := vs.BintrayDetails.CreateHttpClientDetails()
	resp, body, err := httputils.SendPatch(url, content, httpClientsDetails)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Info("Updated version", params.Version+".")
	return nil
}

func (vs *VersionService) Publish(versionPath *Path) error {
	if vs.BintrayDetails.GetUser() == "" {
		vs.BintrayDetails.SetUser(versionPath.Subject)
	}
	url := vs.BintrayDetails.GetApiUrl() + path.Join("content", versionPath.Subject, versionPath.Repo, versionPath.Package, versionPath.Version, "publish")

	log.Info("Publishing version...")
	httpClientsDetails := vs.BintrayDetails.CreateHttpClientDetails()
	resp, body, err := httputils.SendPost(url, nil, httpClientsDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Output(clientutils.IndentJson(body))
	return nil
}

func (vs *VersionService) Delete(versionPath *Path) error {
	if vs.BintrayDetails.GetUser() == "" {
		vs.BintrayDetails.SetUser(versionPath.Subject)
	}
	url := vs.BintrayDetails.GetApiUrl() + path.Join("packages", versionPath.Subject, versionPath.Repo, versionPath.Package, "versions", versionPath.Version)

	log.Info("Deleting version...")
	httpClientsDetails := vs.BintrayDetails.CreateHttpClientDetails()
	resp, body, err := httputils.SendDelete(url, nil, httpClientsDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Info("Deleted version", versionPath.Version+".")
	return nil
}

func (vs *VersionService) Show(versionPath *Path) error {
	if vs.BintrayDetails.GetUser() == "" {
		vs.BintrayDetails.SetUser(versionPath.Subject)
	}
	if versionPath.Version == "" {
		versionPath.Version = "_latest"
	}

	url := vs.BintrayDetails.GetApiUrl() + path.Join("packages", versionPath.Subject, versionPath.Repo, versionPath.Package, "versions", versionPath.Version)

	log.Info("Getting version details...")
	httpClientsDetails := vs.BintrayDetails.CreateHttpClientDetails()
	resp, body, _, _ := httputils.SendGet(url, true, httpClientsDetails)

	if resp.StatusCode != http.StatusOK {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Output(clientutils.IndentJson(body))
	return nil
}

func (vs *VersionService) IsVersionExists(versionPath *Path) (bool, error) {
	url := vs.BintrayDetails.GetApiUrl() + path.Join("packages", versionPath.Subject, versionPath.Repo, versionPath.Package, "versions", versionPath.Version)
	httpClientsDetails := vs.BintrayDetails.CreateHttpClientDetails()

	resp, _, err := httputils.SendHead(url, httpClientsDetails)
	if err != nil {
		return false, err
	}
	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return false, errorutils.CheckError(errors.New("Bintray response: " + resp.Status))
}

func (vs *VersionService) doCreateVersion(params *Params) (*http.Response, []byte, error) {
	if vs.BintrayDetails.GetUser() == "" {
		vs.BintrayDetails.SetUser(params.Subject)
	}

	content, err := createVersionContent(params)
	if err != nil {
		return nil, []byte{}, err
	}
	url := vs.BintrayDetails.GetApiUrl() + path.Join("packages", params.Subject, params.Repo, params.Package, "versions")
	httpClientsDetails := vs.BintrayDetails.CreateHttpClientDetails()
	return httputils.SendPost(url, content, httpClientsDetails)
}

func createVersionContent(params *Params) ([]byte, error) {
	Config := contentConfig{
		Name:                     params.Version,
		Desc:                     params.Desc,
		VcsTag:                   params.VcsTag,
		Released:                 params.Released,
		GithubReleaseNotesFile:   params.GithubReleaseNotesFile,
		GithubUseTagReleaseNotes: params.GithubUseTagReleaseNotes,
	}
	requestContent, err := json.Marshal(Config)
	if err != nil {
		return nil, errorutils.CheckError(errors.New("Failed to execute request."))
	}
	return requestContent, nil
}

type contentConfig struct {
	Name                     string `json:"name,omitempty"`
	Desc                     string `json:"desc,omitempty"`
	VcsTag                   string `json:"vcs_tag,omitempty"`
	Released                 string `json:"released,omitempty"`
	GithubReleaseNotesFile   string `json:"github_release_notes_file,omitempty"`
	GithubUseTagReleaseNotes bool `json:"github_use_tag_release_notes,omitempty"`
}
