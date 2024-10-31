//go:build !tinygo

package qrcode

import "embed"

//go:embed *.go template
var fs embed.FS
