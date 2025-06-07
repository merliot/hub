package components

import (
	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	html "maragu.dev/gomponents/html"
)

// ButtonNew renders the "+ New" button for adding a new device.
func ButtonNew(deviceName, deviceID string) Node {
	return html.Button(
		html.Title("Add a new device to "+deviceName),
		hx.Get("/new-modal/"+deviceID),
		hx.Target("body"),
		hx.Swap("beforeend"),
		Text("+ New"),
	)
}
