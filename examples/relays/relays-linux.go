//go:build !tinygo

package relays

import "embed"

//go:embed images *.go template
var fs embed.FS
