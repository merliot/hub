package main

import (
	"fmt"
	"os"

	"github.com/merliot/hub/pkg/device"
	"github.com/merliot/hub/pkg/models"
)

//go:generate sh -x -c "go run ../gen-models/ ../../models.json ../../pkg/models/models.go"
//go:generate sh -x -c "go run ./ base ../../bin"

func main() {
	progName := os.Args[0]

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Error: No command provided\n")
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [arguments]\n", progName)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "base":
		dir := "../../bin/"
		if len(os.Args) > 2 {
			dir = os.Args[2]
		}

		model := ""
		if len(os.Args) > 3 {
			model = os.Args[3]
		}

		target := ""
		if len(os.Args) > 4 {
			target = os.Args[4]
		}

		server := device.NewServer(device.WithModels(models.AllModels))
		if err := server.Uf2GenerateBaseImages(dir, model, target); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	case "dump":
		if len(os.Args) != 3 {
			fmt.Fprintf(os.Stderr, "Usage: %s dump <file>\n", progName)
			os.Exit(1)
		}
		dump, err := device.Uf2Dump(os.Args[2])
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
