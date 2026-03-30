package utils

import (
	"regexp"
	"strings"
)

// Slugify converts a string into a URL-friendly slug.
// Example: "Hello World!" -> "hello-world"
func Slugify(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Remove special characters
	reg := regexp.MustCompile(`[^a-z0-9\s-]`)
	s = reg.ReplaceAllString(s, "")

	// Replace spaces with hyphens
	s = strings.ReplaceAll(s, " ", "-")

	// Remove consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	s = reg.ReplaceAllString(s, "-")

	// Remove leading/trailing hyphens
	s = strings.Trim(s, "-")

	return s
}
