package components

import (
	"fmt"

	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	html "maragu.dev/gomponents/html"
)

// DevicePageParams holds parameters for the DevicePage gomponent.
type DevicePageParams struct {
	Model      string
	Name       string
	BodyColors string
	Section    string
	Header     Node
	Body       Node
	Footer     Node
}

// DevicePage renders the main device page (device.tmpl).
func DevicePage(p DevicePageParams) Node {
	return El("html",
		html.Lang("en"),
		html.Head(
			html.Meta(html.Name("viewport"), html.Content("width=device-width, initial-scale=1")),
			html.Meta(html.Name("robots"), html.Content("noindex, nofollow")),
			html.Meta(html.Name("referrer"), html.Content("same-origin")),
			html.Link(html.Rel("icon"), html.Type("image/png"), Attr("sizes", "32x32"), html.Href("/images/favicon-32x32.png")),
			html.Link(html.Rel("icon"), html.Type("image/png"), Attr("sizes", "16x16"), html.Href("/images/favicon-16x16.png")),
			html.Link(html.Rel("stylesheet"), html.Href("https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css")),
			html.Script(html.Src("/js/htmx.min.js.gz")),
			html.Script(html.Src("/js/htmx-ext-ws.js.gz")),
			html.Script(html.Src("/js/util.js")),
		),
		html.Body(
			html.Class(p.BodyColors),
			Group([]Node{
				p.Header,
				p.Body,
				p.Footer,
			}),
		),
	)
}

// DeviceStateParams holds parameters for the DeviceState gomponent.
type DeviceStateParams struct {
	Model          string
	ClassOffline   string
	ID             string
	Level          int
	IsOnline       bool
	BgColor        string
	TextColor      string
	BorderColor    string
	Name           string
	SessionID      string
	Body           Node
	RenderChildren Node // Node or function to render children
}

// DeviceState renders the device state panel (device-state.tmpl).
func DeviceState(p DeviceStateParams) Node {
	return html.Div(
		html.Class("model-"+p.Model+" "+p.ClassOffline),
		Attr("id", p.ID),
		hx.Target("this"),
		hx.Swap("outerHTML"),
		html.Div(
			html.Class("flex flex-row ml-"+itoa(p.Level*10)),
			html.Div(
				html.Class("panel flex flex-col m-1 p-2 min-w-[20rem] "+panelClass(p)),
				html.Div(
					html.Class("flex flex-row items-center justify-between"),
					hx.Get("/device/"+p.ID+"/show-view?view=info"),
					html.Span(
						html.Class("text-lg font-bold ml-2.5 w-24 cursor-pointer"),
						Text(p.Name),
					),
					html.Img(
						html.Class("icon"),
						html.Src("/model/"+p.Model+"/images/return.svg"),
					),
				),
				html.Div(
					html.Class("m-2.5 min-w-lg"),
					p.Body,
				),
			),
		),
		If(p.RenderChildren != nil, p.RenderChildren),
	)
}

// panelClass returns the appropriate class string for the panel based on online status.
func panelClass(p DeviceStateParams) string {
	if !p.IsOnline {
		return "border-dotted grayscale text-black bg-white"
	}
	return p.BgColor + " " + p.TextColor + " " + p.BorderColor
}

// itoa is a helper to convert int to string.
func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}

// DeviceStateBody renders the device state body (device-state-body.tmpl).
func DeviceStateBody(id, stateJSON string) Node {
	return html.Div(
		html.Class("flex flex-col"),
		html.Div(
			html.Class("flex flex-row justify-center"),
			html.H3(Text("State")),
		),
		html.Pre(
			html.Class("text-sm"),
			Attr("hx-get", "/device/"+id+"/state"),
			Attr("hx-trigger", "load delay:1s"),
			Attr("hx-target", "this"),
			Attr("hx-swap", "innerHTML"),
			Text(stateJSON),
		),
	)
}
