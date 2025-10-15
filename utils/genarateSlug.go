package utils

import (
	"strings"

	"github.com/gosimple/slug"
)

func GenerateSlug(input string, maxLength int) string {
	s := slug.Make(input)
	if len(s) > maxLength {
		s = s[:maxLength]
	}
	s = strings.Trim(s, "-")
	return s
}

func MaxLength(input string,maxLength int) string {
	if len(input) > maxLength {
		input = input[:maxLength]
	}
	return input
}