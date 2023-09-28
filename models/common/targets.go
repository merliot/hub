package common

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type GpioPins map[string]int

type Target struct {
	FullName string
	GpioPins
}

type Targets map[string]Target

// NOTE If modified, go run cmd/genjs/main.go to generate new
// deployTargetGpios.js file.

var supportedTargets = Targets{
	"x86-64": Target{
		FullName: "Linux x86_64",
		GpioPins: GpioPins{},
	},
	"rpi": Target{
		FullName: "Raspberry Pi",
		GpioPins: GpioPins{
			// maps GPIO label to physical pin number (see gobot.io)
			"GPIO04": 7,
			"GPIO17": 11,
			"GPIO18": 12,
			"GPIO27": 13,
			"GPIO22": 15,
			"GPIO23": 16,
			"GPIO24": 18,
			"GPIO25": 22,
			"GPIO05": 29,
			"GPIO06": 31,
			"GPIO12": 32,
			"GPIO13": 33,
			"GPIO19": 35,
			"GPIO16": 36,
			"GPIO26": 37,
			"GPIO20": 38,
			"GPIO21": 40,
		},
	},
	"nano-rp2040": Target{
		FullName: "Arduino Nano Connect rp2040",
		GpioPins: GpioPins{
			// maps label to GPIO
			"D2":  25,
			"D3":  15,
			"D4":  16,
			"D5":  17,
			"D6":  18,
			"D7":  19,
			"D8":  20,
			"D9":  21,
			"D10": 5,
			"D11": 7,
			"D12": 4,
			"D13": 6,
			"D14": 26,
			"D15": 27,
			"D16": 28,
			"D17": 29,
			"D18": 12,
			"D19": 13,
		},
	},
}

func makeTargets(targets []string) Targets {
	filtered := make(Targets)
	for _, target := range targets {
		if value, ok := supportedTargets[target]; ok {
			filtered[target] = value
		}
	}
	return filtered
}

// GenTargetJS generates JS const deployTargetGpios
func GenTargetJS() {
	file, err := os.Create("models/common/js/deployTargetGpios.js")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	file.WriteString("const deployTargetGpios = {\n")
	for k, v := range supportedTargets {
		gpioKeys := make([]string, 0, len(v.GpioPins))
		for gpioKey := range v.GpioPins {
			gpioKeys = append(gpioKeys, fmt.Sprintf(`"%s"`, gpioKey))
		}
		sort.Slice(gpioKeys, func(i, j int) bool {
			numA := extractNumber(gpioKeys[i])
			numB := extractNumber(gpioKeys[j])
			return numA < numB
		})

		fmt.Fprintf(file, "    \"%s\": [%s],\n", k, strings.Join(gpioKeys, ", "))
	}
	file.WriteString("};\n")
}

// extractNumber extracts the numeric part of a string and converts it to an integer
func extractNumber(s string) int {
	numStr := strings.TrimFunc(s, func(r rune) bool {
		return r < '0' || r > '9'
	})
	num, _ := strconv.Atoi(numStr)
	return num
}
