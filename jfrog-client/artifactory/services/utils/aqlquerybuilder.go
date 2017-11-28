package utils

import (
	"strings"
	"path/filepath"
	"fmt"
	"strconv"
)

// Returns an AQL body string to search file in Artifactory according the the specified arguments requirements.
func createAqlBodyForSpec(params *ArtifactoryCommonParams) (string, error) {
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
	propsQueryPart, err := buildPropsQueryPart(params.Props)
	if err != nil {
		return "", err
	}
	itemTypeQuery := buildItemTypeQueryPart(itemType)
	nePath := buildNePathPart(pathPairsSize == 0 || includeRoot)
	excludeQuery := buildExcludeQueryPart(params.ExcludePatterns, pathPairsSize == 0 || params.Recursive, params.Recursive)

	json := fmt.Sprintf(`{"repo": "%s",%s"$or": [`, repo, propsQueryPart+itemTypeQuery+nePath+excludeQuery)
	if pathPairsSize == 0 {
		json += buildInnerQueryPart(".", searchPattern)
	} else {
		for i := 0; i < pathPairsSize; i++ {
			json += buildInnerQueryPart(pathFilePairs[i].path, pathFilePairs[i].file)
			if i+1 < pathPairsSize {
				json += ","
			}
		}
	}
	json += "]}"
	return json, nil
}

func createAqlQueryForBuild(buildName, buildNumber string) string {
	buildQueryPart :=
		`items.find({` +
			`"$and" : [` +
			`{"artifact.module.build.name": {"$eq": "%s"}},` +
			`{"artifact.module.build.number": {"$eq": "%s"}}` +
			`]})%s`
	return fmt.Sprintf(buildQueryPart, buildName, buildNumber, buildIncludeQueryPart([]string{"name", "repo", "path", "actual_sha1"}))
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

func buildPropsQueryPart(props string) (string, error) {
	if props == "" {
		return "", nil
	}
	properties, err := ParseProperties(props, JoinCommas)
	if err != nil {
		return "", err
	}
	query := ""
	for _, v := range properties.Properties {
		query += buildKeyValQueryPart(v.Key, v.Value) + `,`
	}
	return query, nil
}

func buildKeyValQueryPart(key string, value string) string {
	return fmt.Sprintf(`"@%s": {"$match" : "%s"}`, key, value)
}

func buildItemTypeQueryPart(itemType string) string {
	if itemType != "" {
		return fmt.Sprintf(`"type": {"$eq": "%s"},`, itemType)
	}
	return ""
}

func buildNePathPart(includeRoot bool) string {
	if !includeRoot {
		return `"path": {"$ne": "."},`
	}
	return ""
}

func buildInnerQueryPart(path, name string) string {
	innerQueryPattern := `{"$and":` +
							`[{` +
								`"path": {"$match": "%s"},` +
								`"name": {"$match": "%s"}` +
							`}]}`
	return fmt.Sprintf(innerQueryPattern, path, name)
}

func buildExcludeQueryPart(excludePatterns []string, useLocalPath, recursive bool) string {
	if excludePatterns == nil {
		return ""
	}
	excludeQuery := ""
	var excludePairs []PathFilePair
	for _, excludePattern := range excludePatterns {
		excludePairs = append(excludePairs, createPathFilePairs(prepareSearchPattern(excludePattern, false), recursive)...)
	}

	for _, excludePair := range excludePairs {
		excludePath := excludePair.path
		if !useLocalPath && excludePath == "." {
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
	searchPattern = searchPattern[:len(searchPattern)-1]
	searchPattern = strings.Replace(searchPattern, "(", "", -1)
	searchPattern = strings.Replace(searchPattern, ")", "", -1)

	index := strings.Index(searchPattern, "/")
	searchPattern = searchPattern[index+1:]

	index = strings.LastIndex(searchPattern, "/")
	lastSlashPath := searchPattern
	path := "."
	if index != -1 {
		lastSlashPath = searchPattern[index+1:]
		path = searchPattern[:index]
	}

	pairs := []PathFilePair{{path: path, file: lastSlashPath}}
	for i := 0; i < len(lastSlashPath); i++ {
		if string(lastSlashPath[i]) == "*" {
			pairs = append(pairs, PathFilePair{path: filepath.Join(path, lastSlashPath[:i+1]), file: lastSlashPath[i:]})
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
		if i+1 < size {
			options = append(options, sections[i]+"*/")
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

func getQueryReturnFields(specFile *ArtifactoryCommonParams) []string {
	returnFields := []string{"name", "repo", "path", "actual_md5", "actual_sha1", "size", "type"}
	if specIncludesSortOrLimit(specFile) {
		return appendMissingFields(specFile.SortBy, returnFields)
	}
	return append(returnFields, "property")
}

func specIncludesSortOrLimit(specFile *ArtifactoryCommonParams) bool {
	return len(specFile.SortBy) > 0 || specFile.Limit > 0
}

func appendMissingFields(fields []string, defaultFields []string) []string {
	for _, field := range fields {
		if !stringIsInSlice(field, defaultFields) {
			defaultFields = append(defaultFields, field)
		}
	}
	return defaultFields
}

func stringIsInSlice(string string, strings []string) bool {
	for _, v := range strings {
		if v == string {
			return true
		}
	}
	return false
}

func prepareFieldsForQuery(fields []string) []string {
	for i, val := range fields {
		fields[i] = `"` + val + `"`
	}
	return fields
}

func buildQueryFromSpecFile(specFile *ArtifactoryCommonParams) string {
	aqlBody := specFile.Aql.ItemsFind
	query := fmt.Sprintf(`items.find(%s)%s`, aqlBody, buildIncludeQueryPart(getQueryReturnFields(specFile)))
	query = appendSortQueryPart(specFile, query)
	query = appendOffsetQueryPart(specFile, query)
	return appendLimitQueryPart(specFile, query)
}

func appendOffsetQueryPart(specFile *ArtifactoryCommonParams, query string) string {
	if specFile.Offset > 0 {
		query = fmt.Sprintf(`%s.offset(%s)`, query, strconv.Itoa(specFile.Offset))
	}
	return query
}

func appendLimitQueryPart(specFile *ArtifactoryCommonParams, query string) string {
	if specFile.Limit > 0 {
		query = fmt.Sprintf(`%s.limit(%s)`, query, strconv.Itoa(specFile.Limit))
	}
	return query
}

func appendSortQueryPart(specFile *ArtifactoryCommonParams, query string) string {
	if len(specFile.SortBy) > 0 {
		query = fmt.Sprintf(`%s.sort({%s})`, query, buildSortQueryPart(specFile.SortBy, specFile.SortOrder))
	}
	return query
}

func buildSortQueryPart(sortFields []string, sortOrder string) string {
	if sortOrder == "" {
		sortOrder = "asc"
	}
	return fmt.Sprintf(`"$%s":[%s]`, sortOrder, strings.Join(prepareFieldsForQuery(sortFields), `,`))
}

func createPropsQuery(aqlBody, propKey, propVal string) string {
	propKeyValQueryPart := buildKeyValQueryPart(propKey, propVal)
	propsQuery :=
		`items.find({` +
			`"$and" :[%s,{%s}]` +
		`})%s`
	return fmt.Sprintf(propsQuery, aqlBody, propKeyValQueryPart, buildIncludeQueryPart([]string {"name", "repo", "path", "actual_sha1", "property"}))
}

func buildIncludeQueryPart(fieldsToInclude []string) string {
	fieldsToInclude = prepareFieldsForQuery(fieldsToInclude)
	return fmt.Sprintf(`.include(%s)`, strings.Join(fieldsToInclude, `,`))
}