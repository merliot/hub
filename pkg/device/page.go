//go:build !tinygo

package device

type page struct {
	// Name of page URL to navigate to.  If Url == "" and Name == "",
	// name means tab heading.
	Name  string
	Label string
	Url   string
}

var homePages = []page{
	{"", "WELCOME", ""},
	{"intro", "INTRODUCTION", "/home/intro"},
	{"targets", "TARGETS", "/home/targets"},
	{"", "SOURCE", "https://github.com/merliot/hub"},
	{"contact", "CONTACT", "/home/contact"},
}

var demoPages = []page{
	{"", "DEMO", ""},
	{"devices", "DEVICE VIEW", "/demo/devices"},
	//{"network", "NETWORK VIEW", "/demo/network"},
	{"about-demo", "ABOUT", "/demo/about-demo"},
}

var docPages = []page{
	{"", "GUIDES", ""},
	{"quick-start", "QUICK START", "/doc/quick-start"},
	{"install", "INSTALL GUIDE", "/doc/install"},
	{"devices", "DEVICES", "/doc/devices"},
	{"env-vars", "ENV VARS", "/doc/env-vars"},
	{"mcp-server", "MCP SERVER", "/doc/mcp-server"},
	{"privacy", "PRIVACY", "/doc/privacy"},
	{"faq", "FAQ", "/doc/faq"},

	{"", "DEVELOPER", ""},
	{"", "REFERENCE", "https://pkg.go.dev/github.com/merliot/hub"},
	{"run-source", "RUN SOURCE", "/doc/run-source"},
	{"new-model", "NEW MODEL", "/doc/new-model"},
	{"services", "SERVICES", "/doc/services"},
	{"api", "API", "/doc/api"},
	{"mcp", "MCP", "/doc/mcp"},
	{"ui-device", "UI TO DEVICE", "/doc/ui-device"},
	{"device-ui", "DEVICE TO UI", "/doc/device-ui"},
	{"device-views", "DEVICE VIEWS", "/doc/device-views"},
	{"template-funcs", "TEMPLATE FUNCS", "/doc/template-funcs"},
	{"template-map", "TEMPLATE MAP", "/doc/template-map"},
}
