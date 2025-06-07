package components

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func Model(modelName string) Node {
	return Div(
		Class("p-4"),
		H3(Class("text-xl font-bold mb-4"), Text("Model")),
		P(Text(modelName)),
	)
}
