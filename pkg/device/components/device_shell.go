package components

import (
	. "maragu.dev/gomponents"
	html "maragu.dev/gomponents/html"
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
		html.Lang("en"),
		html.Head(
			html.Meta(html.Name("viewport"), html.Content("width=device-width, initial-scale=1")),
			html.Meta(html.Name("robots"), html.Content("noindex, nofollow")),
			html.Meta(html.Name("referrer"), html.Content("same-origin")),
			html.Link(html.Rel("icon"), html.Type("image/png"), Attr("sizes", "32x32"), html.Href("/images/favicon-32x32.png")),
			html.Link(html.Rel("icon"), html.Type("image/png"), Attr("sizes", "16x16"), html.Href("/images/favicon-16x16.png")),
			html.Title(p.Model+" - "+p.Name),
			html.Link(html.Rel("stylesheet"), html.Href("https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css")),
			html.Script(html.Src("/js/htmx.min.js.gz")),
			html.Script(html.Src("/js/htmx-ext-ws.js.gz")),
			html.Script(html.Src("/js/util.js")),
		),
		html.Body(
			html.Class(p.BodyColors),
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
	return html.Header(
		Group([]Node{
			p.SaveButton,
			// Add more header content here as needed
		}),
	)
}

// DeviceHome renders the device home page (device-home.tmpl).
func DeviceHome(sessionView Node) Node {
	return html.Div(
		html.Class("m-4"),
		sessionView,
	)
}

// Placeholders for device info, download, etc. (to be filled in as needed)
func DeviceInfoPanel() Node {
	return html.Div(
		html.Class("p-2.5 bg-black"),
		html.Span(html.Class("text-red"), Text("Missing device info panel")),
	)
}

func DeviceDownloadPanel() Node {
	return html.Div(
		html.Class("p-2.5 bg-black"),
		html.Span(html.Class("text-red"), Text("Missing device download panel")),
	)
}
