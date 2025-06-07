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

// ButtonSave renders the Save button with conditional HTMX attributes.
func ButtonSave(uniq string, isDirty, saveToClipboard bool) Node {
	return html.Div(
		Attr("id", uniq),
		If(isDirty, html.Button(
			html.Class("flex flex-row items-center ml-4 px-3 py-1 rounded bg-blue-500 text-white hover:bg-blue-600"),
			If(saveToClipboard,
				Group{hx.Get("/save-modal"), hx.Swap("beforeend")},
			),
			If(!saveToClipboard,
				Group{hx.Get("/save"), hx.Swap("none")},
			),
			hx.Target("body"),
			html.Img(
				html.Class("w-6 h-6 mr-2"),
				html.Title("Save device changes"),
				html.Src("/images/save.svg"),
			),
			Text("Save"),
		)),
	)
}

// ButtonInfo renders the info button as an image with HTMX attributes.
func ButtonInfo(name, model, id string) Node {
	return html.Img(
		html.Class("w-6 h-6 mx-1 cursor-pointer"),
		html.Title("Show "+name+" info"),
		html.Src("/model/"+model+"/images/info.svg"),
		hx.Get("/device/"+id+"/show-view?view=info"),
	)
}

// ButtonSettings renders the settings button as an image with HTMX attributes, only if not isRoot.
func ButtonSettings(model, id string, isRoot bool) Node {
	return If(!isRoot, html.Img(
		html.Class("w-6 h-6 mx-1 cursor-pointer"),
		html.Title("Edit device settings"),
		html.Src("/model/"+model+"/images/settings.svg"),
		hx.Get("/device/"+id+"/show-view?view=settings"),
	))
}
