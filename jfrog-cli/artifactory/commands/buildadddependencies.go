package commands

import (
	"errors"
	cliutils "github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils/buildinfo"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils/spec"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/fspatterns"
	specutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	regxp "regexp"
	"strconv"
)

func BuildAddDependencies(dependenciesSpec *spec.SpecFiles, configuration *BuildAddDependenciesConfiguration) (successCount, failCount int, err error) {
	log.Info("Running Build Add Dependencies command...")
	if !configuration.DryRun {
		if err = cliutils.SaveBuildGeneralDetails(configuration.BuildName, configuration.BuildNumber); err != nil {
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
	rootPath, err := fspatterns.GetRootPath(addDepsParams.GetPattern(), addDepsParams.IsRegexp())
	if err != nil {
		return nil, err
	}

	isDir, err := fileutils.IsDir(rootPath)
	if err != nil {
		return nil, err
	}

	if !isDir || fileutils.IsPathSymlink(addDepsParams.GetPattern()) {
		return []string{fspatterns.GetSingleFileToUpload(rootPath, "", false).LocalPath}, nil
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

func saveDependenciesToFileSystem(files map[string]*fileutils.FileDetails, configuration *BuildAddDependenciesConfiguration) error {
	log.Debug("Saving", strconv.Itoa(len(files)), "dependencies.")
	populateFunc := func(partial *buildinfo.Partial) {
		partial.Dependencies = convertFileInfoToDependencies(files)
	}
	return cliutils.SavePartialBuildInfo(configuration.BuildName, configuration.BuildNumber, populateFunc)
}

func convertFileInfoToDependencies(files map[string]*fileutils.FileDetails) []buildinfo.Dependencies {
	buildDependencies := []buildinfo.Dependencies{}
	for filePath, fileInfo := range files {
		dependency := buildinfo.Dependencies{Checksum: &buildinfo.Checksum{}}
		dependency.Md5 = fileInfo.Checksum.Md5
		dependency.Sha1 = fileInfo.Checksum.Sha1
		filename, _ := fileutils.GetFileAndDirFromPath(filePath)
		dependency.Id = filename
		buildDependencies = append(buildDependencies, dependency)
	}
	return buildDependencies
}

type BuildAddDependenciesConfiguration struct {
	BuildName   string
	BuildNumber string
	DryRun      bool
}
