package buildinfo

import (
	"errors"
	"github.com/jfrog/jfrog-cli-go/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/artifactory/services/fspatterns"
	specutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	regxp "regexp"
	"strconv"
)

func AddDependencies(dependenciesSpec *spec.SpecFiles, configuration *AddDependenciesConfiguration) (successCount, failCount int, err error) {
	log.Info("Running Build Add Dependencies command...")
	if !configuration.DryRun {
		if err = utils.SaveBuildGeneralDetails(configuration.BuildName, configuration.BuildNumber); err != nil {
			return 0, 0, err
		}
	}

	dependenciesPaths, errorOccurred := collectDependenciesBySpec(dependenciesSpec)
	dependenciesDetails, errorOccurred, failures := collectDependenciesChecksums(dependenciesPaths, errorOccurred)
	if !configuration.DryRun {
		err = saveDependenciesToFileSystem(dependenciesDetails, configuration)
		if err != nil {
			errorOccurred = true
			log.Error(err)
			// mark all as failures and clean the succeeded
			failures += len(dependenciesDetails)
			dependenciesDetails = make(map[string]*fileutils.FileDetails)
		}
	}
	if errorOccurred {
		err = errors.New("Build Add Dependencies command finished with errors. Please review the logs.")
	}

	return len(dependenciesDetails), failures, err
}

func collectDependenciesChecksums(dependenciesPaths map[string]string, errorOccurred bool) (map[string]*fileutils.FileDetails, bool, int) {
	failures := 0
	dependenciesDetails := make(map[string]*fileutils.FileDetails)
	for _, dependencyPath := range dependenciesPaths {
		var details *fileutils.FileDetails
		var err error
		if fileutils.IsPathSymlink(dependencyPath) {
			log.Info("Adding symlink dependency:", dependencyPath)
			details, err = fspatterns.CreateSymlinkFileDetails()
		} else {
			log.Info("Adding dependency:", dependencyPath)
			details, err = fileutils.GetFileDetails(dependencyPath)
		}
		if err != nil {
			errorOccurred = true
			log.Error(err)
			failures++
			continue
		}
		dependenciesDetails[dependencyPath] = details
	}
	return dependenciesDetails, errorOccurred, failures
}

func collectDependenciesBySpec(dependenciesSpec *spec.SpecFiles) (map[string]string, bool) {
	errorOccurred := false
	dependenciesPaths := make(map[string]string)
	for _, specFile := range dependenciesSpec.Files {
		params, err := prepareArtifactoryParams(specFile)
		if err != nil {
			errorOccurred = true
			log.Error(err)
			continue
		}
		paths, err := getDependenciesBySpecFileParams(params)
		if err != nil {
			errorOccurred = true
			log.Error(err)
			continue
		}
		for _, path := range paths {
			log.Debug("Found matching path:", path)
			dependenciesPaths[path] = path
		}
	}
	return dependenciesPaths, errorOccurred
}

func prepareArtifactoryParams(specFile spec.File) (*specutils.ArtifactoryCommonParams, error) {
	params := specFile.ToArtifactoryCommonParams()
	recursive, err := clientutils.StringToBool(specFile.Recursive, true)
	if err != nil {
		return nil, err
	}

	params.Recursive = recursive
	regexp, err := clientutils.StringToBool(specFile.Regexp, false)
	if err != nil {
		return nil, err
	}

	params.Regexp = regexp
	return params, nil
}

func getDependenciesBySpecFileParams(addDepsParams *specutils.ArtifactoryCommonParams) ([]string, error) {
	addDepsParams.SetPattern(clientutils.ReplaceTildeWithUserHome(addDepsParams.GetPattern()))
	rootPath, err := fspatterns.GetRootPath(addDepsParams.GetPattern(), addDepsParams.IsRegexp(), false)
	if err != nil {
		return nil, err
	}

	isDir, err := fileutils.IsDirExists(rootPath, false)
	if err != nil {
		return nil, err
	}

	if !isDir || fileutils.IsPathSymlink(addDepsParams.GetPattern()) {
		artifact, err := fspatterns.GetSingleFileToUpload(rootPath, "", false, false)
		if err != nil {
			return nil, err
		}
		return []string{artifact.LocalPath}, nil
	}
	return collectPatternMatchingFiles(addDepsParams, rootPath)
}

func collectPatternMatchingFiles(addDepsParams *specutils.ArtifactoryCommonParams, rootPath string) ([]string, error) {
	addDepsParams.SetPattern(clientutils.PrepareLocalPathForUpload(addDepsParams.Pattern, addDepsParams.IsRegexp()))
	excludePathPattern := fspatterns.PrepareExcludePathPattern(addDepsParams)
	patternRegex, err := regxp.Compile(addDepsParams.Pattern)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}

	paths, err := fspatterns.GetPaths(rootPath, addDepsParams.IsRecursive(), addDepsParams.IsIncludeDirs(), true)
	if err != nil {
		return nil, err
	}
	result := []string{}

	for _, path := range paths {
		matches, _, _, err := fspatterns.PrepareAndFilterPaths(path, excludePathPattern, true, false, patternRegex)
		if err != nil {
			log.Error(err)
			continue
		}
		if len(matches) > 0 {
			result = append(result, path)
		}
	}
	return result, nil
}

func saveDependenciesToFileSystem(files map[string]*fileutils.FileDetails, configuration *AddDependenciesConfiguration) error {
	log.Debug("Saving", strconv.Itoa(len(files)), "dependencies.")
	populateFunc := func(partial *buildinfo.Partial) {
		partial.Dependencies = convertFileInfoToDependencies(files)
	}
	return utils.SavePartialBuildInfo(configuration.BuildName, configuration.BuildNumber, populateFunc)
}

func convertFileInfoToDependencies(files map[string]*fileutils.FileDetails) []buildinfo.Dependency {
	var buildDependencies []buildinfo.Dependency
	for filePath, fileInfo := range files {
		dependency := buildinfo.Dependency{Checksum: &buildinfo.Checksum{}}
		dependency.Md5 = fileInfo.Checksum.Md5
		dependency.Sha1 = fileInfo.Checksum.Sha1
		filename, _ := fileutils.GetFileAndDirFromPath(filePath)
		dependency.Id = filename
		buildDependencies = append(buildDependencies, dependency)
	}
	return buildDependencies
}

type AddDependenciesConfiguration struct {
	BuildName   string
	BuildNumber string
	DryRun      bool
}
