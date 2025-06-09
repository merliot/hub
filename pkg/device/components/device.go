package components

import (
	"strconv"

	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

// DevicePageParams holds parameters for the DevicePage gomponent.
type DevicePageParams struct {
	Model      string
	Name       string
	BodyColors string
	Header     Node
	Body       Node
	Footer     Node
}

// DevicePage renders the main device page (device.tmpl).
func DevicePage(p DevicePageParams) Node {
	return El("html",
		Lang("en"),
		Head(
			Meta(Name("viewport"), Content("width=device-width, initial-scale=1")),
			Meta(Name("robots"), Content("noindex, nofollow")),
			Meta(Name("referrer"), Content("same-origin")),
			TitleEl(Text(p.Model+" - "+p.Name)),
			Link(Rel("icon"), Type("image/png"), Attr("sizes", "32x32"), Href("/images/favicon-32x32.png")),
			Link(Rel("icon"), Type("image/png"), Attr("sizes", "16x16"), Href("/images/favicon-16x16.png")),
			Link(Rel("stylesheet"), Href("https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css")),
			Script(Src("/js/htmx.min.js.gz")),
			Script(Src("/js/htmx-ext-ws.js.gz")),
			Script(Src("/js/util.js")),
		),
		Body(
			Class(p.BodyColors),
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
	return Div(
		Class("model-"+p.Model+" "+p.ClassOffline),
		Attr("id", p.ID),
		hx.Target("this"),
		hx.Swap("outerHTML"),
		Div(
			Class("flex flex-row ml-"+strconv.Itoa(p.Level*10)),
			Div(
				Class("panel flex flex-col m-1 p-2 min-w-[20rem] "+panelClass(p)),
				Div(
					Class("flex flex-row items-center justify-between"),
					hx.Get("/device/"+p.ID+"/show-view?view=info"),
					Span(
						Class("text-lg font-bold ml-2.5 w-24 cursor-pointer"),
						Text(p.Name),
					),
					Img(
						Class("icon"),
						Src("/model/"+p.Model+"/images/return.svg"),
					),
				),
				Div(
					Class("m-2.5 min-w-lg"),
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

// DeviceStateBody renders the device state body (device-state-body.tmpl).
func DeviceStateBody(id, stateJSON string) Node {
	return Div(
		Class("flex flex-col"),
		Div(
			Class("flex flex-row justify-center"),
			H3(Text("State")),
		),
		Pre(
			Class("text-sm"),
			Attr("hx-get", "/device/"+id+"/state"),
			Attr("hx-trigger", "load delay:1s"),
			Attr("hx-target", "this"),
			Attr("hx-swap", "innerHTML"),
			Text(stateJSON),
		),
	)
}
