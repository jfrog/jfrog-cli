package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
	"strings"
	"errors"
)

func Delete(deleteSpec *utils.SpecFiles, flags *DeleteFlags) (err error) {
	err = utils.PreCommandSetup(flags)
	if err != nil {
		return
	}
	resultItems, err := GetPathsToDelete(deleteSpec, flags)
	if err != nil {
		return err
	}
	if err = DeleteFiles(resultItems, flags); err != nil {
		return
	}
	log.Info("Deleted", len(resultItems), "items.")
	return
}

func GetPathsToDelete(deleteSpec *utils.SpecFiles, flags *DeleteFlags) (resultItems []utils.AqlSearchResultItem, err error) {
	log.Info("Searching artifacts...")
	for i := 0; i < len(deleteSpec.Files); i++ {
		var isDirectoryDeleteBool bool
		isSimpleDirectoryDeleteBool, e := isSimpleDirectoryDelete(deleteSpec.Get(i))
		if e != nil {
			err = e
			return
		}
		if !isSimpleDirectoryDeleteBool {
			isDirectoryDeleteBool, e = isDirectoryDelete(deleteSpec.Get(i))
			if e != nil {
				err = e
				return
			}
		}
		switch {
		case deleteSpec.Get(i).GetSpecType() == utils.AQL:
			resultItemsTemp, e := utils.AqlSearchBySpec(deleteSpec.Get(i).Aql, flags)
			if e != nil {
				err = e
				return
			}
			resultItems = append(resultItems, resultItemsTemp...)

		case isSimpleDirectoryDeleteBool:
			simplePathItem := utils.AqlSearchResultItem{Path:deleteSpec.Get(i).Pattern}
			resultItems = append(resultItems, []utils.AqlSearchResultItem{simplePathItem}...)

		case isDirectoryDeleteBool:
			tempResultItems, e := utils.AqlSearchDefaultReturnFields(deleteSpec.Get(i).Pattern, true, "", flags)
			if e != nil {
				err = e
				return
			}
			paths, e := getDirsForDeleteFromFilesPaths(deleteSpec.Get(i).Pattern, tempResultItems)
			if e != nil {
				err = e
				return
			}
			resultItems = append(resultItems, paths...)
		default:
			isRecursive, e := cliutils.StringToBool(deleteSpec.Get(i).Recursive, true)
			if e != nil {
				err = e
				return
			}
			tempResultItems, e := utils.AqlSearchDefaultReturnFields(deleteSpec.Get(i).Pattern,
				isRecursive, deleteSpec.Get(i).Props, flags)
			if e != nil {
				err = e
				return
			}
			resultItems = append(resultItems, tempResultItems...)
		}
	}
	utils.LogSearchResults(len(resultItems))
	return
}

// We have simple dir delete when:
//    1) The deleteFile is a dir path, ends with "/"
//    2) The deleteFile doest contains wildcards
//    3) The user hasn't sent any props
//    4) The delete is recursive
func isSimpleDirectoryDelete(deleteFile *utils.Files) (bool, error) {
	isRecursive, err := cliutils.StringToBool(deleteFile.Recursive, true)
	if err != nil {
		return false, err
	}
	return utils.IsSimpleDirectoryPath(deleteFile.Pattern) && isRecursive && deleteFile.Props == "", nil
}

// The diffrence between isSimpleDirectoryDelete to isDirectoryDelete is:
// isDirectoryDelete returns true when the deleteFile path contains wildcatds.
func isDirectoryDelete(deleteFile *utils.Files) (bool, error) {
	isRecursive, err := cliutils.StringToBool(deleteFile.Recursive, true)
	if err != nil {
		return false, err
	}
	return utils.IsDirectoryPath(deleteFile.Pattern) && isRecursive && deleteFile.Props == "", nil
}

func getDirsForDeleteFromFilesPaths(deletePattern string, filesToDelete []utils.AqlSearchResultItem) ([]utils.AqlSearchResultItem, error) {
	paths := make(map[string]bool)
	for _, file := range filesToDelete {
		path, err := utils.WildcardToDirsPath(deletePattern, file.GetFullUrl())
		if err != nil {
			return []utils.AqlSearchResultItem{}, err
		}
		if len(path) > 0 {
			paths[path] = true
		}
	}
	var result []utils.AqlSearchResultItem
	paths = reduceDirResult(paths)
	for k := range paths {
		repo, path, name := ioutils.SplitArtifactPathToRepoPathName(k)
		result = append(result, utils.AqlSearchResultItem{Repo: repo, Path:path, Name:name})
	}
	return result, nil
}

// Remove unnecessary paths.
// For example if we have two paths for delete a/b/c/ and a/b/
// it's enough to delete only a/b/
func reduceDirResult(paths map[string]bool) map[string]bool {
	for k := range paths {
		for k2 := range paths {
			if k != k2 && strings.HasPrefix(k, k2) {
				delete(paths, k);
				continue
			}
		}
	}
	return paths
}

func DeleteFiles(resultItems []utils.AqlSearchResultItem, flags *DeleteFlags) error {
	for _, v := range resultItems {
		fileUrl, err := utils.BuildArtifactoryUrl(flags.ArtDetails.Url, v.GetFullUrl(), make(map[string]string))
		if err != nil {
			return err
		}
		if flags.DryRun {
			log.Info("[Dry run] Deleting:", v.GetFullUrl())
			continue
		}

		log.Info("Deleting:", v.GetFullUrl())
		httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.ArtDetails)
		resp, body, err := ioutils.SendDelete(fileUrl, nil, httpClientsDetails)
		if err != nil {
			return err
		}

		log.Debug("Artifactory response:", resp.Status)
		if resp.StatusCode != 204 {
			return cliutils.CheckError(errors.New(string(body)))
		}
	}
	return nil
}

type DeleteFlags struct {
	ArtDetails *config.ArtifactoryDetails
	DryRun     bool
}

func (flags *DeleteFlags) GetArtifactoryDetails() *config.ArtifactoryDetails {
	return flags.ArtDetails
}

func (flags *DeleteFlags) IsDryRun() bool {
	return flags.DryRun
}