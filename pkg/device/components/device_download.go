package components

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type DeviceDownloadTargetParams struct {
	SessionID      string
	SelectedTarget string
	WantsWifi      bool
	WantsHttpPort  bool
}

func DeviceDownloadTarget(p DeviceDownloadTargetParams) Node {
	return Div(
		Class("p-4"),
		H3(Class("text-xl font-bold mb-4"), Text("Device Download Target")),
		P(Text("Session ID: "+p.SessionID)),
		P(Text("Selected Target: "+p.SelectedTarget)),
		If(p.WantsWifi, P(Text("WiFi configuration required."))),
		If(p.WantsHttpPort, P(Text("HTTP port configuration required."))),
	)
}
