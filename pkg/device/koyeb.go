//go:build !tinygo

package device

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (s *server) deployKoyeb(w http.ResponseWriter, r *http.Request) {

	var id = r.PathValue("id")
	var sessionId = r.PathValue("sessionId")

	d, exists := s.devices.get(id)
	if !exists {
		err := fmt.Errorf("Refusing to deploy: unknown device id '%s'", id)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.downloadMsgClear(d, sessionId)

	if d.isSet(flagLocked) {
		err := fmt.Errorf("Refusing to deploy: device is locked")
		s.downloadMsgError(d, sessionId, err)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// The r.URL values are passed in from the download <form>.  These
	// values are the proposed new device config, and should decode into
	// the device.  If accepted, the device is updated and the config is
	// stored in DeployParams.

	changed, err := d.formConfig(r.URL.RawQuery)
	if err != nil {
		s.downloadMsgError(d, sessionId, err)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// If the device config has changed, kick the downlink device offline.
	// It will try to reconnect, but fail, because the DeployParams now
	// don't match this (uplink) device.  Once the downlink device is
	// redeployed, the downlink device will connect.

	if changed {
		if err := s.save(); err != nil {
			s.downloadMsgError(d, sessionId, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.downlinks.linkClose(d.Id)
	}

	// Send a /downloaded msg up so uplinks can update their DeployParams

	msg := msgDownloaded{d.DeployParams}
	d.newPacket().SetPath("/downloaded").Marshal(&msg).RouteUp()

	// Redirect the browser to Koyeb to deploy the device

	devs, _ := json.Marshal(d.familyTree())
	dialurls := strings.Replace(r.Referer(), "http", "ws", 1) + "ws"

	u, _ := url.Parse("https://app.koyeb.com/deploy")

	q := u.Query()

	// See https://www.koyeb.com/docs/build-and-deploy/deploy-to-koyeb-button

	q.Set("type", "docker")
	q.Set("name", d.Model+"-"+d.Id)
	q.Set("instance_type", "eco-micro")
	q.Set("ports", "8000;http;/")
	q.Set("image", "merliot/hub")

	q.Set("env[DIAL_URLS]", dialurls)
	q.Set("env[LOG_LEVEL]", s.logLevel)
	q.Set("env[PING_PERIOD]", strconv.Itoa(s.wsxPingPeriod))
	q.Set("env[BACKGROUND]", s.background)
	q.Set("env[DEVICES]", string(devs))
	q.Set("env[AUTO_SAVE]", "false")

	// These are left blank intentionally to not give away any secrets.
	// The user must edit the Koyeb service settings to update the vars.

	q.Set("env[USER]", "")
	q.Set("env[PASSWD]", "")
	q.Set("env[WIFI_SSIDS]", "")
	q.Set("env[WIFI_PASSPHRASES]", "")

	u.RawQuery = q.Encode()

	// TODO figure out how to make this go to a new tab target="_blank"

	w.Header().Set("HX-Redirect", u.String())
}
