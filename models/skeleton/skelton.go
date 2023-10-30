// Skeleton Device
//
// Use this skeleton device to start a new device.  Rename instances of
// "skeleton" with your device model name.
//
// NOTE: Most of the comments in the code are instructional and can be removed
// or updated.

package skeleton

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/merliot/dean"
	"github.com/merliot/hub/models/common"
)

//
// Embed the device files in the device
//

//go:embed *
var fs embed.FS

//
// This is the device main structure which holds the device state.  Exported
// structure members will be JSON-encoded into/outof messages.
//

type Skeleton struct {
	// Skeleton device inherits from Common device
	*common.Common
	templates *template.Template

	//
	// Device-specific members here...
	//
}

//
// Specify the targets the device model supports.  See models/common/targets.go
// for list of all targets.
//

var targets = []string{"x86-64"}

//
// New creates a new device instance
//

func New(id, model, name string) dean.Thinger {
	println("NEW SKELETON")
	s := &Skeleton{}
	s.Common = common.New(id, model, name, targets).(*common.Common)
	s.CompositeFs.AddFS(fs)
	s.templates = s.CompositeFs.ParseFS("template/*")

	//
	// Device-specific members initialized here...
	//

	return s
}

//
// Message handlers for subscribed messages.  Minimally, the device should
// handle "state" and "get/state" messages to get and save state.  These
// handlers are pretty boiler-plate between devices, so they can be copied
// as-is.
//

// Save device state from the JSON-encoded message, and then broadcast the
// message to others on the bus.
func (s *Skeleton) save(msg *dean.Msg) {
	msg.Unmarshal(s).Broadcast()
}

// Reply back to requestor with device state JSON-encoded into the message.
func (s *Skeleton) getState(msg *dean.Msg) {
	s.Path = "state"
	msg.Marshal(s).Reply()
}

//
// Subscribers subscribe to message received on the device bus.
//

func (s *Skeleton) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     s.save,
		"get/state": s.getState,

		//
		// Add subscribers for other messages here...
		// 

	}
}

//
// ServeHTTP serves the device's web content.
//

func (s *Skeleton) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Use the common API to serve fs
	g.Common.API(g.templates, w, r)
}
