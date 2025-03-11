//go:build !tinygo

package device

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"unicode"
)

// generateRandomId generates a hex-encoded 4-byte random ID in format
// 'xxxxxxxx'
func generateRandomId() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// validateId validates an ID string for use in URLs.
// It ensures the ID is not empty, starts with a letter or number,
// and contains only letters, numbers, dashes, or underscores (no spaces).
func validateId(id string) error {
	if len(id) == 0 {
		return errors.New("Id cannot be empty")
	}

	runes := []rune(id)

	// Check first character
	if !unicode.IsLetter(runes[0]) && !unicode.IsDigit(runes[0]) {
		return errors.New("Id must start with a letter or number")
	}

	// Check remaining characters
	for i, r := range runes {
		if i == 0 {
			continue
		}
		if !unicode.IsLetter(r) &&
			!unicode.IsDigit(r) &&
			r != '-' &&
			r != '_' {
			return errors.New("Id can only contain letters, numbers, dashes, or underscores")
		}
	}

	return nil
}
