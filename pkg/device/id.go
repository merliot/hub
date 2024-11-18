//go:build !tinygo

package device

import (
	"crypto/rand"
	"encoding/hex"
)

// generateRandomId generates a hex-encoded 4-byte random ID in format
// 'xxxxxxxx'
func generateRandomId() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
