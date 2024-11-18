//go:generate go run ./

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

type model struct {
	Package string
	Source  string
	Maker   string
}

type models map[string]model

const modelsTemplate = `// This file auto-generated from ./cmd/gen-models using models.json as input

package models

import (
	"github.com/merliot/hub/pkg/device"
{{- range . }}
	"{{ .Package }}"
{{- end }}
)

var AllModels = device.ModelMap{
{{- range $key, $value := . }}
	"{{$key}}": {{title $key}},
{{- end }}
}

{{- range $key, $value := . }}
var {{title $key}} = device.Model{
	Package: "{{$value.Package}}",
	Source: "{{$value.Source}}",
	Maker: {{$value.Maker}},
}

{{- end }}
`

func main() {

	var models models

	data, err := ioutil.ReadFile("../../models.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &models)
	if err != nil {
		panic(err)
	}

	outFile, err := os.Create("../../pkg/models/models.go")
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	// Use template to write the models.go file
	tmpl, err := template.New("models").Funcs(template.FuncMap{
		"title": func(s string) string {
			return strings.Title(s)
		},
	}).Parse(modelsTemplate)
	if err != nil {
		panic(err)
	}

	// Execute the template with the models data
	if err := tmpl.Execute(outFile, models); err != nil {
		panic(err)
	}

	fmt.Println("Created Models from models.json")
}
