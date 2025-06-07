package components

import (
	. "maragu.dev/gomponents"
	html "maragu.dev/gomponents/html"
)

func Model(modelName string) Node {
	return html.Div(
		html.Class("p-4"),
		html.H3(html.Class("text-xl font-bold mb-4"), Text("Model")),
		html.P(Text(modelName)),
	)
}
