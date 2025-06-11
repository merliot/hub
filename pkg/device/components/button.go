package components

import (
	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

// ButtonNew renders the "+ New" button for adding a new device.
func ButtonNew(deviceName, deviceID string) Node {
	return Button(
		Title("Add a new device to "+deviceName),
		hx.Get("/new-modal/"+deviceID),
		hx.Target("body"),
		hx.Swap("beforeend"),
		Text("+ New"),
	)
}

// ButtonSave renders the Save button with conditional HTMX attributes.
func ButtonSave(uniq string, isDirty, saveToClipboard bool) Node {
	return Div(
		Attr("id", uniq),
		If(isDirty, Button(
			Class("flex flex-row items-center ml-4 px-3 py-1 rounded bg-blue-500 text-white hover:bg-blue-600"),
			If(saveToClipboard,
				Group{hx.Get("/save-modal"), hx.Swap("beforeend")},
			),
			If(!saveToClipboard,
				Group{hx.Get("/save"), hx.Swap("none")},
			),
			hx.Target("body"),
			Img(
				Class("w-6 h-6 mr-2"),
				Title("Save device changes"),
				Src("/images/save.svg"),
			),
			Text("Save"),
		)),
	)
}

// ButtonInfo renders the info button as an image with HTMX attributes.
func ButtonInfo(name, model, id string) Node {
	return Img(
		Class("w-6 h-6 mx-1 cursor-pointer"),
		Title("Show "+name+" info"),
		Src("/model/"+model+"/images/info.svg"),
		hx.Get("/device/"+id+"/show-view?view=info"),
	)
}

// ButtonSettings renders the settings button as an image with HTMX attributes, only if not isRoot.
func ButtonSettings(model, id string, isRoot bool) Node {
	return If(!isRoot, Img(
		Class("w-6 h-6 mx-1 cursor-pointer"),
		Title("Edit device settings"),
		Src("/model/"+model+"/images/settings.svg"),
		hx.Get("/device/"+id+"/show-view?view=settings"),
	))
}

// ButtonTrashcan renders the trash/delete button as an image with HTMX attributes.
func ButtonTrashcan(model, id string) Node {
	return Img(
		Class("w-6 h-6 mx-1 cursor-pointer"),
		Title("Delete device"),
		Src("/images/trash.svg"),
		hx.Delete("/device/"+id+"/destroy"),
		hx.Confirm("Are you sure you want to delete this device?"),
	)
}

// ButtonHammer renders the hammer/tool button as an image with HTMX attributes.
func ButtonHammer(model, id string) Node {
	return Img(
		Class("w-6 h-6 mx-1 cursor-pointer"),
		Title("Tool/Maintenance"),
		Src("/images/hammer.svg"),
		hx.Get("/device/"+id+"/show-view?view=tool"),
	)
}

// ButtonLocked renders the locked button as an image with HTMX attributes.
func ButtonLocked(model, id string) Node {
	return Img(
		Class("w-6 h-6 mx-1 cursor-pointer opacity-50"),
		Title("Device is locked"),
		Src("/images/locked.svg"),
	)
}
