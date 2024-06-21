package tools

import "regexp"

func Regexp(str, substr string, n int) [][]string {
	return regexp.MustCompile(substr).FindAllStringSubmatch(str, n)
}
