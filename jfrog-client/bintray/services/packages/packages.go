package packages

import (
	"encoding/json"
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/bintray/auth"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/httpclient"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"net/http"
	"path"
	"strings"
)

func NewService(client *httpclient.HttpClient) *PackageService {
	us := &PackageService{client: client}
	return us
}

func NewPackageParams() *Params {
	return &Params{Path: &Path{}}
}

type PackageService struct {
	client         *httpclient.HttpClient
	BintrayDetails auth.BintrayDetails
}

type Path struct {
	Subject string
	Repo    string
	Package string
}

type Params struct {
	*Path
	Desc                   string
	Labels                 string
	Licenses               string
	CustomLicenses         string
	VcsUrl                 string
	WebsiteUrl             string
	IssueTrackerUrl        string
	GithubRepo             string
	GithubReleaseNotesFile string
	PublicDownloadNumbers  bool
	PublicStats            bool
}

func (ps *PackageService) Create(params *Params) error {
	log.Info("Creating package...")
	resp, body, err := ps.doCreatePackage(params)
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

func (ps *PackageService) Update(params *Params) error {
	if ps.BintrayDetails.GetUser() == "" {
		ps.BintrayDetails.SetUser(params.Subject)
	}
	content, err := createPackageContent(params)
	if err != nil {
		return err
	}

	url := ps.BintrayDetails.GetApiUrl() + path.Join("packages", params.Subject, params.Repo, params.Package)

	log.Info("Updating package...")
	httpClientsDetails := ps.BintrayDetails.CreateHttpClientDetails()
	resp, body, err := httputils.SendPatch(url, []byte(content), httpClientsDetails)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Info("Updated package", params.Package+".")
	return nil
}

func (ps *PackageService) Delete(packagePath *Path) error {
	if ps.BintrayDetails.GetUser() == "" {
		ps.BintrayDetails.SetUser(packagePath.Subject)
	}
	url := ps.BintrayDetails.GetApiUrl() + path.Join("packages", packagePath.Subject, packagePath.Repo, packagePath.Package)

	log.Info("Deleting package...")
	httpClientsDetails := ps.BintrayDetails.CreateHttpClientDetails()
	resp, body, err := httputils.SendDelete(url, nil, httpClientsDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Info("Deleted package", packagePath.Package+".")
	return nil
}

func (ps *PackageService) Show(packagePath *Path) error {
	if ps.BintrayDetails.GetUser() == "" {
		ps.BintrayDetails.SetUser(packagePath.Subject)
	}
	url := ps.BintrayDetails.GetApiUrl() + path.Join("packages", packagePath.Subject, packagePath.Repo, packagePath.Package)

	log.Info("Getting package details...")
	httpClientsDetails := ps.BintrayDetails.CreateHttpClientDetails()
	resp, body, _, _ := httputils.SendGet(url, true, httpClientsDetails)
	if resp.StatusCode != http.StatusOK {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Output(clientutils.IndentJson(body))
	return nil
}

func (ps *PackageService) IsPackageExists(packagePath *Path) (bool, error) {
	url := ps.BintrayDetails.GetApiUrl() + path.Join("packages", packagePath.Subject, packagePath.Repo, packagePath.Package)
	httpClientsDetails := ps.BintrayDetails.CreateHttpClientDetails()

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

func (ps *PackageService) doCreatePackage(params *Params) (*http.Response, []byte, error) {
	if ps.BintrayDetails.GetUser() == "" {
		ps.BintrayDetails.SetUser(params.Subject)
	}
	content, err := createPackageContent(params)
	if err != nil {
		return nil, []byte{}, err
	}

	url := ps.BintrayDetails.GetApiUrl() + path.Join("packages", params.Subject, params.Repo)
	httpClientsDetails := ps.BintrayDetails.CreateHttpClientDetails()
	return httputils.SendPost(url, content, httpClientsDetails)
}

func createPackageContent(params *Params) ([]byte, error) {
	labels := []string{}
	if params.Labels != "" {
		labels = strings.Split(params.Labels, ",")
	}
	licenses := []string{}
	if params.Licenses != "" {
		licenses = strings.Split(params.Licenses, ",")
	}
	customLicenses := []string{}
	if params.CustomLicenses != "" {
		customLicenses = strings.Split(params.CustomLicenses, ",")
	}

	Config := contentConfig{
		Name:                   params.Package,
		Desc:                   params.Desc,
		Labels:                 labels,
		Licenses:               licenses,
		CustomLicenses:         customLicenses,
		VcsUrl:                 params.VcsUrl,
		WebsiteUrl:             params.WebsiteUrl,
		IssueTrackerUrl:        params.IssueTrackerUrl,
		GithubRepo:             params.GithubRepo,
		GithubReleaseNotesFile: params.GithubReleaseNotesFile,
		PublicDownloadNumbers:  params.PublicDownloadNumbers,
		PublicStats:            params.PublicStats,
	}
	requestContent, err := json.Marshal(Config)
	if err != nil {
		return nil, errorutils.CheckError(errors.New("Failed to execute request."))
	}
	return requestContent, nil
}

type contentConfig struct {
	Name                   string   `json:"name,omitempty"`
	Desc                   string   `json:"desc,omitempty"`
	Labels                 []string `json:"labels,omitempty"`
	Licenses               []string `json:"licenses,omitempty"`
	CustomLicenses         []string `json:"custom_licenses,omitempty"`
	VcsUrl                 string   `json:"vcs_url,omitempty"`
	WebsiteUrl             string   `json:"website_url,omitempty"`
	IssueTrackerUrl        string   `json:"issue_tracker_url,omitempty"`
	GithubRepo             string   `json:"github_repo,omitempty"`
	GithubReleaseNotesFile string   `json:"github_release_notes_file,omitempty"`
	PublicDownloadNumbers  bool   `json:"public_download_numbers,omitempty"`
	PublicStats            bool   `json:"public_stats,omitempty"`
}
