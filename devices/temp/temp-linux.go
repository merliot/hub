//go:build !tinygo

package temp

import "embed"

//go:embed *.go template
var fs embed.FS
