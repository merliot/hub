//go:build !tinygo

package gps

import "embed"

//go:embed *.go images template
var fs embed.FS
