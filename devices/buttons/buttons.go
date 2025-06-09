package buttons

import (
	"fmt"
	"net/url"
	"time"

	"github.com/merliot/hub/pkg/device"
	"github.com/merliot/hub/pkg/device/components"
	io "github.com/merliot/hub/pkg/io/button"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type buttons struct {
	Buttons [4]io.Button `schema:"desc=Buttons"`
}

func (b *buttons) Decode(values url.Values) error {
	// Custom decoder since tinygo's reflect.ArrayOf() panics
	if len(values) != 0 {
		for i := range b.Buttons {
			button := &b.Buttons[i]
			nameKey := fmt.Sprintf("Buttons[%d].Name", i)
			button.Name = values.Get(nameKey)
			gpioKey := fmt.Sprintf("Buttons[%d].Gpio", i)
			button.Gpio = values.Get(gpioKey)
		}
	}
	return nil
}

type msgUpdate struct {
	Button int
	State  bool
}

func NewModel() device.Devicer {
	return &buttons{}
}

func (b *buttons) GetConfig() device.Config {
	return device.Config{
		Model:      "buttons",
		Parents:    []string{"hub"},
		State:      b,
		FS:         &fs,
		Targets:    []string{"rpi", "nano-rp2040"},
		PollPeriod: 10 * time.Millisecond,
		BgColor:    "lilac",
		FgColor:    "black",
		PacketHandlers: device.PacketHandlers{
			"update": &device.PacketHandler[msgUpdate]{b.update},
		},
	}
}

func (b *buttons) Setup() error {
	for i := range b.Buttons {
		button := &b.Buttons[i]
		if err := button.Setup(); err != nil {
			return err
		}
	}
	return nil
}

func (b *buttons) Poll(pkt *device.Packet) {
	for i := range b.Buttons {
		button := &b.Buttons[i]
		last := button.State
		curr := button.Get()
		if curr != last {
			var update = msgUpdate{i, curr}
			pkt.SetPath("update").Marshal(&update).BroadcastUp()
		}
	}
}

func (b *buttons) update(pkt *device.Packet) {
	var update msgUpdate
	pkt.Unmarshal(&update)
	button := &b.Buttons[update.Button]
	button.State = update.State
	pkt.BroadcastUp()
}

func (b *buttons) Detail() Node {
	added := false
	rows := []Node{}
	for _, button := range b.Buttons {
		if button.Gpio != "" && button.Name != "" {
			added = true
			rows = append(rows, Div(
				Class("flex flex-row items-center"),
				Span(Class("w-16 text-right mx-6"), Text(button.Name)),
				Img(Class("h-10"), Src("/model/buttons/images/button-"+stateToString(button.State)+".png")),
				Span(Class("mx-6 px-2.5 text-sm bg-amber-500 text-white rounded"), Text(button.Gpio)),
			))
		}
	}
	if !added {
		return components.UndefinedDetail("Buttons")
	}
	return Div(
		Class("flex flex-col items-center justify-center"),
		Attr("id", "buttons-detail"),
		Group(rows),
	)
}

func (b *buttons) Overview() Node {
	added := false
	cols := []Node{}
	for _, button := range b.Buttons {
		if button.Gpio != "" && button.Name != "" {
			added = true
			cols = append(cols, Div(
				Class("flex flex-col items-center mx-1"),
				Span(Class("text-sm"), Text(button.Name)),
				Img(Class("h-6"), Src("/model/buttons/images/button-"+stateToString(button.State)+".png")),
			))
		}
	}
	if !added {
		return components.UndefinedOverview()
	}
	return Div(
		Class("flex flex-row items-center justify-evenly"),
		Attr("id", "buttons-overview"),
		Group(cols),
	)
}

func stateToString(state bool) string {
	if state {
		return "on"
	}
	return "off"
}
