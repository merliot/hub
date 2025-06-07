package components

import (
	. "maragu.dev/gomponents"
	html "maragu.dev/gomponents/html"
)

func Instructions(content string) Node {
	return html.Div(
		html.Class("p-4"),
		html.H3(html.Class("text-xl font-bold mb-4"), Text("Instructions")),
		html.Pre(Text(content)),
	)
}
