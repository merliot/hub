package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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

var AllModels = device.Models{
{{- range $key, $value := . }}
	"{{$key}}": &{{title $key}},
{{- end }}
}

{{- range $key, $value := . }}
var {{title $key}} = device.Model{
	Package: "{{$value.Package}}",
	Maker: {{$value.Maker}},
}

{{- end }}
`

func main() {

	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <input models.json path> <output models.go path>")
		os.Exit(1)
	}

	inputPath := os.Args[1]
	outputPath := os.Args[2]

	var models models

	data, err := ioutil.ReadFile(inputPath)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &models)
	if err != nil {
		panic(err)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}

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

	outFile.Close()

	// Clean it up
	exec.Command("gofmt", "-s", "-w", outputPath).CombinedOutput()
}
