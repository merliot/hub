package components

import (
	. "maragu.dev/gomponents"
	html "maragu.dev/gomponents/html"
)

type DeviceDownloadTargetParams struct {
	SessionID      string
	SelectedTarget string
	WantsWifi      bool
	WantsHttpPort  bool
}

func DeviceDownloadTarget(p DeviceDownloadTargetParams) Node {
	return html.Div(
		html.Class("p-4"),
		html.H3(html.Class("text-xl font-bold mb-4"), Text("Device Download Target")),
		html.P(Text("Session ID: "+p.SessionID)),
		html.P(Text("Selected Target: "+p.SelectedTarget)),
		If(p.WantsWifi, html.P(Text("WiFi configuration required."))),
		If(p.WantsHttpPort, html.P(Text("HTTP port configuration required."))),
	)
}
