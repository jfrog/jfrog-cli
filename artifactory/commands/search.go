package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"encoding/json"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
)

type SearchResult struct {
	Path string `json:"path,omitempty"`
}

func Search(search string, flags *SearchFlags) error {
	utils.PreCommandSetup(flags)
	returnFields := []string{"\"name\"", "\"repo\"", "\"path\""}
	resultItems, err := utils.AqlSearch(search, flags, returnFields)
	if err != nil {
	    return err
	}
	result, _ := json.Marshal(aqlResultToSearchResult(resultItems))
	fmt.Println(string(cliutils.IndentJson(result)))
	return nil
}

func aqlResultToSearchResult(aqlResult []utils.AqlSearchResultItem) (result []SearchResult) {
	result = make([]SearchResult, len(aqlResult))
	for i, v := range aqlResult {
		tempResult := new(SearchResult)
		if v.Path != "." {
			tempResult.Path = v.Repo + "/" + v.Path + "/" + v.Name;
		} else {
			tempResult.Path = v.Repo + "/" + v.Name;
		}
		result[i] = *tempResult
	}
	return
}

type SearchFlags struct {
	ArtDetails *config.ArtifactoryDetails
	Props      string
	Recursive  bool
	DryRun     bool
}

func (flags *SearchFlags) GetArtifactoryDetails() *config.ArtifactoryDetails {
	return flags.ArtDetails
}

func (flags *SearchFlags) IsRecursive() bool {
	return flags.Recursive
}

func (flags *SearchFlags) GetProps() string {
	return flags.Props
}

func (flags *SearchFlags) IsDryRun() bool {
	return flags.DryRun
}