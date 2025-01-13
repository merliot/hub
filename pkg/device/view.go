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

type views map[string]*view // key: sessionId

func (d *device) lastView(sessionId string) (view string, level int) {

	d.viewsMu.RLock()
	defer d.viewsMu.RUnlock()

	if v, exists := d.views[sessionId]; exists {
		return v.last, v.level
	}

	return "overview", 0
}

func (d *device) saveView(sessionId, last string, level int) {

	d.viewsMu.Lock()
	defer d.viewsMu.Unlock()

	v, exists := d.views[sessionId]
	if !exists {
		v = &view{}
		d.views[sessionId] = v
	}
	v.last = last
	v.level = level
	v.lastUpdate = time.Now()
}

func gcViews(sessionId string) {

	devicesMu.RLock()
	defer devicesMu.RUnlock()

	for _, d := range devices {
		d.viewsMu.Lock()
		if _, exists := d.views[sessionId]; exists {
			delete(d.views, sessionId)
		}
		d.viewsMu.Unlock()
	}
}
