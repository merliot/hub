//go:build tinygo

package device

// Tinygo doesn't have an environment, so simulate one
var envs = make(map[string]string)

func Getenv(name, defaultValue string) string {
	value, ok := envs[name]
	if !ok {
		return defaultValue
	}
	return value
}

func Setenv(name, value string) {
	envs[name] = value
}
