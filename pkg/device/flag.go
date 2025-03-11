package device

type flags uint32

// Device flags
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

func (f *flags) set(flags flags) {
	*f = *f | flags
}

func (f *flags) unSet(flags flags) {
	*f = *f & ^flags
}

func (f flags) isSet(flags flags) bool {
	return f&flags == flags
}
