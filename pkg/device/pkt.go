package device

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const maxLength = 100 // Max length of packet string

// NoMsg is an empty message type for PacketHandle's
type NoMsg struct{}

// Packet is the basic container for messages sent inner and between devices.
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
	// Stash server pointer
	*server
}

func (s *server) newPacket() *Packet {
	return &Packet{server: s}
}

func (s *server) newPacketFromRequest(r *http.Request, v any) (*Packet, error) {
	var pkt = &Packet{
		Path:      r.URL.Path,
		SessionId: r.Header.Get("session-id"),
		server:    s,
	}
	if _, ok := v.(*NoMsg); ok {
		return pkt, nil
	}
	r.ParseForm()
	err := decode(v, r.Form)
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

	// Convert msg to string and truncate if needed
	msgStr := fmt.Sprintf("%v", msg)
	if len(msgStr) > maxLength {
		msgStr = msgStr[:maxLength] + "..."
	}

	if p.SessionId == "" {
		return fmt.Sprintf("[%s%s*] %v", p.Dst, p.Path, msgStr)
	} else {
		return fmt.Sprintf("[%s%s] %v", p.Dst, p.Path, msgStr)
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

// ClearMsg empties the packet message
func (p *Packet) ClearMsg() *Packet {
	p.Msg, _ = json.Marshal(map[string]interface{}{})
	return p
}

// routeDown routes the packet down to a downlink.  Which downlink is
// determined by a lookup in the routing table for the "next-hop" downlink, the
// downlink which is towards the destination.
func (p *Packet) routeDown() error {
	LogDebug("routeDown", "pkt", p)

	s := p.server
	if s == nil {
		return fmt.Errorf("Packet.server not set")
	}

	d, ok := s.devices.get(p.Dst)
	if !ok {
		return deviceNotFound(p.Dst)
	}

	nexthop := d.nexthop
	if nexthop.isSet(flagMetal) {
		nexthop.handle(pkt)
	} else {
		s.downlinks.route(nexthop.Id, pkt)
	}

	return nil
}

func (p *Packet) handle() error {
	s := p.server
	if s == nil {
		return fmt.Errorf("Packet.server not set")
	}

	if pkt.Dst == "" {
		// Run server handler
		if handler, ok := s.packetHandlers[pkt.Path]; ok {
			LogDebug("Handling", "pkt", pkt)
			handler.cb(pkt)
		}
		return nil
	}

	d, ok := s.devices.get(pkt.Dst)
	if !ok {
		return deviceNotFound(pkt.Dst)
	}

	// Run device handler
	d.handle(pkt)
	return nil
}

// RouteUp routes the packet up to:
//
//  1. Each active uplink the device is dialed into.  Each uplink is a
//     websocket connected on /ws.  The packet is JSON-encoded before sending on
//     the websocket, and JSON-decoded by the receiving uplink device.
//
//  2. Sessions, where a session is an http(s) client (browser, etc) that has
//     also opened, and is listening on, a websocket at /wsx.
//
//     If SessionId is set on packet, the packet is routed to the session.  If
//     SessionId is not set on the packet, the packet is broadcast to all
//     sessions.
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
func (p *Packet) RouteUp() error {
	LogDebug("RouteUp", "pkt", p)

	s := p.server
	if s == nil {
		return fmt.Errorf("Packet.server not set")
	}

	s.uplinks.routeAll(p)

	err := s.sessions.routeAll(p)
	if err != nil {
		LogError("RouteUp", "err", err)
	}

	return err
}

// RouteUp is a device packet handler that routes the packet up
func RouteUp(p *Packet) {
	p.RouteUp()
}

func (p *Packet) BroadcastUp() {
	// Route to all sessions
	p.SetSession("").RouteUp()
}

func BroadcastUp(p *Packet) {
	p.BroadcastUp()
}

func (p *Packet) render(w io.Writer, sessionId string) error {
	//LogDebug("Packet.render", "sessionId", sessionId, "pkt", p)

	s := p.server
	if s == nil {
		return fmt.Errorf("Packet.server not set")
	}

	d, err := s.devices.get(p.Dst)
	if err != nil {
		return err
	}

	return d.renderPkt(w, sessionId, pkt)
}
