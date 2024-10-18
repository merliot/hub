package target

type GpioPin int                 // machine Pin
type GpioPins map[string]GpioPin // key: pin display name (e.g. 'D0'); value: machine Pin (e.g. 40)

type Target struct {
	FullName string
	GpioPins
	TinyGo bool
}

type Targets map[string]Target // key: target short name (e.g. 'rpi')

func MakeTargets(targets []string) Targets {
	filtered := make(Targets)
	for _, target := range targets {
		if t, ok := AllTargets[target]; ok {
			filtered[target] = t
		}
	}
	return filtered
}

func TinyGoTargets(targets []string) []string {
	var out []string
	for _, target := range targets {
		if t, ok := AllTargets[target]; ok {
			if t.TinyGo {
				out = append(out, target)
			}
		}
	}
	return out
}
