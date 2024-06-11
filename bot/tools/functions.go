package tools

import (
	"regexp"
)

var DefaultImage = "https://imgur.com/2wAkxNb.png"

func IsContain(list []string, target string) bool {
	for _, element := range list {
		if element == target {
			return true
		}
	}

	return false
}

func Append(list []string, items ...string) []string {
	for _, item := range items {
		if !IsContain(list, item) {
			list = append(list, item)
		}
	}

	return list
}

func Regexp(str, substr string, n int) [][]string {
	return regexp.MustCompile(substr).FindAllStringSubmatch(str, n)
}
