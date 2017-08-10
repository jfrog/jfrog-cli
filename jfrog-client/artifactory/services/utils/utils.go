package utils

func StripThreadId(dependenciesBuildInfo [][]FileInfo) []FileInfo {
	var buildInfo []FileInfo
	for _, v := range dependenciesBuildInfo {
		buildInfo = append(buildInfo, v...)
	}
	return buildInfo
}
