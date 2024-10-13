package pattern

import (
	"regexp"
	"strings"
)

func Match(str string, patterns []string) bool {
	for _, v := range patterns {
		if wildCardMatch(str, v) {
			return true
		}
	}
	return false
}

func wildCardBuilder(pattern string) string {
	var result strings.Builder
	for i, literal := range strings.Split(pattern, "*") {

		// Replace * with .*
		if i > 0 {
			result.WriteString(".*")
		}

		result.WriteString(regexp.QuoteMeta(literal))
	}
	return result.String()
}

func wildCardMatch(str string, pattern string) bool {
	result, _ := regexp.MatchString(wildCardBuilder(pattern), str)
	return result
}
