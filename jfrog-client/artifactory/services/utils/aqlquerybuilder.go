package utils

import (
	"strings"
	"path/filepath"
	"fmt"
)

// Returns an AQL body string to search file in Artifactory according the the specified arguments requirements.
func createAqlBodyForItem(params *ArtifactoryCommonParams) (string, error) {
	var itemType string
	if params.IncludeDirs {
		itemType = "any"
	}
	searchPattern := prepareSearchPattern(params.Pattern, true)
	repoIndex := strings.Index(searchPattern, "/")

	repo := searchPattern[:repoIndex]
	searchPattern = searchPattern[repoIndex+1:]

	pathFilePairs := createPathFilePairs(searchPattern, params.Recursive)
	includeRoot := strings.LastIndex(searchPattern, "/") < 0
	pathPairsSize := len(pathFilePairs)
	propsQuery, err := buildPropsQuery(params.Props)
	if err != nil {
		return "", err
	}
	itemTypeQuery := buildItemTypeQuery(itemType)
	nePath := buildNePath(pathPairsSize == 0 || includeRoot)
	excludeQuery := createExcludeQuery(params.ExcludePatterns, pathPairsSize == 0 || params.Recursive)

	json := `{"repo": "` + repo + `",` + propsQuery + itemTypeQuery + nePath + excludeQuery + `"$or": [`
	if pathPairsSize == 0 {
		json += buildInnerQuery(".", searchPattern)
	} else {
		for i := 0; i < pathPairsSize; i++ {
			json += buildInnerQuery(pathFilePairs[i].path, pathFilePairs[i].file)
			if i+1 < pathPairsSize {
				json += ","
			}
		}
	}
	json += "]}"
	return json, nil
}

func createAqlQueryForBuild(buildName, buildNumber string) string {
	return "items.find(" +
			"{\"$and\": [" +
				"{\"artifact.module.build.name\": {\"$eq\": \"" + buildName + "\"}}," +
				"{\"artifact.module.build.number\": {\"$eq\": \"" + buildNumber + "\"}}" +
			"]}).include(\"name\",\"repo\",\"path\",\"actual_sha1\")"
}

func prepareSearchPattern(pattern string, repositoryExists bool) string {
	if repositoryExists && !strings.Contains(pattern, "/") {
		pattern += "/"
	}
	if strings.HasSuffix(pattern, "/") {
		pattern += "*"
	}

	// Remove parenthesis
	pattern = strings.Replace(pattern, "(", "", -1)
	pattern = strings.Replace(pattern, ")", "", -1)
	return pattern
}

func buildPropsQuery(props string) (string, error) {
	if props == "" {
		return "", nil
	}
	propList := strings.Split(props, ";")
	query := ""
	for _, prop := range propList {
		key, value, err := SplitProp(prop)
		if err != nil {
			return "", err
		}
		query += "\"@" + key + "\": {\"$match\" : \"" + value + "\"},"
	}
	return query, nil
}

func buildItemTypeQuery(itemType string) string {
	if itemType != "" {
		return `"type": {"$eq": "` + itemType + `"},`
	}
	return ""
}

func buildNePath(includeRoot bool) string {
	if !includeRoot {
		return `"path": {"$ne": "."},`
	}
	return ""
}

func buildInnerQuery(path, name string) string {
	innerQueryPattern := `{"$and":` +
							`[{` +
								`"path": {"$match": "%s"},` +
								`"name": {"$match": "%s"}` +
							`}]}`
	return fmt.Sprintf(innerQueryPattern, path, name)
}

func createExcludeQuery(excludePatterns []string, useLocalPath bool) string {
	if excludePatterns == nil {
		return ""
	}
	excludeQuery := ""
	var excludePairs []PathFilePair
	for _, excludePattern := range excludePatterns {
		excludePairs = append(excludePairs, createPathFilePairs(prepareSearchPattern(excludePattern, false), useLocalPath)...)
	}

	for _, excludePair := range excludePairs {
		excludePath := excludePair.path
		if excludePath == "." {
			excludePath = "*"
		}
		excludeQuery += fmt.Sprintf(`"$or": [{"path": {"$nmatch": "%s"}, "name": {"$nmatch": "%s"}}],`, excludePath, excludePair.file)
	}
	return excludeQuery
}

// We need to translate the provided download pattern to an AQL query.
// In Artifactory, for each artifact the name and path of the artifact are saved separately including folders.
// We therefore need to build an AQL query that covers all possible folders the provided
// pattern can include.
// For example, the pattern a/*b*c*/ can include the two following folders:
// a/b/c, a/bc/, a/x/y/z/b/c/
// To achieve that, this function parses the pattern by splitting it by its * characters.
// The end result is a list of PathFilePair structs.
// Each struct represent a possible path and folder name pair to be included in AQL query with an "or" relationship.
func createPathFolderPairs(searchPattern string) []PathFilePair {
	// Remove parenthesis
	searchPattern = searchPattern[:len(searchPattern) - 1]
	searchPattern = strings.Replace(searchPattern, "(", "", -1)
	searchPattern = strings.Replace(searchPattern, ")", "", -1)

	index := strings.Index(searchPattern, "/")
	searchPattern = searchPattern[index + 1:]

	index = strings.LastIndex(searchPattern, "/")
	lastSlashPath := searchPattern
	path := "."
	if index != -1 {
		lastSlashPath = searchPattern[index + 1:]
		path = searchPattern[:index]
	}

	pairs := []PathFilePair{{path:path, file:lastSlashPath}}
	for i := 0; i < len(lastSlashPath); i++ {
		if string(lastSlashPath[i]) == "*" {
			pairs = append(pairs, PathFilePair{path:filepath.Join(path, lastSlashPath[:i + 1]), file:lastSlashPath[i:]})
		}
	}
	return pairs
}

// We need to translate the provided download pattern to an AQL query.
// In Artifactory, for each artifact the name and path of the artifact are saved separately.
// We therefore need to build an AQL query that covers all possible paths and names the provided
// pattern can include.
// For example, the pattern a/* can include the two following file:
// a/file1.tgz and also a/b/file2.tgz
// To achieve that, this function parses the pattern by splitting it by its * characters.
// The end result is a list of PathFilePair structs.
// Each struct represent a possible path and file name pair to be included in AQL query with an "or" relationship.
func createPathFilePairs(pattern string, recursive bool) []PathFilePair {
	var defaultPath string
	if recursive {
		defaultPath = "*"
	} else {
		defaultPath = "."
	}

	pairs := []PathFilePair{}
	if pattern == "*" {
		pairs = append(pairs, PathFilePair{defaultPath, "*"})
		return pairs
	}

	slashIndex := strings.LastIndex(pattern, "/")
	var path string
	var name string
	if slashIndex < 0 {
		pairs = append(pairs, PathFilePair{".", pattern})
		path = ""
		name = pattern
	} else {
		path = pattern[:slashIndex]
		name = pattern[slashIndex+1:]
		pairs = append(pairs, PathFilePair{path, name})
	}
	if !recursive {
		return pairs
	}
	if name == "*" {
		path += "/*"
		pairs = append(pairs, PathFilePair{path, "*"})
		return pairs
	}
	pattern = name

	sections := strings.Split(pattern, "*")
	size := len(sections)
	for i := 0; i < size; i++ {
		options := []string{}
		if i + 1 < size {
			options = append(options, sections[i] + "*/")
		}
		for _, option := range options {
			str := ""
			for j := 0; j < size; j++ {
				if j > 0 {
					str += "*"
				}
				if j == i {
					str += option
				} else {
					str += sections[j]
				}
			}
			split := strings.Split(str, "/")
			filePath := split[0]
			fileName := split[1]
			if fileName == "" {
				fileName = "*"
			}
			if path != "" {
				if !strings.HasSuffix(path, "/") {
					path += "/"
				}
			}
			pairs = append(pairs, PathFilePair{path + filePath, fileName})
		}
	}
	return pairs
}

type PathFilePair struct {
	path string
	file string
}