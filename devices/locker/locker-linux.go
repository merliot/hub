//go:build !tinygo

package locker

import "embed"

//go:embed *.go images template
var fs embed.FS
