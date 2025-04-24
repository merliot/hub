package device

import (
	"sort"
	"strings"
)

type deviceFlags struct {
	flags
}

const (
	FlagProgenitive  flags = 1 << iota // May have children
	FlagHttpPortMust                   // Device must have an HTTP port
	flagOnline                         // Device is online
	flagLocked                         // Device is locked
	flagDemo                           // Running in DEMO mode
	flagMetal                          // Device is running on real hardware
	flagGhost                          // Device is dead but may be resurrected later
	flagRoot                           // Device is root
)

func (df deviceFlags) list() string {
	var list []string
	names := map[flags]string{
		FlagProgenitive:  "PROGENITIVE",
		FlagHttpPortMust: "HTTP_PORT_MUST",
		flagOnline:       "ONLINE",
		flagLocked:       "LOCKED",
		flagDemo:         "DEMO",
		flagMetal:        "METAL",
		flagGhost:        "GHOST",
		flagRoot:         "ROOT",
	}
	for flag, name := range names {
		if df.isSet(flag) {
			list = append(list, name)
		}
	}
	sort.Strings(list)
	return strings.Join(list, "|")
}
