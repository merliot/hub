package components

import (
	. "maragu.dev/gomponents"
	html "maragu.dev/gomponents/html"
)

func EditName(name string) Node {
	return html.Form(
		html.Class("p-4"),
		html.Label(html.Class("block mb-2 font-bold"), Text("Edit Name")),
		html.Input(
			html.Type("text"),
			html.Name("name"),
			html.Value(name),
			html.Class("border rounded p-2 mb-4"),
		),
		html.Button(
			html.Type("submit"),
			html.Class("px-4 py-2 bg-blue-500 text-white rounded"),
			Text("Save"),
		),
	)
}
