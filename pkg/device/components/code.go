package components

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func CodeList(names []string) Node {
	return Div(
		Class("p-4"),
		H3(Class("text-xl font-bold mb-4"), Text("Code Files")),
		Ul(
			Group(Map(names, func(name string) Node {
				return Li(Class("mb-1"), Text(name))
			})),
		),
	)
}
