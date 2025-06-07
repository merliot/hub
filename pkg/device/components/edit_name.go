package components

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func EditName(name string) Node {
	return Form(
		Class("p-4"),
		Label(Class("block mb-2 font-bold"), Text("Edit Name")),
		Input(
			Type("text"),
			Name("name"),
			Value(name),
			Class("border rounded p-2 mb-4"),
		),
		Button(
			Type("submit"),
			Class("px-4 py-2 bg-blue-500 text-white rounded"),
			Text("Save"),
		),
	)
}
