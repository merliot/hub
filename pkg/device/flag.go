package device

type flags uint32

const (
	FlagProgenitive  flags = 1 << iota // May have children
	FlagHttpPortMust                   // Device must have an HTTP port
	flagOnline                         // Device is online
	flagDirty                          // Has unsaved changes
	flagLocked                         // Device is locked
	flagDemo                           // Running in DEMO mode
	flagMetal                          // Device is running on real hardware
	flagGhost                          // Device is dead but may be resurrected later
	flagRoot                           // Device is root
)

func (f *flags) _set(flags flags) {
	*f = *f | flags
}

func (d *device) set(flags flags) {
	d.Lock()
	defer d.Unlock()
	d._set(flags)
}

func (f *flags) _unSet(flags flags) {
	*f = *f & ^flags
}

func (d *device) unSet(flags flags) {
	d.Lock()
	defer d.Unlock()
	d._unSet(flags)
}

func (f flags) _isSet(flags flags) bool {
	return f&flags == flags
}

func (d *device) isSet(flags flags) bool {
	d.RLock()
	defer d.RUnlock()
	return d._isSet(flags)
}
