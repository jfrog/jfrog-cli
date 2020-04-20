package dependencies

import (
	"encoding/xml"
	"fmt"
	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli/artifactory/utils/dotnet"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io/ioutil"
	"path/filepath"
	"strings"
)

var packagesFilePath = "packages.config"

// Register packages.config extractor
func init() {
	register(&packagesExtractor{})
}

// packages.config dependency extractor
type packagesExtractor struct {
	allDependencies  map[string]*buildinfo.Dependency
	childrenMap      map[string][]string
	rootDependencies []string
}

func (extractor *packagesExtractor) IsCompatible(projectName, projectRoot string) (bool, error) {
	packagesConfigPath := filepath.Join(projectRoot, packagesFilePath)
	exists, err := fileutils.IsFileExists(packagesConfigPath, false)
	if exists {
		log.Debug("Found", packagesConfigPath, "file for project:", projectName)
		return true, err
	}
	return false, err
}

func (extractor *packagesExtractor) DirectDependencies() ([]string, error) {
	return getDirectDependencies(extractor.allDependencies, extractor.childrenMap), nil
}

func (extractor *packagesExtractor) AllDependencies() (map[string]*buildinfo.Dependency, error) {
	return extractor.allDependencies, nil
}

func (extractor *packagesExtractor) ChildrenMap() (map[string][]string, error) {
	return extractor.childrenMap, nil
}

// Create new packages.config extractor
func (extractor *packagesExtractor) new(projectName, projectRoot string) (Extractor, error) {
	newExtractor := &packagesExtractor{allDependencies: map[string]*buildinfo.Dependency{}, childrenMap: map[string][]string{}}
	packagesConfig, err := newExtractor.loadPackagesConfig(projectRoot)
	if err != nil {
		return nil, err
	}

	globalPackagesCache, err := newExtractor.getGlobalPackagesCache()
	if err != nil {
		return nil, err
	}

	err = newExtractor.extract(packagesConfig, globalPackagesCache)
	return newExtractor, err
}

func (extractor *packagesExtractor) extract(packagesConfig *packagesConfig, globalPackagesCache string) error {
	for _, nuget := range packagesConfig.XmlPackages {
		id := strings.ToLower(nuget.Id)
		nPackage := &nugetPackage{id: id, version: nuget.Version, dependencies: map[string]bool{}}
		// First lets check if the original version exists within the file system:
		pack, err := createNugetPackage(globalPackagesCache, nuget, nPackage)
		if err != nil {
			return err
		}
		if pack == nil {
			// If doesn't exists lets build the array of alternative versions.
			alternativeVersions := createAlternativeVersionForms(nuget.Version)
			// Now lets do a loop to run over the alternative possibilities
			for i := 0; i < len(alternativeVersions); i++ {
				nPackage.version = alternativeVersions[i]
				pack, err = createNugetPackage(globalPackagesCache, nuget, nPackage)
				if err != nil {
					return err
				}
				if pack != nil {
					break
				}
			}
		}
		if pack != nil {
			extractor.allDependencies[id] = pack.dependency
			extractor.childrenMap[id] = pack.getDependencies()
		} else {
			log.Warn(fmt.Sprintf("The following NuGet package %s with version %s was not found in the NuGet cache %s and therefore not"+
				" added to the dependecy tree. This might be because the package already exists in a different NuGet cache,"+
				" possibly the SDK cache. Removing the package from this cache may resolve the issue", nuget.Id, nuget.Version, globalPackagesCache))
		}
	}
	return nil
}

// NuGet allows the version will be with missing or unnecessary zeros
// This method will return a list of the possible alternative versions
// "1.0" --> []string{"1.0.0.0", "1.0.0", "1"}
// "1" --> []string{"1.0.0.0", "1.0.0", "1.0"}
// "1.2" --> []string{"1.2.0.0", "1.2.0"}
// "1.22.33" --> []string{"1.22.33.0"}
// "1.22.33.44" --> []string{}
// "1.0.2" --> []string{"1.0.2.0"}
func createAlternativeVersionForms(originalVersion string) []string {
	versionSlice := strings.Split(originalVersion, ".")
	versionSliceSize := len(versionSlice)
	for i := 4; i > versionSliceSize; i-- {
		versionSlice = append(versionSlice, "0")
	}

	var alternativeVersions []string

	for i := 4; i > 0; i-- {
		version := strings.Join(versionSlice[:i], ".")
		if version != originalVersion {
			alternativeVersions = append(alternativeVersions, version)
		}
		if !strings.HasSuffix(version, ".0") {
			return alternativeVersions
		}
	}
	return alternativeVersions
}

func (extractor *packagesExtractor) loadPackagesConfig(rootPath string) (*packagesConfig, error) {
	packagesFilePath := filepath.Join(rootPath, packagesFilePath)
	content, err := ioutil.ReadFile(packagesFilePath)
	if err != nil {
		return nil, err
	}

	config := &packagesConfig{}
	err = xml.Unmarshal(content, config)
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	return config, nil
}

type dfsHelper struct {
	visited  bool
	notRoot  bool
	circular bool
}

func getDirectDependencies(allDependencies map[string]*buildinfo.Dependency, childrenMap map[string][]string) []string {
	helper := map[string]*dfsHelper{}
	for id := range allDependencies {
		helper[id] = &dfsHelper{}
	}

	for id := range allDependencies {
		if helper[id].visited {
			continue
		}
		searchRootDependencies(helper, id, allDependencies, childrenMap, map[string]bool{id: true})
	}
	var rootDependencies []string
	for id, nodeData := range helper {
		if !nodeData.notRoot || nodeData.circular {
			rootDependencies = append(rootDependencies, id)
		}
	}

	return rootDependencies
}

func searchRootDependencies(dfsHelper map[string]*dfsHelper, currentId string, allDependencies map[string]*buildinfo.Dependency, childrenMap map[string][]string, traversePath map[string]bool) {
	if dfsHelper[currentId].visited {
		return
	}
	for _, next := range childrenMap[currentId] {
		if _, ok := allDependencies[next]; !ok {
			// No such dependency
			continue
		}
		if traversePath[next] {
			for circular := range traversePath {
				dfsHelper[circular].circular = true
			}
			continue
		}

		// Not root dependency
		dfsHelper[next].notRoot = true
		traversePath[next] = true
		searchRootDependencies(dfsHelper, next, allDependencies, childrenMap, traversePath)
		delete(traversePath, next)
	}
	dfsHelper[currentId].visited = true
}

func createNugetPackage(packagesPath string, nuget xmlPackage, nPackage *nugetPackage) (*nugetPackage, error) {
	nupkgPath := filepath.Join(packagesPath, nPackage.id, nPackage.version, strings.Join([]string{nPackage.id, nPackage.version, "nupkg"}, "."))

	exists, err := fileutils.IsFileExists(nupkgPath, false)

	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	fileDetails, err := fileutils.GetFileDetails(nupkgPath)
	if err != nil {
		return nil, err
	}
	nPackage.dependency = &buildinfo.Dependency{Id: nuget.Id + ":" + nuget.Version, Checksum: &buildinfo.Checksum{Sha1: fileDetails.Checksum.Sha1, Md5: fileDetails.Checksum.Md5}}

	// Nuspec file that holds the metadata for the package.
	nuspecPath := filepath.Join(packagesPath, nPackage.id, nPackage.version, strings.Join([]string{nPackage.id, "nuspec"}, "."))
	nuspecContent, err := ioutil.ReadFile(nuspecPath)
	if err != nil {
		return nil, errorutils.CheckError(err)
	}

	nuspec := &nuspec{}
	err = xml.Unmarshal(nuspecContent, nuspec)
	if err != nil {
		pack := nPackage.id + ":" + nPackage.version
		log.Warn("Package:", pack, "couldn't be parsed due to:", err.Error(), ". Skipping the package dependency.")
		return nPackage, nil
	}

	for _, dependency := range nuspec.Metadata.Dependencies.Dependencies {
		nPackage.dependencies[strings.ToLower(dependency.Id)] = true
	}

	for _, group := range nuspec.Metadata.Dependencies.Groups {
		for _, dependency := range group.Dependencies {
			nPackage.dependencies[strings.ToLower(dependency.Id)] = true
		}
	}

	return nPackage, nil
}

type nugetPackage struct {
	id           string
	version      string
	dependency   *buildinfo.Dependency
	dependencies map[string]bool // Set of dependencies
}

func (nugetPackage *nugetPackage) getDependencies() []string {
	var dependencies []string
	for key := range nugetPackage.dependencies {
		dependencies = append(dependencies, key)
	}

	return dependencies
}

func (extractor *packagesExtractor) getGlobalPackagesCache() (string, error) {
	localsCmd, err := dotnet.NewToolchainCmd(dotnet.Nuget)
	if err != nil {
		return "", err
	}
	//nuget locals global-packages -list
	localsCmd.Command = []string{"locals", "global-packages"}
	localsCmd.CommandFlags = []string{"-list"}

	output, err := gofrogcmd.RunCmdOutput(localsCmd)
	if err != nil {
		return "", err
	}

	globalPackagesPath := strings.TrimSpace(strings.TrimPrefix(output, "global-packages:"))
	exists, err := fileutils.IsDirExists(globalPackagesPath, false)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", errorutils.CheckError(fmt.Errorf("Could not find global packages path at: %s", globalPackagesPath))
	}
	return globalPackagesPath, nil
}

// packages.config xml objects for unmarshalling
type packagesConfig struct {
	XMLName     xml.Name     `xml:"packages"`
	XmlPackages []xmlPackage `xml:"package"`
}

type xmlPackage struct {
	Id      string `xml:"id,attr"`
	Version string `xml:"version,attr"`
}

type nuspec struct {
	XMLName  xml.Name `xml:"package"`
	Metadata metadata `xml:"metadata"`
}

type metadata struct {
	Dependencies xmlDependencies `xml:"dependencies"`
}

type xmlDependencies struct {
	Groups       []group      `xml:"group"`
	Dependencies []xmlPackage `xml:"dependency"`
}

type group struct {
	TargetFramework string       `xml:"targetFramework,attr"`
	Dependencies    []xmlPackage `xml:"dependency"`
}
