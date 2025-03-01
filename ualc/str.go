package str

import (
	"strings"
)

// Index returns the index of the first occurrence of needle in haystack,
// or -1 if not found.
func Index(haystack, needle string) int {
	return strings.Index(haystack, needle)
}

// Split splits s into an array of substrings, using sep as the delimiter.
func Split(s, sep string) []string {
	return strings.Split(s, sep)
}

// Join joins the elements of arr (which should be an array of strings) into
// one string, with sep placed between them.
func Join(arr []string, sep string) string {
	return strings.Join(arr, sep)
}