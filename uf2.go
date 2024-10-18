package hub

import (
	"html/template"
)

// Random string to embed in UF2 we can search for later to locate params
const uf2Magic = "Call the Doctor!  Miss you Dan."

type uf2Params struct {
	MagicStart   string
	Ssid         string
	Passphrase   string
	Id           string
	Model        string
	Name         string
	DeployParams template.HTML
	User         string
	Passwd       string
	DialURLs     string
	MagicEnd     string
}
