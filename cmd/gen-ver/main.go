package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var template = `package hub

var version = "%s"
`

//go:generate go run main.go
func main() {
	// Get the latest Git version tag
	tag := ""
	tagCmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	stdoutStderr, err := tagCmd.CombinedOutput()
	if err == nil {
		tag = strings.TrimSpace(string(stdoutStderr))
	} else {
		fmt.Println("Error:", err.Error(), string(stdoutStderr))
	}

	// Get the latest Git SHA
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	sha := strings.TrimSpace(string(output))
	sha = sha[:7]

	file, err := os.Create("../../version.go")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer file.Close()

	version := sha
	if tag != "" {
		version = tag + "-" + sha
	}

	content := fmt.Sprintf(template, version)

	_, err = file.WriteString(content)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Println(version, "written to version.go")
}
