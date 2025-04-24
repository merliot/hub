package device

type flags uint32

func (f *flags) set(flags flags) {
	*f = *f | flags
}

func (f *flags) unSet(flags flags) {
	*f = *f & ^flags
}

func (f flags) isSet(flags flags) bool {
	return f&flags == flags
}
