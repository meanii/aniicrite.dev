package models

import (
	"regexp"
	"strings"
)

var (
	nonSlug   = regexp.MustCompile(`[^a-z0-9]+`)
	trimDash  = regexp.MustCompile(`^-+|-+$`)
	multiDash = regexp.MustCompile(`-{2,}`)
)

// Slugify converts arbitrary text into a URL-safe slug.
func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonSlug.ReplaceAllString(s, "-")
	s = multiDash.ReplaceAllString(s, "-")
	s = trimDash.ReplaceAllString(s, "")
	return s
}
