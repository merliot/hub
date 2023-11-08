//go:build !tinygo

package charge

import (
	"embed"
	"html/template"
)

//go:embed *
var fs embed.FS

type targetStruct struct {
	templates *template.Template
}

func (c *Charge) targetNewX() {
	c.CompositeFs.AddFS(fs)
	c.templates = c.CompositeFs.ParseFS("template/*")
}

