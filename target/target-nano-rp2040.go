//go:build nano_rp2040

package target

func Pin(pin string) (GpioPin, bool) {
	gpio, ok := AllTargets["nano-rp2040"].GpioPins[pin]
	return gpio, ok
}
