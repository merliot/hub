//go:build !tinygo

package hub

import (
	"os"
)

// Getenv returns the environment variable's value, or if empty, the
// defaultValue
func Getenv(name string, defaultValue string) string {
	value, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}
	return value
}

// Setenv sets the environment variable
func Setenv(name, value string) {
	os.Setenv(name, value)
}
