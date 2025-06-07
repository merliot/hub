package components

import (
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type SessionInfo struct {
	ID        string
	Age       string
	LastViews []SessionViewInfo
}

type SessionViewInfo struct {
	DeviceID string
	View     string
	Level    int
}

type SessionsPageParams struct {
	Sessions []SessionInfo
}

// SessionsPage renders the sessions.tmpl view.
func SessionsPage(p SessionsPageParams) Node {
	return El("html",
		Lang("en"),
		Head(
			Meta(Name("viewport"), Content("width=device-width, initial-scale=1")),
			Link(Rel("icon"), Type("image/png"), Attr("sizes", "32x32"), Href("/images/favicon-32x32.png")),
			Link(Rel("icon"), Type("image/png"), Attr("sizes", "16x16"), Href("/images/favicon-16x16.png")),
			Title("Sessions"),
			Link(Rel("stylesheet"), Href("https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css")),
			Meta(Attr("http-equiv", "refresh"), Content("2")),
		),
		Body(
			Class("bg-black text-purple-200 m-4"),
			H2(Text("Active Sessions")),
			If(len(p.Sessions) == 0,
				P(Text("No active sessions")),
			),
			Group(
				Map(p.Sessions, func(s SessionInfo) Node {
					return Div(
						Class("bg-yellow-300 text-black m-1 p-2 rounded-3xl max-w-lg"),
						H3(Text("Session ID: "+s.ID)),
						P(Text("Last Update: "+s.Age)),
						H4(Text("Last Views:")),
						Ul(
							Group(
								Map(s.LastViews, func(v SessionViewInfo) Node {
									return Li(Text(v.DeviceID + ": " + v.View + ", " + itoa(v.Level)))
								}),
							),
						),
					)
				}),
			),
		),
	)
}

type SessionViewParams struct {
	SessionID  string
	PingPeriod int
	Body       Node // Rendered view content
}

// SessionView renders the session.tmpl/session-view.tmpl view.
func SessionView(p SessionViewParams) Node {
	return Div(
		Class("offline"),
		Attr("id", "session"),
		Attr("hx-headers", `{"session-id": "`+p.SessionID+`"}`),
		Attr("hx-trigger", "every "+itoa(p.PingPeriod)+"s"),
		Attr("ws-send", ""),
		Attr("hx-ext", "ws"),
		Attr("ws-connect", "/wsx?session-id="+p.SessionID),
		p.Body,
	)
}

// itoa is a helper to convert int to string.
func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
