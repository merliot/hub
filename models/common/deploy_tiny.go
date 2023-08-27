//go:build tinygo

package common

import (
	"html/template"
	"net/http"
)

func (c *Common) deploy(templates *template.Template, w http.ResponseWriter, r *http.Request) {
	http.Error(w, "deploy not implemented for tinygo", http.StatusBadRequest)
}
