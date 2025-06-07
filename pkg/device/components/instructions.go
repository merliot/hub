package components

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func Instructions(content string) Node {
	return Div(
		Class("p-4"),
		H3(Class("text-xl font-bold mb-4"), Text("Instructions")),
		Pre(Text(content)),
	)
}
