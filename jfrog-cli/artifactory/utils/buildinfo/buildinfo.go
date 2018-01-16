package buildinfo

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/auth"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
	"time"
)

func New() *BuildInfo {
	return &BuildInfo{
		Agent:      &Agent{Name: utils.ClientAgent, Version: utils.GetVersion()},
		BuildAgent: &Agent{Name: "GENERIC", Version: utils.GetVersion()},
		Modules:    make([]Module, 0),
		Vcs:        &Vcs{},
	}
}

func (targetBuildInfo *BuildInfo) Append(buildInfo *BuildInfo) {
	targetBuildInfo.Modules = append(targetBuildInfo.Modules, buildInfo.Modules...)
}

type BuildInfo struct {
	Name                 string   `json:"name,omitempty"`
	Number               string   `json:"number,omitempty"`
	Agent                *Agent   `json:"agent,omitempty"`
	BuildAgent           *Agent   `json:"buildAgent,omitempty"`
	Modules              []Module `json:"modules,omitempty"`
	Started              string   `json:"started,omitempty"`
	Properties           Env      `json:"properties,omitempty"`
	ArtifactoryPrincipal string   `json:"artifactoryPrincipal,omitempty"`
	*Vcs
}

type Agent struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

type Module struct {
	Properties   interface{}    `json:"properties,omitempty"`
	Id           string         `json:"id,omitempty"`
	Artifacts    []Artifacts    `json:"artifacts,omitempty"`
	Dependencies []Dependencies `json:"dependencies,omitempty"`
}

type Artifacts struct {
	Name string `json:"name,omitempty"`
	*Checksum
}

type Dependencies struct {
	Id     string   `json:"id,omitempty"`
	Scopes []string `json:"scopes,omitempty"`
	*Checksum
}

type Checksum struct {
	Sha1 string `json:"sha1,omitempty"`
	Md5  string `json:"md5,omitempty"`
}

type Env map[string]string

type Vcs struct {
	Url      string `json:"vcsUrl,omitempty"`
	Revision string `json:"vcsRevision,omitempty"`
}

type Partials []*Partial

type Partial struct {
	Artifacts    []Artifacts    `json:"Artifacts,omitempty"`
	Dependencies []Dependencies `json:"Dependencies,omitempty"`
	Env          Env            `json:"Env,omitempty"`
	Timestamp    int64          `json:"Timestamp,omitempty"`
	*Vcs
	ModuleId string `json:"ModuleId,omitempty"`
}

func (partials Partials) Len() int {
	return len(partials)
}

func (partials Partials) Less(i, j int) bool {
	return partials[i].Timestamp < partials[j].Timestamp
}

func (partials Partials) Swap(i, j int) {
	partials[i], partials[j] = partials[j], partials[i]
}

type General struct {
	Timestamp time.Time `json:"Timestamp,omitempty"`
}

type Flags struct {
	ArtDetails auth.ArtifactoryDetails
	DryRun     bool
	EnvInclude string
	EnvExclude string
}

func (flags *Flags) GetArtifactoryDetails() auth.ArtifactoryDetails {
	return flags.ArtDetails
}

func (flags *Flags) SetArtifactoryDetails(artDetails auth.ArtifactoryDetails) {
	flags.ArtDetails = artDetails
}

func (flags *Flags) IsDryRun() bool {
	return flags.DryRun
}
