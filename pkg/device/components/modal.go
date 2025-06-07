package components

import (
	. "maragu.dev/gomponents"
	html "maragu.dev/gomponents/html"
	hx "maragu.dev/gomponents-htmx"
)

// ModalSave renders the save devices modal.
func ModalSave(devicesJSON string) Node {
	return html.Div(
		html.Class("fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 z-50"),
		html.Div(
			html.Class("bg-white rounded-2xl shadow-xl p-8 w-full max-w-lg"),
			html.H3(html.Class("text-xl font-bold mb-4"), Text("SAVE DEVICES")),
			html.P(Text("Devices were loaded from the DEVICES environment variable and changes to the devices must be saved back to DEVICES. Copy the JSON devices content below to your DEVICES environment variable.")),
			html.P(Text("If the hub is running on the cloud, update DEVICES environment variable on the hub service. This way, saved device changes are loaded on next reboot of the hub.")),
			html.Pre(
				html.ID("devices"),
				html.Class("overflow-y-auto text-green-600 bg-gray-100 rounded p-4 my-4"),
				Attr("style", "max-height: calc(60vh - 100px);"),
				Text(devicesJSON),
			),
			html.Form(
				html.Attr("action", "/devices"),
				html.Method("GET"),
				html.Div(
					html.Class("flex flex-row justify-end mt-8"),
					html.Button(
						html.Class("px-4 py-2 rounded bg-gray-300 hover:bg-gray-400 mr-2"),
						hx.Put("/nop"),
						hx.Target(".modal"),
						hx.Swap("delete"),
						Text("Close"),
					),
					html.Button(
						html.Class("px-4 py-2 rounded bg-blue-500 text-white hover:bg-blue-600 mr-2"),
						html.Type("button"),
						Attr("onclick", "copy2clipboard()"),
						Text("Copy to Clipboard"),
					),
					html.Button(
						html.Class("px-4 py-2 rounded bg-green-500 text-white hover:bg-green-600"),
						html.Type("submit"),
						Text("Save to File"),
					),
				),
			),
		),
		Script(
			Raw(`function copy2clipboard() {
	if (navigator.clipboard) {
		const content = document.getElementById('devices').innerText;
		navigator.clipboard.writeText(content);
	} else {
		alert("Browser blocking clipboard API access...insecure http:// connection?")
	}
}`),
		),
	)
}

// ModelCollapsed renders the model-collapsed.tmpl fragment.
func ModelCollapsed(bgColor, fgColor, model string) Node {
	return html.Div(
		html.Class("flex flex-row p-4 min-w-[20rem] bg-"+bgColor+" text-"+fgColor+" border-"+fgColor+" justify-between border-solid border-2 rounded-3xl"),
		hx.Target("this"),
		hx.Swap("outerHTML"),
		html.Span(html.Class("text-lg font-bold ml-2.5 w-24"), Text(model)),
		html.Img(
			html.Class("w-6 h-6 mx-1 cursor-pointer"),
			html.Src("/images/expand.svg"),
			hx.Get("/model/"+model+"/model?view=expanded"),
		),
	)
}

// InstructionsMCPCollapsed renders the instructions-mcp-collapsed.tmpl fragment.
func InstructionsMCPCollapsed() Node {
	return html.Div(
		html.Class("flex flex-col items-center"),
		html.ID("mcp-instructions"),
		html.H3(
			html.Class("cursor-pointer"),
			hx.Get("/instructions-mcp?view=expanded"),
			hx.Target("#mcp-instructions"),
			hx.Swap("outerHTML"),
			Text("Instructions ▼"),
		),
	)
}

// ModalNew renders the create new device modal.
type ModelOption struct {
	Name   string
	Model  string
	BgColor string
	FgColor string
}

type ModalNewParams struct {
	ParentID string
	NewID    string
	Models   []ModelOption
	IsLocked bool
}

func ModalNew(p ModalNewParams) Node {
	return html.Div(
		html.Class("fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 z-50"),
		html.Div(
			html.Class("bg-white rounded-2xl shadow-xl p-8 w-full max-w-lg"),
			html.H3(html.Class("text-xl font-bold mb-4"), Text("CREATE A NEW DEVICE")),
			html.Form(
				hx.Post("/create"),
				hx.Target(".modal"),
				hx.Swap("delete"),
				html.Input(
					html.Type("hidden"),
					html.Name("ParentId"),
					html.Value(p.ParentID),
				),
				html.Div(
					html.Class("flex flex-col mb-5 w-40"),
					html.Label(html.Class("font-bold mb-2"), Text("Name the Device")),
					html.Input(
						html.Type("text"),
						html.Name("Child.Name"),
						html.Placeholder("Name"),
						Attr("maxlength", "20"),
						html.Required(),
					),
				),
				html.Div(
					html.Class("flex flex-col mb-5"),
					html.Label(html.Class("font-bold mb-2"), Text("Device ID")),
					html.Div(
						html.Class("flex flex-row"),
						html.Input(
							html.Disabled(),
							html.Type("text"),
							html.Placeholder("ID"),
							html.Value(p.NewID),
						),
						html.Input(
							html.Type("hidden"),
							html.Name("Child.Id"),
							html.Value(p.NewID),
						),
						html.Span(
							html.Class("w-6 h-6 mx-1"),
							Text("🔒"),
						),
					),
				),
				html.Div(
					html.Class("flex flex-col mb-5"),
					Text("Select a Model"),
					html.Div(
						html.Class("flex flex-col overflow-y-auto"),
						Group(
							Map(p.Models, func(m ModelOption) Node {
								return html.Label(
									html.Class("flex flex-row cursor-pointer radio-container"),
									html.Input(
										html.Type("radio"),
										html.Name("Child.Model"),
										html.Value(m.Name),
										html.Required(),
									),
									html.Div(
										html.Class("radio-content p-0.5 border-solid border-2 rounded-2xl"),
										ModelCollapsed(m.BgColor, m.FgColor, m.Model),
									),
								)
							}),
						),
					),
				),
				html.Div(
					html.Class("flex flex-row justify-between items-center"),
					html.Span(html.Class("text-red-500"), html.ID("error")),
					html.Div(
						html.Class("flex flex-row justify-end"),
						html.Button(
							html.Class("px-4 py-2 rounded bg-gray-300 hover:bg-gray-400 mr-2"),
							hx.Put("/nop"),
							hx.Target(".modal"),
							hx.Swap("delete"),
							Text("Close"),
						),
						If(p.IsLocked,
							html.Button(
								html.Class("px-4 py-2 rounded bg-blue-500 text-white hover:bg-blue-600"),
								html.Type("button"),
								Attr("onclick", "alert('Sorry, cannot create device: hub is locked')"),
								Text("Create"),
							),
							html.Button(
								html.Class("px-4 py-2 rounded bg-blue-500 text-white hover:bg-blue-600"),
								html.Type("submit"),
								Text("Create"),
							),
						),
					),
				),
			),
		),
		),
	)
}

// ModalMCP renders the download MCP server modal.
type PlatformOption struct {
	Os   string
	Arch string
	Desc string
}

type ModalMCPParams struct {
	Name      string
	Platforms []PlatformOption
}

func ModalMCP(p ModalMCPParams) Node {
	return html.Div(
		html.Class("fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 z-50"),
		html.Div(
			html.Class("bg-white rounded-2xl shadow-xl p-8 w-full max-w-lg"),
			html.H3(html.Class("text-xl font-bold mb-4"), Text("DOWNLOAD MCP SERVER")),
			html.P(Text("Download Model Context Protocol (MCP) server for "+p.Name+". The MCP server lets large language models (LLMs), such as Claude, interact with "+p.Name+". Basically, it means we can plug the physical world of devices into an LLM.")),
			html.P(Text("What could go wrong?")),
			html.Form(
				html.Attr("action", "/download-mcp-server"),
				html.Method("GET"),
				html.Class("mt-4"),
				html.Div(
					html.Class("mt-4"),
					html.Label(
						html.For("platform"),
						html.Class("block mb-2"),
						Text("SELECT YOUR PLATFORM"),
					),
					html.Select(
						html.Class("p-2 border rounded"),
						html.ID("platform"),
						html.Name("platform"),
						html.Required(),
						Group(
							Option("", "-- Select Platform --"),
							Map(p.Platforms, func(opt PlatformOption) Node {
								return Option(opt.Os+"-"+opt.Arch, opt.Desc)
							}),
						),
					),
				),
				html.Div(
					html.Class("flex flex-row justify-end mt-8"),
					html.Button(
						html.Class("px-4 py-2 rounded bg-gray-300 hover:bg-gray-400 mr-2"),
						hx.Put("/nop"),
						hx.Target(".modal"),
						hx.Swap("delete"),
						Text("Close"),
					),
					html.Button(
						html.Class("px-4 py-2 rounded bg-blue-500 text-white hover:bg-blue-600"),
						html.Type("submit"),
						Text("Download"),
					),
				),
			),
			InstructionsMCPCollapsed(),
		),
	)
}

// Option is a helper for select options.
func Option(value, label string) Node {
	return html.Option(html.Value(value), Text(label))
} 