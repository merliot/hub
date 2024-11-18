//go:build wioterminal

package target

func Pin(pin string) (GpioPin, bool) {
	gpio, ok := AllTargets["wioterminal"].GpioPins[pin]
	return gpio, ok
}
