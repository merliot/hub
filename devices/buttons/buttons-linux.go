//go:build !tinygo

package buttons

import "embed"

//go:embed images *.go template
var fs embed.FS
