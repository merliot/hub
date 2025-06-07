package components

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type DeviceShellParams struct {
	Model      string
	Name       string
	BodyColors string
	Section    string
	Header     Node
	Body       Node
	Footer     Node
}

// DeviceShell renders the main device HTML shell (device.tmpl).
func DeviceShell(p DeviceShellParams) Node {
	return El("html",
		Lang("en"),
		Head(
			Meta(Name("viewport"), Content("width=device-width, initial-scale=1")),
			Meta(Name("robots"), Content("noindex, nofollow")),
			Meta(Name("referrer"), Content("same-origin")),
			Link(Rel("icon"), Type("image/png"), Attr("sizes", "32x32"), Href("/images/favicon-32x32.png")),
			Link(Rel("icon"), Type("image/png"), Attr("sizes", "16x16"), Href("/images/favicon-16x16.png")),
			Title(p.Model+" - "+p.Name),
			Link(Rel("stylesheet"), Href("https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css")),
			Script(Src("/js/htmx.min.js.gz")),
			Script(Src("/js/htmx-ext-ws.js.gz")),
			Script(Src("/js/util.js")),
		),
		Body(
			Class(p.BodyColors),
			Group([]Node{p.Header, p.Body, p.Footer}),
		),
	)
}

type DeviceHeaderParams struct {
	SaveButton Node
	// Add more fields as needed for additional header content
}

// DeviceHeader renders the device header (device-header.tmpl).
func DeviceHeader(p DeviceHeaderParams) Node {
	return Header(
		Group([]Node{
			p.SaveButton,
			// Add more header content here as needed
		}),
	)
}

// DeviceHome renders the device home page (device-home.tmpl).
func DeviceHome(sessionView Node) Node {
	return Div(
		Class("m-4"),
		sessionView,
	)
}

// Placeholders for device info, download, etc. (to be filled in as needed)
func DeviceInfoPanel() Node {
	return Div(
		Class("p-2.5 bg-black"),
		Span(Class("text-red"), Text("Missing device info panel")),
	)
}

func DeviceDownloadPanel() Node {
	return Div(
		Class("p-2.5 bg-black"),
		Span(Class("text-red"), Text("Missing device download panel")),
	)
}
