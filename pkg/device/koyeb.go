//go:build !tinygo

package device

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func (d *device) deployKoyeb(w http.ResponseWriter, r *http.Request) {

	var sessionId = r.PathValue("sessionId")

	d.downloadMsgClear(sessionId)

	if d.isSet(flagLocked) {
		err := fmt.Errorf("Refusing to deploy: device is locked")
		d.downloadMsgError(sessionId, err)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// The r.URL values are passed in from the download <form>.  These
	// values are the proposed new device config, and should decode into
	// the device.  If accepted, the device is updated and the config is
	// stored in DeployParams.

	changed, err := d.formConfig(r.URL.RawQuery)
	if err != nil {
		d.downloadMsgError(sessionId, err)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// If the device config has changed, kick the downlink device offline.
	// It will try to reconnect, but fail, because the DeployParams now
	// don't match this (uplink) device.  Once the downlink device is
	// updated (with the image we created above) the downlink device
	// will connect.

	if changed {
		root.save()
		downlinkClose(d.Id)
	}

	// Send a /downloaded msg up so uplinks can update their DeployParams

	msg := MsgDownloaded{d.DeployParams}
	pkt := Packet{Dst: d.Id, Path: "/downloaded"}
	pkt.Marshal(&msg).RouteUp()

	// Redirect the browser to Koyeb to build the device

	devs, _ := json.Marshal(d.devices())
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
	q.Set("env[LOG_LEVEL]", logLevel)
	q.Set("env[PING_PERIOD]", pingPeriod)
	q.Set("env[BACKGROUND]", Getenv("BACKGROUND", ""))
	q.Set("env[DEVICES]", string(devs))

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
