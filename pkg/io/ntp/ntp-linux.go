//go:build !tinygo

package ntp

func SetSystemTime() error {
	// TODO use NTP to set runtime system time
	return nil
}
