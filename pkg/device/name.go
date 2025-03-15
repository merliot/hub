//go:build !tinygo

package device

import (
	"errors"
	"unicode"
)

// validateName ensures the name is not empty, starts with a letter or number,
// and contains only letters, numbers, dashes, underscores, or spaces.
func validateName(name string) error {
	if len(name) == 0 {
		return errors.New("Name cannot be empty")
	}

	runes := []rune(name)

	// Check first character
	if !unicode.IsLetter(runes[0]) && !unicode.IsDigit(runes[0]) {
		return errors.New("Name must start with a letter or number")
	}

	// Check remaining characters
	for i, r := range runes {
		if i == 0 {
			continue
		}
		if !unicode.IsLetter(r) &&
			!unicode.IsDigit(r) &&
			r != ' ' &&
			r != '.' &&
			r != '-' &&
			r != '_' {
			return errors.New("Name can only contain letters, numbers, spaces, dots, dashes, or underscores")
		}
	}

	return nil
}
