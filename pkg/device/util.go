package device

import (
	"fmt"
	"time"
)

func formatDuration(d time.Duration) string {
	const (
		day   = time.Hour * 24
		month = day * 30  // Approximate month
		year  = day * 365 // Approximate year
	)

	// Define the units and their respective labels
	units := []struct {
		duration time.Duration
		label    string
	}{
		{year, "y"},
		{month, "m"},
		{day, "d"},
		{time.Hour, "h"},
		{time.Minute, "m"},
		{time.Second, "s"},
	}

	// Result string
	result := ""
	for _, unit := range units {
		if value := d / unit.duration; value > 0 {
			result += fmt.Sprintf("%d%s ", value, unit.label)
			d -= value * unit.duration
		}
	}

	// Trim any trailing space and return
	return result[:len(result)-1] // Slice off the trailing space
}
