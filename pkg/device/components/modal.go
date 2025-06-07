package components

import (
	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

// ModalSave renders the save devices modal.
func ModalSave(devicesJSON string) Node {
	return Div(
		Class("fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 z-50"),
		Div(
			Class("bg-white rounded-2xl shadow-xl p-8 w-full max-w-lg"),
			H3(Class("text-xl font-bold mb-4"), Text("SAVE DEVICES")),
			P(Text("Devices were loaded from the DEVICES environment variable and changes to the devices must be saved back to DEVICES. Copy the JSON devices content below to your DEVICES environment variable.")),
			P(Text("If the hub is running on the cloud, update DEVICES environment variable on the hub service. This way, saved device changes are loaded on next reboot of the hub.")),
			Pre(
				ID("devices"),
				Class("overflow-y-auto text-green-600 bg-gray-100 rounded p-4 my-4"),
				Attr("style", "max-height: calc(60vh - 100px);"),
				Text(devicesJSON),
			),
			Form(
				Attr("action", "/devices"),
				Attr("method", "GET"),
				Div(
					Class("flex flex-row justify-end mt-8"),
					Button(
						Class("px-4 py-2 rounded bg-gray-300 hover:bg-gray-400 mr-2"),
						hx.Put("/nop"),
						hx.Target(".modal"),
						hx.Swap("delete"),
						Text("Close"),
					),
					Button(
						Class("px-4 py-2 rounded bg-blue-500 text-white hover:bg-blue-600 mr-2"),
						Type("button"),
						Attr("onclick", "copy2clipboard()"),
						Text("Copy to Clipboard"),
					),
					Button(
						Class("px-4 py-2 rounded bg-green-500 text-white hover:bg-green-600"),
						Type("submit"),
						Text("Save to File"),
					),
				),
			),
		),
		Raw(`
function copy2clipboard() {
	if (navigator.clipboard) {
		const content = document.getElementById('devices').innerText;
		navigator.clipboard.writeText(content);
	} else {
		alert("Browser blocking clipboard API access...insecure http:// connection?")
	}
}
		`),
	)
}

// ModelCollapsed renders the model-collapsed.tmpl fragment.
func ModelCollapsed(bgColor, fgColor, model string) Node {
	return Div(
		Class("flex flex-row p-4 min-w-[20rem] bg-"+bgColor+" text-"+fgColor+" border-"+fgColor+" justify-between border-solid border-2 rounded-3xl"),
		hx.Target("this"),
		hx.Swap("outerHTML"),
		Span(Class("text-lg font-bold ml-2.5 w-24"), Text(model)),
		Img(
			Class("w-6 h-6 mx-1 cursor-pointer"),
			Src("/images/expand.svg"),
			hx.Get("/model/"+model+"/model?view=expanded"),
		),
	)
}

// InstructionsMCPCollapsed renders the instructions-mcp-collapsed.tmpl fragment.
func InstructionsMCPCollapsed() Node {
	return Div(
		Class("flex flex-col items-center"),
		ID("mcp-instructions"),
		H3(
			Class("cursor-pointer"),
			hx.Get("/instructions-mcp?view=expanded"),
			hx.Target("#mcp-instructions"),
			hx.Swap("outerHTML"),
			Text("Instructions ▼"),
		),
	)
}

// ModalNew renders the create new device modal.
type ModelOption struct {
	Name    string
	Model   string
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
	return Div(
		Class("fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 z-50"),
		Div(
			Class("bg-white rounded-2xl shadow-xl p-8 w-full max-w-lg"),
			H3(Class("text-xl font-bold mb-4"), Text("CREATE A NEW DEVICE")),
			Form(
				hx.Post("/create"),
				hx.Target(".modal"),
				hx.Swap("delete"),
				Input(
					Type("hidden"),
					Name("ParentId"),
					Value(p.ParentID),
				),
				Div(
					Class("flex flex-col mb-5 w-40"),
					Label(Class("font-bold mb-2"), Text("Name the Device")),
					Input(
						Type("text"),
						Name("Child.Name"),
						Placeholder("Name"),
						Attr("maxlength", "20"),
						Required(),
					),
				),
				Div(
					Class("flex flex-col mb-5"),
					Label(Class("font-bold mb-2"), Text("Device ID")),
					Div(
						Class("flex flex-row"),
						Input(
							Disabled(),
							Type("text"),
							Placeholder("ID"),
							Value(p.NewID),
						),
						Input(
							Type("hidden"),
							Name("Child.Id"),
							Value(p.NewID),
						),
						Span(
							Class("w-6 h-6 mx-1"),
							Text("🔒"),
						),
					),
				),
				Div(
					Class("flex flex-col mb-5"),
					Text("Select a Model"),
					Div(
						Class("flex flex-col overflow-y-auto"),
						Map(p.Models, func(m ModelOption) Node {
							return Label(
								Class("flex flex-row cursor-pointer radio-container"),
								Input(
									Type("radio"),
									Name("Child.Model"),
									Value(m.Name),
									Required(),
								),
								Div(
									Class("radio-content p-0.5 border-solid border-2 rounded-2xl"),
									ModelCollapsed(m.BgColor, m.FgColor, m.Model),
								),
							)
						}),
					),
				),
				Div(
					Class("flex flex-row justify-between items-center"),
					Span(Class("text-red-500"), ID("error")),
					Div(
						Class("flex flex-row justify-end"),
						Button(
							Class("px-4 py-2 rounded bg-gray-300 hover:bg-gray-400 mr-2"),
							hx.Put("/nop"),
							hx.Target(".modal"),
							hx.Swap("delete"),
							Text("Close"),
						),
						If(p.IsLocked,
							Button(
								Class("px-4 py-2 rounded bg-blue-500 text-white hover:bg-blue-600"),
								Type("button"),
								Attr("onclick", "alert('Sorry, cannot create device: hub is locked')"),
								Text("Create"),
							),
						),
						If(!p.IsLocked,
							Button(
								Class("px-4 py-2 rounded bg-blue-500 text-white hover:bg-blue-600"),
								Type("submit"),
								Text("Create"),
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
	return Div(
		Class("fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 z-50"),
		Div(
			Class("bg-white rounded-2xl shadow-xl p-8 w-full max-w-lg"),
			H3(Class("text-xl font-bold mb-4"), Text("DOWNLOAD MCP SERVER")),
			P(Text("Download Model Context Protocol (MCP) server for "+p.Name+". The MCP server lets large language models (LLMs), such as Claude, interact with "+p.Name+". Basically, it means we can plug the physical world of devices into an LLM.")),
			P(Text("What could go wrong?")),
			Form(
				Attr("action", "/download-mcp-server"),
				Attr("method", "GET"),
				Class("mt-4"),
				Div(
					Class("mt-4"),
					Label(
						For("platform"),
						Class("block mb-2"),
						Text("SELECT YOUR PLATFORM"),
					),
					Select(
						Class("p-2 border rounded"),
						ID("platform"),
						Name("platform"),
						Required(),
						Map(p.Platforms, func(opt PlatformOption) Node {
							return SelectOption(opt.Os+"-"+opt.Arch, opt.Desc)
						}),
					),
				),
				Div(
					Class("flex flex-row justify-end mt-8"),
					Button(
						Class("px-4 py-2 rounded bg-gray-300 hover:bg-gray-400 mr-2"),
						hx.Put("/nop"),
						hx.Target(".modal"),
						hx.Swap("delete"),
						Text("Close"),
					),
					Button(
						Class("px-4 py-2 rounded bg-blue-500 text-white hover:bg-blue-600"),
						Type("submit"),
						Text("Download"),
					),
				),
			),
			InstructionsMCPCollapsed(),
		),
	)
}

// Option is a helper for select options.
func SelectOption(value, label string) Node {
	return Option(Value(value), Text(label))
}
