package components

import (
	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
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
	return Div(
		Class("model-"+p.Model+" "+p.ClassOffline),
		Attr("id", p.ID),
		hx.Target("this"),
		hx.Swap("outerHTML"),
		Div(
			Class("flex flex-row ml-"+itoa(p.Level*10)),
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
					Class("m-2.5 max-w-3xl"),
					DeviceInfoBody(p),
				),
			),
		),
	)
	// RenderChildren would go here if needed
}

func DeviceInfoBody(p DeviceInfoParams) Node {
	return Div(
		Class("flex flex-col"),
		Div(
			Class("flex flex-row justify-center"),
			H3(Text("Info")),
		),
		Raw(`<style>.info-grid { display: grid; grid-template-columns: auto auto; column-gap: 20px; row-gap: 10px; align-items: center; } .info-grid .row-header { font-weight: bold; text-align: right; } .hr-row { grid-column: span 2; } hr { width: 100%; }</style>`),
		Div(
			Class("info-grid"),
			Div(Class("row-header"), Span(Text("ID"))),
			Div(Span(Text(p.ID))),
			Div(Class("row-header"), Span(Text("Model"))),
			Div(Span(Text(p.Model))),
			Div(Class("row-header"), Span(Text("Name"))),
			Div(
				Class("flex flex-row"),
				Attr("id", p.ID+"-edit-name"),
				Span(Class("mr-4"), Text(p.Name)),
				If(p.IsLocked,
					Img(
						Src("/images/edit.svg"),
						Attr("onclick", "alert('Sorry, cannot change name: device is locked')"),
					),
				),
				If(!p.IsLocked,
					Img(
						Class("icon"),
						Src("/images/edit.svg"),
						hx.Get("/device/"+p.ID+"/edit-name"),
						hx.Target("#"+p.ID+"-edit-name"),
						hx.Swap("innerHTML"),
					),
				),
			),
			Div(Class("row-header"), Span(Text("Target"))),
			Div(Span(Text(p.Targets[p.Target].FullName))),
			Div(Class("row-header"), Span(Text("Web Server Port"))),
			Div(
				If(p.Port == "", Span(Class("italic"), Text("not running"))),
				If(p.Port != "", Span(Text(p.Port))),
			),
			Div(Class("row-header"), Span(Text("Uptime"))),
			Div(
				Attr("id", p.ID+"-uptime"),
				hx.Post("/device/"+p.ID+"/get-uptime"),
				hx.Trigger("load"),
				hx.Swap("none"),
			),
			Div(Class("row-header"), Span(Text("Model Package"))),
			Div(Span(Text(p.Package))),
		),
		// Buttons and extra info
		// ...
	)
}
