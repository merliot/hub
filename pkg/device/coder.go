package device

import (
	"net/url"

	"github.com/go-playground/form/v4"
)

var formEncoder = form.NewEncoder()
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

// Encoder is a custom URL values encoder
type Encoder interface {
	// Encode URL values
	Encode() (url.Values, error)
}

func encode(v any) (values url.Values, err error) {
	// If v has a custom encoder, use it
	if d, ok := v.(Encoder); ok {
		return d.Encode()
	}
	// Otherwise fall back to form Encoder.
	return formEncoder.Encode(v)
}
