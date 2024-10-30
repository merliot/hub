package hub

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// NoMsg is used as an empty message type when creating a Handle.
type NoMsg struct{}

// Packet is the basic container for messages sent between devices.
type Packet struct {
	// SessionId is the originating session id.  Empty means packet isn't
	// pinned to any session.
	SessionId string
	// Dst is the device id of the destination device
	Dst string
	// Path identifies the message content.  Path format is same as
	// url.URL.Path, with the leading slash.  e.g. /takeone.
	Path string
	// Msg is the packet payload.  Use NoMsg for no message.
	Msg json.RawMessage
}

func newPacketFromURL(url *url.URL, v any) (*Packet, error) {
	var pkt = &Packet{
		Path: url.Path,
	}
	if _, ok := v.(*NoMsg); ok {
		return pkt, nil
	}
	err := decode(v, url.Query())
	if err != nil {
		return nil, err
	}
	pkt.Msg, err = json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return pkt, nil
}

// String returns packet as string in format "[dst id/path] msg"
func (p *Packet) String() string {
	var msg any
	json.Unmarshal(p.Msg, &msg)
	if p.SessionId == "" {
		return fmt.Sprintf("[%s%s] %v", p.Dst, p.Path, msg)
	} else {
		return fmt.Sprintf("[%s%s*] %v", p.Dst, p.Path, msg)
	}
}

// Marshal the packet message payload as JSON from v
func (p *Packet) Marshal(v any) *Packet {
	var err error
	p.Msg, err = json.Marshal(v)
	if err != nil {
		fmt.Printf("JSON marshal error %s\r\n", err.Error())
	}
	return p
}

// Unmarshal the packet message payload as JSON into v
func (p *Packet) Unmarshal(v any) *Packet {
	if len(p.Msg) > 0 {
		if err := json.Unmarshal(p.Msg, v); err != nil {
			fmt.Printf("JSON unmarshal error %s\r\n", err.Error())
		}
	}
	return p
}

// SetSession pins the packet to session
func (p *Packet) SetSession(sessionId string) *Packet {
	p.SessionId = sessionId
	return p
}

// SetDst sets the packet destination, a device id
func (p *Packet) SetDst(dst string) *Packet {
	p.Dst = dst
	return p
}

// SetPath sets the packet path
func (p *Packet) SetPath(path string) *Packet {
	p.Path = path
	return p
}

// RouteDown routes the packet down to a downlink.  Which downlink is
// determined by a lookup in the routing table for the "next-hop" downlink, the
// downlink which is towards the destination.
func (p *Packet) RouteDown() {
	LogInfo("RouteDown", "pkt", p)
	downlinksRoute(p)
}

// RouteUp routes the packet up to:
//
//  1. Each listening session, where a session is an http(s) client (browser,
//     etc) that has also opened, and is listening on, a websocket at /wsx.
//
//     The packet is transformed into an html snippet before being sent on the
//     websocket to the client (see htmx, websockets).  The packet path and the
//     current session's view name the html template used for the
//     transformation.  The template name is in the format:
//
//     {path}-{view}.tmpl
//
//     For example, consider routing the packet with the message:
//
//     var msg = MsgClicked{Relay: 2, State: true}
//     pkt.SetPath("/clicked").Marshal(&msg).RouteUp()
//
//     And say the current view is "overview".  The template name is:
//
//     clicked-overview.tmpl
//
//     The template is executed and the resulting html snippet is sent on the
//     websocket.  Per htmx, the html snippet is swap by DOM id, so using a
//     unique id in the template like:
//
//     <div id="{{uniq `relay`}}">
//     ...
//     </div>
//
//  2. Each active uplink the device is dialed into.  Each uplink is a
//     websocket connected on /ws.  The packet is JSON-encoded before sending on
//     the websocket, and JSON-decoded by the receiving uplink device.
func (p *Packet) RouteUp() {
	LogInfo("RouteUp", "pkt", p)
	sessionsRoute(p)
	uplinksRoute(p)
}
