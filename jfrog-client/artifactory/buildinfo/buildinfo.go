package buildinfo

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/auth"
	"time"
)

func New() *BuildInfo {
	return &BuildInfo{
		Agent:      &Agent{},
		BuildAgent: &Agent{Name: "GENERIC"},
		Modules:    make([]Module, 0),
		Vcs:        &Vcs{},
	}
}

func (targetBuildInfo *BuildInfo) SetBuildAgentVersion(buildAgentVersion string) {
	targetBuildInfo.BuildAgent.Version = buildAgentVersion
}

func (targetBuildInfo *BuildInfo) SetAgentName(agentName string) {
	targetBuildInfo.Agent.Name = agentName
}

func (targetBuildInfo *BuildInfo) SetAgentVersion(agentVersion string) {
	targetBuildInfo.Agent.Version = agentVersion
}

// Append the modules of the recieved build info to this build info.
// If the two build info instances contain modules with identical names, these modules are merged.
// When merging the modules, the artifacts and dependencies remain unique according to their checksums.
func (targetBuildInfo *BuildInfo) Append(buildInfo *BuildInfo) {
	for _, newModule := range buildInfo.Modules {
		exists := false
		for i, _ := range targetBuildInfo.Modules {
			if newModule.Id == targetBuildInfo.Modules[i].Id {
				mergeModules(&newModule, &targetBuildInfo.Modules[i])
				exists = true
				break
			}
		}
		if !exists {
			targetBuildInfo.Modules = append(targetBuildInfo.Modules, newModule)
		}
	}
}

// Merge the first module into the second module.
func mergeModules(merge *Module, into *Module) {
	mergeArtifacts(&merge.Artifacts, &into.Artifacts)
	mergeDependencies(&merge.Dependencies, &into.Dependencies)
}

func mergeArtifacts(mergeArtifacts *[]Artifact, intoArtifacts *[]Artifact) {
	for _, mergeArtifact := range *mergeArtifacts {
		exists := false
		for _, artifact := range *intoArtifacts {
			if mergeArtifact.Sha1 == artifact.Sha1 {
				exists = true
				break
			}
		}
		if !exists {
			*intoArtifacts = append(*intoArtifacts, mergeArtifact)
		}
	}
}

func mergeDependencies(mergeDependencies *[]Dependency, intoDependencies *[]Dependency) {
	for _, mergeDependency := range *mergeDependencies {
		exists := false
		for _, dependency := range *intoDependencies {
			if mergeDependency.Sha1 == dependency.Sha1 {
				exists = true
				break
			}
		}
		if !exists {
			*intoDependencies = append(*intoDependencies, mergeDependency)
		}
	}
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
	BuildUrl             string   `json:"url,omitempty"`
	*Vcs
}

type Agent struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

type Module struct {
	Properties   interface{}  `json:"properties,omitempty"`
	Id           string       `json:"id,omitempty"`
	Artifacts    []Artifact   `json:"artifacts,omitempty"`
	Dependencies []Dependency `json:"dependencies,omitempty"`
}

type Artifact struct {
	Name string `json:"name,omitempty"`
	*Checksum
}

type Dependency struct {
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
	Artifacts    []Artifact   `json:"Artifacts,omitempty"`
	Dependencies []Dependency `json:"Dependencies,omitempty"`
	Env          Env          `json:"Env,omitempty"`
	Timestamp    int64        `json:"Timestamp,omitempty"`
	*Vcs
	ModuleId     string       `json:"ModuleId,omitempty"`
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

type Configuration struct {
	ArtDetails auth.ArtifactoryDetails
	BuildUrl   string
	DryRun     bool
	EnvInclude string
	EnvExclude string
}

func (config *Configuration) GetArtifactoryDetails() auth.ArtifactoryDetails {
	return config.ArtDetails
}

func (config *Configuration) SetArtifactoryDetails(artDetails auth.ArtifactoryDetails) {
	config.ArtDetails = artDetails
}

func (config *Configuration) IsDryRun() bool {
	return config.DryRun
}
