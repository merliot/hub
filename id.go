//go:build !tinygo

package hub

import (
	"crypto/rand"
	"encoding/hex"
)

// generateRandomId generates a hex-encoded 8-byte random ID in format
// 'xxxxxxxx-xxxxxxxx'
func generateRandomId() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	hexString := hex.EncodeToString(bytes)
	return hexString[:8] + "-" + hexString[8:]
}
