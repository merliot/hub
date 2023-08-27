package common

type Targets map[string]string

var supportedTargets = Targets{
	"x86-64":      "Linux x86_64",
	"rpi":         "Raspberry Pi",
	"nano-rp2040": "Arduino Nano Connect rp2040",
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

