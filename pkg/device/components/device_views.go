package components

import (
	"strconv"

	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

// DeviceDetailParams holds parameters for DeviceDetail.
type DeviceDetailParams struct {
	Model          string
	ClassOffline   string
	ID             string
	Uniq           string
	Level          int
	IsOnline       bool
	BgColor        string
	TextColor      string
	BorderColor    string
	Name           string
	DeployParams   string
	SessionID      string
	RenderChildren Node
	Buttons        []Node
	Body           Node
}

// DeviceDetail renders the device detail panel.
func DeviceDetail(p DeviceDetailParams) Node {
	var body Node
	if p.Body != nil {
		body = p.Body
	} else if p.DeployParams == "" {
		body = UndefinedDetail(p.Name)
	} else {
		body = BodyDetail()
	}
	return Div(
		Class("model-"+p.Model+" "+p.ClassOffline),
		Attr("id", p.Uniq),
		hx.Target("this"),
		hx.Swap("outerHTML"),
		Div(
			Class("flex flex-row ml-"+strconv.Itoa(p.Level*10)),
			Div(
				Class("flex flex-col m-1 p-2 min-w-fit "+panelClass(DeviceStateParams{
					IsOnline: p.IsOnline, BgColor: p.BgColor, TextColor: p.TextColor, BorderColor: p.BorderColor,
				})),
				Div(
					Class("flex flex-row mb-5 items-center justify-between"),
					Span(
						Class("text-lg font-bold ml-2.5 w-full cursor-pointer"),
						hx.Get("/device/"+p.ID+"/show-view?view=overview"),
						Text(p.Name),
					),
					Div(
						Class("flex flex-row"),
						Group(p.Buttons),
					),
				),
				body,
			),
		),
		If(p.RenderChildren != nil, p.RenderChildren),
	)
}

// DeviceOverviewParams holds parameters for DeviceOverview.
type DeviceOverviewParams struct {
	Model        string
	ClassOffline string
	ID           string
	Level        int
	IsOnline     bool
	BgColor      string
	TextColor    string
	BorderColor  string
	Name         string
	DeployParams string
}

// DeviceOverview renders the device overview panel.
func DeviceOverview(p DeviceOverviewParams) Node {
	var body Node
	if p.DeployParams == "" {
		body = UndefinedOverview()
	} else {
		body = BodyOverview()
	}
	return Div(
		Class("model-"+p.Model+" "+p.ClassOffline),
		Attr("id", p.ID),
		hx.Target("this"),
		hx.Swap("outerHTML"),
		Div(
			Class("flex flex-row ml-"+strconv.Itoa(p.Level*10)),
			Div(
				Class("panel flex flex-row m-1 p-2 min-w-[20rem] justify-between "+panelClass(DeviceStateParams{
					IsOnline: p.IsOnline, BgColor: p.BgColor, TextColor: p.TextColor, BorderColor: p.BorderColor,
				})+" cursor-pointer"),
				hx.Get("/device/"+p.ID+"/show-view?view=detail"),
				Div(
					Class("flex flex-row w-full items-center"),
					Span(Class("text-lg font-bold ml-2.5 w-24"), Text(p.Name)),
					body,
				),
			),
		),
	)
}

// DeviceSettingsParams holds parameters for DeviceSettings.
type DeviceSettingsParams struct {
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
	RenderChildren Node
}

// DeviceSettings renders the device settings panel.
func DeviceSettings(p DeviceSettingsParams) Node {
	return Div(
		Class("model-"+p.Model+" "+p.ClassOffline),
		Attr("id", p.ID),
		hx.Target("this"),
		hx.Swap("outerHTML"),
		Div(
			Class("flex flex-row ml-"+strconv.Itoa(p.Level*10)),
			Div(
				Class("panel flex flex-col m-1 p-2 min-w-[20rem] "+panelClass(DeviceStateParams{
					IsOnline: p.IsOnline, BgColor: p.BgColor, TextColor: p.TextColor, BorderColor: p.BorderColor,
				})),
				Div(
					Class("flex flex-row items-center justify-between"),
					hx.Get("/device/"+p.ID+"/show-view?view=detail"),
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
					Class("m-2.5 max-w-lg"),
					DeviceDownloadBody(p.ID, p.SessionID),
				),
			),
		),
		If(p.RenderChildren != nil, p.RenderChildren),
	)
}

// BodyDetail renders a placeholder for body-detail.tmpl.
func BodyDetail() Node {
	return Div(
		Class("p-2.5 bg-black"),
		Span(Class("text-red-500"), Text("Missing body-detail.tmpl")),
	)
}

// UndefinedDetail renders the undefined-detail.tmpl fragment.
func UndefinedDetail(name string) Node {
	return Div(
		Class("flex flex-row min-h-28 mb-5 items-center justify-center"),
		Text("Click"),
		// ButtonSettings would be called here in real integration
		Text(" to setup and download "+name),
	)
}

// BodyOverview renders a placeholder for body-overview.tmpl.
func BodyOverview() Node {
	return Div(
		Class("p-2.5 bg-black"),
		Span(Class("text-red"), Text("Missing body-overview.tmpl")),
	)
}

// UndefinedOverview renders the undefined-overview.tmpl fragment.
func UndefinedOverview() Node {
	return Span(Text("Undefined"))
}

// DeviceDownloadBody is a placeholder for the download body.
func DeviceDownloadBody(id, sessionID string) Node {
	return Div(
		Class("p-2.5 bg-black"),
		Text("Missing device-download-body.tmpl"),
	)
}
