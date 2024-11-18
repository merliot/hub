package target

// TODO maybe store this in a JSON file?

var AllTargets = Targets{
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
		TinyGo:   true,
		GpioPins: GpioPins{},
	},
	"wioterminal": Target{
		FullName: "Seeed Wio Terminal",
		TinyGo:   true,
		GpioPins: GpioPins{
			"D0": 40,
			"D1": 41,
			"D2": 7,
			"D3": 36,
			"D4": 37,
			"D5": 38,
			"D6": 4,
			"D7": 39,
			"D8": 6,
		},
	},
	"nano-rp2040": Target{
		FullName: "Arduino Nano Connect rp2040",
		TinyGo:   true,
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
