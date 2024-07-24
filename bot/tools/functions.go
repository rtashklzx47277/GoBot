package tools

import "regexp"

func Regexp(str, substr string) (string, bool) {
	match := regexp.MustCompile(substr).FindAllStringSubmatch(str, 1)

	if len(match) == 0 {
		return "", false
	}

	return match[0][1], true
}

func IsContain(list []string, target string) bool {
	for _, element := range list {
		if element == target {
			return true
		}
	}

	return false
}
