package components

import (
	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	html "maragu.dev/gomponents/html"
)

type DeviceInfoParams struct {
	Model        string
	ClassOffline string
	ID           string
	Level        int
	IsOnline     bool
	BgColor      string
	TextColor    string
	BorderColor  string
	Name         string
	SessionID    string
	Targets      map[string]struct{ FullName string }
	Target       string
	Port         string
	Package      string
	IsLocked     bool
	StateJSON    string
}

func DeviceInfo(p DeviceInfoParams) Node {
	return html.Div(
		html.Class("model-"+p.Model+" "+p.ClassOffline),
		Attr("id", p.ID),
		hx.Target("this"),
		hx.Swap("outerHTML"),
		html.Div(
			html.Class("flex flex-row ml-"+itoa(p.Level*10)),
			html.Div(
				html.Class("panel flex flex-col m-1 p-2 min-w-[20rem] "+panelClass(DeviceStateParams{
					IsOnline: p.IsOnline, BgColor: p.BgColor, TextColor: p.TextColor, BorderColor: p.BorderColor,
				})),
				html.Div(
					html.Class("flex flex-row items-center justify-between"),
					hx.Get("/device/"+p.ID+"/show-view?view=detail"),
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
					html.Class("m-2.5 max-w-3xl"),
					DeviceInfoBody(p),
				),
			),
		),
	)
	// RenderChildren would go here if needed
}

func DeviceInfoBody(p DeviceInfoParams) Node {
	return html.Div(
		html.Class("flex flex-col"),
		html.Div(
			html.Class("flex flex-row justify-center"),
			html.H3(Text("Info")),
		),
		Raw(`<style>.info-grid { display: grid; grid-template-columns: auto auto; column-gap: 20px; row-gap: 10px; align-items: center; } .info-grid .row-header { font-weight: bold; text-align: right; } .hr-row { grid-column: span 2; } hr { width: 100%; }</style>`),
		html.Div(
			html.Class("info-grid"),
			html.Div(html.Class("row-header"), html.Span(Text("ID"))),
			html.Div(html.Span(Text(p.ID))),
			html.Div(html.Class("row-header"), html.Span(Text("Model"))),
			html.Div(html.Span(Text(p.Model))),
			html.Div(html.Class("row-header"), html.Span(Text("Name"))),
			html.Div(
				html.Class("flex flex-row"),
				Attr("id", p.ID+"-edit-name"),
				html.Span(html.Class("mr-4"), Text(p.Name)),
				If(p.IsLocked,
					html.Img(
						html.Src("/images/edit.svg"),
						Attr("onclick", "alert('Sorry, cannot change name: device is locked')"),
					),
				),
				If(!p.IsLocked,
					html.Img(
						html.Class("icon"),
						html.Src("/images/edit.svg"),
						hx.Get("/device/"+p.ID+"/edit-name"),
						hx.Target("#"+p.ID+"-edit-name"),
						hx.Swap("innerHTML"),
					),
				),
			),
			html.Div(html.Class("row-header"), html.Span(Text("Target"))),
			html.Div(html.Span(Text(p.Targets[p.Target].FullName))),
			html.Div(html.Class("row-header"), html.Span(Text("Web Server Port"))),
			html.Div(
				If(p.Port == "", html.Span(html.Class("italic"), Text("not running"))),
				If(p.Port != "", html.Span(Text(p.Port))),
			),
			html.Div(html.Class("row-header"), html.Span(Text("Uptime"))),
			html.Div(
				Attr("id", p.ID+"-uptime"),
				hx.Post("/device/"+p.ID+"/get-uptime"),
				hx.Trigger("load"),
				hx.Swap("none"),
			),
			html.Div(html.Class("row-header"), html.Span(Text("Model Package"))),
			html.Div(html.Span(Text(p.Package))),
		),
		// Buttons and extra info
		// ...
	)
}
