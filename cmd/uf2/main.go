package main

import (
	"fmt"
	"os"

	"github.com/merliot/hub"
	"github.com/merliot/hub/models"
)

//go:generate go run ../gen-models/
//go:generate go run ./ base

func main() {
	progName := os.Args[0]

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Error: No command provided\n")
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [arguments]\n", progName)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "base":
		hub.Models = models.AllModels

		model := ""
		if len(os.Args) > 2 {
			model = os.Args[2]
		}

		target := ""
		if len(os.Args) > 3 {
			target = os.Args[3]
		}

		if err := hub.Uf2GenerateBaseImages("uf2s/", model, target); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	case "dump":
		if len(os.Args) != 3 {
			fmt.Fprintf(os.Stderr, "Usage: %s dump <file>\n", progName)
			os.Exit(1)
		}
		dump, err := hub.Uf2Dump(os.Args[2])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(dump)

		/*
			case "create":
				// Define flags for the create command.
				createCmd := flag.NewFlagSet("create", flag.ExitOnError)
				target := createCmd.String("target", "", "Target device (required)")
				specFile := createCmd.String("spec", "", "Path to JSON spec file")

				_ = createCmd.Parse(os.Args[2:])

				if *target == "" {
					fmt.Fprintf(os.Stderr, "Error: -target is required\n")
					fmt.Fprintf(os.Stderr, "Usage: %s create -target <target> [-spec <spec file>]\n", progName)
					os.Exit(1)
				}
				createCommand(*target, *specFile)
		*/
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n", os.Args[1])
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [arguments]\n", progName)
		os.Exit(1)
	}
}
