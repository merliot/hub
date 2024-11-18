package device

import (
	"net/url"

	"github.com/go-playground/form/v4"
)

var formDecoder = form.NewDecoder()

// Decoder is a custom URL values decoder
type Decoder interface {
	// Decode URL values
	Decode(values url.Values) error
}

func decode(v any, values url.Values) error {
	// If v has a custom decoder, use it
	if d, ok := v.(Decoder); ok {
		return d.Decode(values)
	}
	// Otherwise fall back to form Decoder.
	return formDecoder.Decode(v, values)
}
