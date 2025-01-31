//go:build !tinygo

package device

import (
	"time"
)

type view struct {
	last       string
	level      int
	lastUpdate time.Time
}

func (d *device) lastView(sessionId string) (string, int) {

	if v, exists := d.views.Load(sessionId); exists {
		view := v.(*view)
		return view.last, view.level
	}

	return "overview", 0
}

func (d *device) saveView(sessionId, last string, level int) {
	v, _ := d.views.LoadOrStore(sessionId, &view{})
	view := v.(*view)
	view.last = last
	view.level = level
	view.lastUpdate = time.Now()
}

func gcViews(sessionId string) {

	devicesMu.RLock()
	defer devicesMu.RUnlock()

	for _, d := range devices {
		d.views.Delete(sessionId)
	}
}
