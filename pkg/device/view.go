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

	return d.InitialView, 0
}

func (d *device) saveView(sessionId, last string, level int) {
	v, _ := d.views.LoadOrStore(sessionId, &view{})
	view := v.(*view)
	view.last = last
	view.level = level
	view.lastUpdate = time.Now()
}

func (s *server) gcDeviceViews(d *device) {
	d.views.Range(func(key, _ any) bool {
		sessionId := key.(string)
		_, exists := s.sessions.get(sessionId)
		if !exists {
			d.views.Delete(sessionId)
		}
		return true
	})
}

func (s *server) gcViews() {
	minute := 1 * time.Minute
	ticker := time.NewTicker(minute)
	defer ticker.Stop()
	for range ticker.C {
		s.devices.drange(func(id string, d *device) bool {
			s.gcDeviceViews(d)
			return true
		})
	}
}
