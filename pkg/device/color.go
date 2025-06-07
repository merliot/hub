//go:build !tinygo

package device

func (d *device) bgColor() string {
	if d.Config.BgColor == "" {
		return "bg-white"
	}
	switch d.Config.BgColor {
	case "sunflower":
		return "bg-yellow-300"
	case "gold":
		return "bg-yellow-400"
	case "text", "violet-creme":
		return "bg-purple-200"
	default:
		return "bg-" + d.Config.BgColor
	}
}

func (d *device) textColor() string {
	if d.Config.FgColor == "" {
		return "text-black"
	}
	switch d.Config.FgColor {
	case "text", "violet-creme":
		return "text-purple-200"
	default:
		return "text-" + d.Config.FgColor
	}
}

func (d *device) borderColor() string {
	if d.Config.BgColor == "" {
		return "border-white"
	}
	switch d.Config.BgColor {
	case "sunflower":
		return "border-yellow-300"
	case "gold":
		return "border-yellow-400"
	case "text", "violet-creme":
		return "border-purple-200"
	default:
		return "border-" + d.Config.BgColor
	}
}
