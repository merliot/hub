package main

import (
	"log"
	"os"

	"github.com/merliot/hub/uf2"
)

func main() {
	if len(os.Args) != 5 {
		log.Fatal("Usage: program uf2-input-file uf2-output-file from-string to-string")
	}

	input := os.Args[1]
	output := os.Args[2]
	from := os.Args[3]
	to := os.Args[4]

	if len(from) != len(to) {
		log.Fatal("from-string length must equal to-string")
	}

	uf2, err := uf2.Read(input)
	if err != nil {
		log.Fatal("Error reading/parsing uf2 file:", input, err.Error())
	}

	uf2.ReplaceBytes([]byte(from), []byte(to))

	err = uf2.Write(output)
	if err != nil {
		log.Fatal("Error writing uf2 file:", output, err.Error())
	}
}
