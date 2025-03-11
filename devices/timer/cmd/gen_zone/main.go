package main

import (
	"bufio"
	"fmt"
	"os"
	"sort" // Added for sorting
	"strings"
)

//go:generate go run ./

func main() {
	// Open the input file
	inputFile, err := os.Open("zone1970.tab")
	if err != nil {
		fmt.Printf("Error opening zone1970.tab: %v\n", err)
		return
	}
	defer inputFile.Close()

	// Create the output file
	outputFile, err := os.Create("../../zone.go")
	if err != nil {
		fmt.Printf("Error creating zone.go: %v\n", err)
		return
	}
	defer outputFile.Close()

	// Create a slice to store timezone names
	var timezones []string

	// Read the input file line by line
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split the line by tabs
		fields := strings.Split(line, "\t")
		// Check if we have at least 3 columns
		if len(fields) >= 3 {
			// Add the third column (timezone name) to our slice
			timezones = append(timezones, fields[2])
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	// Sort the timezones alphabetically
	sort.Strings(timezones)

	// Write the Go file header
	_, err = outputFile.WriteString(`// This file automatically generated from cmd/gen_zone

package timer

// timezones is a static array of timezone names from zone1970.tab
var timezones = []string{
`)
	if err != nil {
		fmt.Printf("Error writing to zone.go: %v\n", err)
		return
	}

	// Write each timezone as a quoted string
	for _, tz := range timezones {
		_, err = outputFile.WriteString(fmt.Sprintf("\t\"%s\",\n", tz))
		if err != nil {
			fmt.Printf("Error writing timezone: %v\n", err)
			return
		}
	}

	// Write the closing bracket
	_, err = outputFile.WriteString("}\n")
	if err != nil {
		fmt.Printf("Error writing closing bracket: %v\n", err)
		return
	}

	fmt.Println("Successfully generated zone.go")
}
