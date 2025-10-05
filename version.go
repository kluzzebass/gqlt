package gqlt

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var versionFile string

// Version returns the current version of the gqlt library
func Version() string {
	return strings.TrimSpace(versionFile)
}
