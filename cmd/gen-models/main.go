package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"

	"github.com/merliot/device/models"
)

var tmpl = template.Must(template.New("").Parse(`package main

import (
	"github.com/merliot/dean"
{{- range $key, $value := . }}
	"{{ $value.Module }}"
{{- end }}
)

var models = map[string]dean.ThingMaker{
{{- range $key, $value := . }}
	"{{ $key }}": {{ $value.Maker }},
{{- end }}
}
`))

func main() {

	inputFile := flag.String("input", "", "path to input models.json file")
	outputFile := flag.String("output", "", "path to output models.go file")
	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Please provide an input file using the -input flag.")
		return
	}

	models, err := models.Load(*inputFile)
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Create(*outputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = tmpl.Execute(file, models)
	if err != nil {
		log.Fatal(err)
	}
}
