package tools

import "regexp"

func Regexp(str, substr string, n int) [][]string {
	return regexp.MustCompile(substr).FindAllStringSubmatch(str, n)
}

func IsContain(list []string, target string) bool {
	for _, element := range list {
		if element == target {
			return true
		}
	}

	return false
}
