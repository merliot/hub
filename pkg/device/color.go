//go:build !tinygo

package device

func (d *device) bgColor() string {
	if d.Config.BgColor == "" {
		return "bg-space-white"
	}
	return "bg-" + d.Config.BgColor
}

func (d *device) textColor() string {
	if d.Config.FgColor == "" {
		return "text-black"
	}
	return "text-" + d.Config.FgColor
}

func (d *device) borderColor() string {
	if d.Config.BgColor == "" {
		return "border-space-white"
	}
	return "border-" + d.Config.BgColor
}
