package hub

type flags uint32

const (
	FlagProgenitive   flags = 1 << iota // May have children
	FlagWantsHttpPort                   // HTTP port not optional
	flagOnline                          // Device is online
	flagDirty                           // Has unsaved changes
	flagLocked                          // Administratively locked
	flagDemo                            // Running in DEMO mode
	flagMetal                           // Device is running on real hardware
)

func (f *flags) Set(flags flags) {
	*f = *f | flags
}

func (f *flags) Unset(flags flags) {
	*f = *f & ^flags
}

func (f flags) IsSet(flags flags) bool {
	return f&flags == flags
}
