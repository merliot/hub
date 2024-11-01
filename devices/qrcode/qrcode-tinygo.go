//go:build tinygo

package qrcode

import (
	"embed"
	"net/http"
)

var fs embed.FS

func (q *qrcode) generate(w http.ResponseWriter, r *http.Request) {}
