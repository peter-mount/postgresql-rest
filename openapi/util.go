package openapi

import (
	"strings"
)

func addSlash(p string) string {
	if p == "" || strings.HasPrefix(p, "/") {
		return p
	}
	return "/" + p
}
func AddPrefix(prefix, path string) string {
	if prefix == "" {
		return addSlash(path)
	}
	return addSlash(prefix) + addSlash(path)
}
