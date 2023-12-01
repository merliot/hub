package common

type GpioPins map[string]int

type Target struct {
	FullName string
	GpioPins
}

type Targets map[string]Target

// NOTE If modified, go run cmd/genjs/main.go to generate new
// deployTargetGpios.js file.

var supportedTargets = Targets{
	"demo": Target{
		FullName: "Demo Mode",
		GpioPins: GpioPins{
			"DEMO0": 0,
			"DEMO1": 1,
			"DEMO2": 2,
			"DEMO3": 3,
		},
	},
	"x86-64": Target{
		FullName: "Linux x86-64",
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
	"pyportal": Target{
		FullName: "Adafruit PyPortal",
		GpioPins: GpioPins{},
	},
	"wioterminal": Target{
		FullName: "Seeed Wio Terminal",
		GpioPins: GpioPins{},
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
