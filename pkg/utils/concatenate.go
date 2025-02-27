package utils

import "strings"

func ConcatStrings(parts ...string) string {
	var sb strings.Builder
	for _, part := range parts {
		sb.WriteString(part)
	}
	return sb.String()
}

func ConcatSlices[T any](start []T, slices ...[]T) []T {
	r := start
	for _, slice := range slices {
		r = append(r, slice...)
	}
	return r
}

func BuildString(sb *strings.Builder, parts ...string) {
	for _, part := range parts {
		sb.WriteString(part)
	}
}
