package relays

import (
	"fmt"
	"net/url"

	"github.com/merliot/hub/pkg/device"
	"github.com/merliot/hub/pkg/device/components"
	io "github.com/merliot/hub/pkg/io/relay"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type relays struct {
	Relays [4]io.Relay `schema:"desc=Relays"`
}

func (r *relays) Decode(values url.Values) error {

	// We shouldn't need a custom URL values decoder here, but tinygo's
	// reflect.ArrayOf() panics when trying to decode URL values with an
	// array using the form decorder.
	//
	// TODO this func can go away if reflect.ArrayOf() is implemented in
	// tinygo.

	if len(values) != 0 {
		for i := range r.Relays {
			relay := &r.Relays[i]
			nameKey := fmt.Sprintf("Relays[%d].Name", i)
			relay.Name = values.Get(nameKey)
			gpioKey := fmt.Sprintf("Relays[%d].Gpio", i)
			relay.Gpio = values.Get(gpioKey)
		}
	}
	return nil
}

type msgClick struct {
	Relay int `schema:"required,desc=Relay index"`
}

func (m msgClick) Desc() string {
	return "Click (toggle) the relay"
}

type msgClicked struct {
	Relay int
	State bool
}

func NewModel() device.Devicer {
	return &relays{}
}

func (r *relays) GetConfig() device.Config {
	return device.Config{
		Model:   "relays",
		Parents: []string{"hub"},
		State:   r,
		FS:      &fs,
		Targets: []string{"rpi", "nano-rp2040"},
		BgColor: "ice",
		FgColor: "black",
		PacketHandlers: device.PacketHandlers{
			"/click":  &device.PacketHandler[msgClick]{r.click},
			"clicked": &device.PacketHandler[msgClicked]{r.clicked},
		},
	}
}

func (r *relays) Setup() error {
	for i := range r.Relays {
		relay := &r.Relays[i]
		if err := relay.Setup(); err != nil {
			return err
		}
	}
	return nil
}

func (r *relays) click(pkt *device.Packet) {
	var click msgClick
	pkt.Unmarshal(&click)
	if click.Relay >= 0 && click.Relay < len(r.Relays) {
		relay := &r.Relays[click.Relay]
		relay.Set(!relay.State)
		var clicked = msgClicked{click.Relay, relay.State}
		pkt.SetPath("clicked").Marshal(&clicked).BroadcastUp()
	}
}

func (r *relays) clicked(pkt *device.Packet) {
	var clicked msgClicked
	pkt.Unmarshal(&clicked)
	relay := &r.Relays[clicked.Relay]
	relay.Set(clicked.State)
	pkt.BroadcastUp()
}

func (r *relays) Poll(pkt *device.Packet)     {}
func (r *relays) DemoSetup() error            { return r.Setup() }
func (r *relays) DemoPoll(pkt *device.Packet) { r.Poll(pkt) }

// Implement Detail() and Overview() for relays
func (r *relays) Detail() Node {
	added := false
	rows := []Node{}
	for i, relay := range r.Relays {
		if relay.Gpio != "" && relay.Name != "" {
			added = true
			rows = append(rows, Div(
				Class("flex flex-row items-center cursor-pointer"),
				Attr("hx-post", "/device/relays/click?Relay="+itoa(i)),
				Attr("hx-swap", "none"),
				Span(Class("w-16 text-right mx-6"), Text(relay.Name)),
				Img(Class("h-10"), Src("/model/relays/images/relay-"+stateToString(relay.State)+".png")),
				Span(Class("mx-6 px-2.5 text-sm bg-amber-500 text-white rounded"), Text(relay.Gpio)),
			))
		}
	}
	if !added {
		return components.UndefinedDetail("Relays")
	}
	return Div(
		Class("flex flex-col items-center justify-center"),
		Attr("id", "relays-detail"),
		Group(rows),
	)
}

func (r *relays) Overview() Node {
	added := false
	cols := []Node{}
	for _, relay := range r.Relays {
		if relay.Gpio != "" && relay.Name != "" {
			added = true
			cols = append(cols, Div(
				Class("flex flex-col items-center mx-1"),
				Span(Class("text-sm"), Text(relay.Name)),
				Img(Class("h-6"), Src("/model/relays/images/relay-"+stateToString(relay.State)+".png")),
			))
		}
	}
	if !added {
		return components.UndefinedOverview()
	}
	return Div(
		Class("flex flex-row items-center justify-evenly"),
		Attr("id", "relays-overview"),
		Group(cols),
	)
}

// Helper to convert relay state to string for image path
func stateToString(state bool) string {
	if state {
		return "on"
	}
	return "off"
}

// Helper to convert int to string
func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
