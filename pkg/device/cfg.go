package device

import (
	"embed"
	"encoding/json"
	"net/http"
	"time"
)

// APIs are custom APIs for the device
type APIs map[string]http.HandlerFunc // key: path

// Config is the device model configuration
type Config struct {
	// Model is the device model name, e.g. "relays"
	Model string

	// Parents are the supported parent models, e.g. ["hub"]
	Parents []string

	// Device flags see FlagXxxx
	Flags flags

	// The device state, usually set to the containing device struct instance
	State any `json:"-"`

	// The device's embedded static file system
	FS *embed.FS `json:"-"`

	// Targets support by device, e.g. ["rpi", "nano-rp2040"]
	Targets []string

	// PollPeriod is the device polling period. If not specified (zero),
	// polling is effectively disabled by setting to math.MaxInt64. If
	// specified, the range is [10ms..forever).  Values less than 10ms
	// are forced to 10ms.
	PollPeriod time.Duration

	// PacketHandlers is a custom map of packet handlers.
	//
	// e.g.
	//
	// PacketHandlers: PacketHandlers{
	//	"/click":   &PacketHandler[msgClick]{r.click},
	//	"clicked": &PacketHandler[msgClicked]{r.clicked},
	// }
	//
	// The map key is a path (e.g. "/click").  A http handler is installed
	// to take POST requests at the full path /device/{id}/{path}.  An http
	// POST request at a full path results in callback into the device
	// (e.g.  r.click).  The callback is passed a *Packet.  The *Packet
	// contains a JSON encoded message based of the type of the packet
	// handler.  Each packet handler is typed (e.g. msgClick).  The message
	// type struct exports fields that are filled from the http request
	// query parameters.
	//
	// e.g.
	//
	// A POST request to /device/foo/click?Relay=2 would map to the message:
	//
	// type msgClick struct {
	//         Relay int
	// }
	//
	// And the callback can get access to the message from the *Packet:
	//
	// func (r *relays) click(p *Packet) {
	//         var click msgClick
	//         p.Unmarshal(&click)
	//
	//         // process click msg
	//
	//         relay := &r.Relays[click.Relay]
	//         relay.Set(!relay.State)
	//
	//         // broadcast clicked notification msg up
	//
	//         var clicked = msgClicked{click.Relay, relay.State}
	//         p.SetPath("clicked").Marshal(&clicked).BroadcastUp()
	// }
	//
	// Additionally, the packet handlers are use to handle packets arriving
	// on downlinks from other devices in the device tree.  The handler can
	// process the packet to update state and route the packet up the
	// device tree.
	//
	// e.g.
	//
	// type msgClicked struct {
	//         Relay int
	//         State bool
	// }
	//
	// func (r *relays) clicked(p *Packet) {
	//         var clicked msgClicked
	//         pkt.Unmarshal(&clicked)
	//         relay := &r.Relays[clicked.Relay]
	//         relay.Set(clicked.State)
	//         pkt.BroadcastUp()
	// }
	PacketHandlers `json:"-"`

	// APIs are custom APIs for the device.
	//
	// e.g.
	//
	// APIs: device.APIs{
	//         "POST /generate":    q.generate,
	//         "GET /edit-content": q.editContent,
	// }
	//
	// The custom APIs are available at /device/{id}/xxx.  For example, a
	// http request to POST /device/foo/generate would call q.generate, an
	// http.HandlerFunc.
	APIs `json:"-"`

	// FuncMap are custom template functions for the device.  Most devices
	// will not have any custom functions.
	//
	// e.g.
	//
	// FuncMap: device.FuncMap{
	//         "png": q.png,
	// },
	//
	// Device templates call the function using {{fn}} syntax.  For example:
	//
	// <div class="m-8">
	//     <img src="{{png state.Content -5}}">
	// </div>
	FuncMap `json:"-"`

	// BgColor is the device background color
	BgColor string

	// FgColor is the device forground (text, border) color
	FgColor string

	// InitialView is the initial display view mode when device is first
	// displayed.  Value can be "overview" (default) or "detail".
	InitialView string
}

func (c Config) getPrettyJSON() []byte {
	json, err := json.MarshalIndent(&c, "", "\t")
	if err != nil {
		return []byte(err.Error())
	}
	return json
}
