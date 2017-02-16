package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/httputils"
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
	resultItems, err := getPathsToDeleteInternal(deleteSpec, flags)
	if err != nil {
		return err
	}
	if err = deleteFiles(resultItems, flags); err != nil {
		return
	}
	log.Info("Deleted", len(resultItems), "items.")
	return
}

func GetPathsToDelete(deleteSpec *utils.SpecFiles, flags *DeleteFlags) ([]utils.AqlSearchResultItem, error) {
	if err := utils.PreCommandSetup(flags); err != nil {
		return nil, err
	}
	return getPathsToDeleteInternal(deleteSpec, flags)
}

func getPathsToDeleteInternal(deleteSpec *utils.SpecFiles, flags *DeleteFlags) (resultItems []utils.AqlSearchResultItem, err error) {
	log.Info("Searching artifacts...")
	for i := 0; i < len(deleteSpec.Files); i++ {
		// Search paths using AQL.
		if deleteSpec.Get(i).GetSpecType() == utils.AQL {
			if resultItemsTemp, e := utils.AqlSearchBySpec(deleteSpec.Get(i), flags); e == nil {
				resultItems = append(resultItems, resultItemsTemp...)
				continue
			} else {
				err = e
				return
			}
		}
		// Simple directory delete, no need to search in Artifactory.
		if simpleDir, e := isSimpleDirectoryDelete(deleteSpec.Get(i)); simpleDir && e == nil {
			simplePathItem := utils.AqlSearchResultItem{Path:deleteSpec.Get(i).Pattern}
			resultItems = append(resultItems, []utils.AqlSearchResultItem{simplePathItem}...)
			continue
		} else if e != nil {
			err = e
			return
		}
		// Directory with wildcard pattern, searching with special AQL query.
		if directoryDelete, e := isDirectoryDelete(deleteSpec.Get(i)); directoryDelete && e == nil {
			query, e := utils.BuildAqlFolderSearchQuery(deleteSpec.Get(i).Pattern, utils.GetDefaultQueryReturnFields())
			if e != nil {
				err = e
				return
			}
			tempResultItems, e := utils.AqlSearch(query, flags)
			if e != nil {
				err = e
				return
			}
			paths := reduceDirResult(tempResultItems)
			resultItems = append(resultItems, paths...)
			continue
		} else if e != nil {
			err = e
			return
		}
		// All other use cases, pattern with/without wildcard files.
		tempResultItems, e := utils.AqlSearchDefaultReturnFields(deleteSpec.Get(i), flags)
		if e != nil {
			err = e
			return
		}
		resultItems = append(resultItems, tempResultItems...)
	}
	utils.LogSearchResults(len(resultItems))
	return
}

// We have simple dir delete when:
//    1) The deleteFile is a dir path, ends with "/"
//    2) The deleteFile doest contains wildcards
//    3) The user hasn't sent any props
//    4) The delete is recursive
func isSimpleDirectoryDelete(deleteFile *utils.File) (bool, error) {
	isRecursive, err := cliutils.StringToBool(deleteFile.Recursive, true)
	if err != nil {
		return false, err
	}
	return utils.IsSimpleDirectoryPath(deleteFile.Pattern) && isRecursive && deleteFile.Props == "", nil
}

// The diffrence between isSimpleDirectoryDelete to isDirectoryDelete is:
// isDirectoryDelete returns true when the deleteFile path contains wildcatds.
func isDirectoryDelete(deleteFile *utils.File) (bool, error) {
	isRecursive, err := cliutils.StringToBool(deleteFile.Recursive, true)
	if err != nil {
		return false, err
	}
	return utils.IsDirectoryPath(deleteFile.Pattern) && isRecursive && deleteFile.Props == "", nil
}

// Remove unnecessary paths.
// For example if we have two paths for delete a/b/c/ and a/b/
// it's enough to delete only a/b/
func reduceDirResult(foldersToDelete []utils.AqlSearchResultItem) []utils.AqlSearchResultItem {
	paths := make(map[string]utils.AqlSearchResultItem)
	for _, file := range foldersToDelete {
		if file.Name == "." {
			continue
		}
		paths[file.GetFullUrl()] = file
	}

	for k := range paths {
		for k2 := range paths {
			if k != k2 && strings.HasPrefix(k, k2) {
				delete(paths, k);
				continue
			}
		}
	}
	var result []utils.AqlSearchResultItem
	for _, v := range paths {
		v.Name += "/"
		result = append(result, v)
	}
	return result
}

func DeleteFiles(resultItems []utils.AqlSearchResultItem, flags *DeleteFlags) error {
	if err := utils.PreCommandSetup(flags); err != nil {
		return err
	}
	return deleteFiles(resultItems, flags)
}

func deleteFiles(resultItems []utils.AqlSearchResultItem, flags *DeleteFlags) error {
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
		resp, body, err := httputils.SendDelete(fileUrl, nil, httpClientsDetails)
		if err != nil {
			return err
		}
		if resp.StatusCode != 204 {
			return cliutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + cliutils.IndentJson(body)))
		}

		log.Debug("Artifactory response:", resp.Status)
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