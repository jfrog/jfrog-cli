package version

import (
	"strings"
	"strconv"
)

// If ver1 == ver2 returns 0
// If ver1 > ver2 returns 1
// If ver1 < ver2 returns -1
func Compare(ver1, ver2 string) int {
	if ver1 == ver2 {
		return 0
	}

	ver1Tokens := strings.Split(ver1, ".")
	ver2Tokens := strings.Split(ver2, ".")

	maxIndex := len(ver1Tokens)
	if len(ver2Tokens) > maxIndex {
		maxIndex = len(ver2Tokens)
	}

	for tokenIndex := 0; tokenIndex < maxIndex; tokenIndex++ {
		ver1Token := "0"
		if len(ver1Tokens) >= tokenIndex+1 {
			ver1Token = strings.TrimSpace(ver1Tokens[tokenIndex])
		}
		ver2Token := "0"
		if len(ver2Tokens) >= tokenIndex+1 {
			ver2Token = strings.TrimSpace(ver2Tokens[tokenIndex])
		}
		compare := compareTokens(ver1Token, ver2Token)
		if compare != 0 {
			return compare
		}
	}

	return 0
}

func compareTokens(ver1Token, ver2Token string) int {
	// Ignoring error because we strip all the non numeric values in advance.
	ver1TokenInt, _ := strconv.Atoi(getFirstNumeral(ver1Token))
	ver2TokenInt, _ := strconv.Atoi(getFirstNumeral(ver2Token))

	switch {
	case ver1TokenInt > ver2TokenInt:
		return 1
	case ver1TokenInt < ver2TokenInt:
		return -1
	default:
		return strings.Compare(ver1Token, ver2Token)
	}
}

func getFirstNumeral(token string) string {
	numeric := ""
	for i := 0; i < len(token); i++ {
		if _, err := strconv.Atoi(token[i:i]); err != nil {
			break
		}
		numeric += token[i:i]
	}
	if len(numeric) == 0 {
		return "0"
	}
	return numeric
}
