package device

import "sync/atomic"

type flags struct {
	atomic.Uint32
}

const (
	FlagProgenitive  uint32 = 1 << iota // May have children
	FlagHttpPortMust                    // Device must have an HTTP port
	flagOnline                          // Device is online
	flagDirty                           // Has unsaved changes
	flagLocked                          // Device is locked
	flagDemo                            // Running in DEMO mode
	flagMetal                           // Device is running on real hardware
	flagGhost                           // Device is dead but may be resurrected later
	flagRoot                            // Device is root
)

func (f *flags) set(flags uint32) {
	for {
		orig := f.Load()
		updated := orig | flags
		if f.CompareAndSwap(orig, updated) {
			return
		}
	}
}

func (f *flags) unSet(flags uint32) {
	for {
		orig := f.Load()
		updated := orig &^ flags
		if f.CompareAndSwap(orig, updated) {
			return
		}
	}
}

func (f flags) isSet(flags uint32) bool {
	return (f.Load() & flags) == flags
}
