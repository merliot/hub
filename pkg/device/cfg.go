package device

import (
	"embed"
	"html/template"
	"time"
)

// Config is the device model configuration
type Config struct {
	// Model is the device model name
	Model string
	// Parents are the supported parent models
	Parents []string
	// Flags see FlagXxxx
	Flags flags
	// The device state
	State   any
	stateMu mutex
	// The device's embedded static file system
	FS *embed.FS
	// Targets support by device
	Targets []string
	// PollPeriod is the device polling period.  The default is 1 second.
	// The range is [1..forever) seconds.
	PollPeriod time.Duration
	// BgColor is the device background color
	BgColor string
	// FgColor is the device forground (text, border) color
	FgColor string
	// PacketHandlers
	PacketHandlers
	// Custom device APIs
	APIs
	// Custom device funcs
	template.FuncMap
}
