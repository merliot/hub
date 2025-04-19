//go:build !tinygo

package device

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

func (s *server) buildTinyGoImage(d *device, w http.ResponseWriter, r *http.Request, dir, target string) error {

	referer := r.Referer()
	dialurls := strings.Replace(referer, "http", "ws", 1) + "ws"
	ssid := r.URL.Query().Get("ssid")

	if len(s.wifiSsids) != len(s.wifiPassphrases) {
		return errors.New("Wifi SSIDS and Passphrases don't match")
	}

	passphrase := ""
	found := false
	for i, ss := range s.wifiSsids {
		if ss == ssid {
			passphrase = s.wifiPassphrases[i]
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("Wifi SSID '%s' not valid", ssid)
	}

	var p = uf2ParamsBlock{
		MagicStart: uf2Magic,
		uf2Params: uf2Params{
			Ssid:         ssid,
			Passphrase:   passphrase,
			Id:           d.Id,
			Model:        d.Model,
			Name:         d.Name,
			DeployParams: d.DeployParams,
			User:         s.user,
			Passwd:       s.passwd,
			DialURLs:     dialurls,
			LogLevel:     s.logLevel,
		},
		MagicEnd: uf2Magic,
	}

	base := filepath.Join("bin", d.Model+"-"+target+".uf2")
	installer := filepath.Join(dir, d.Model+"-"+d.Id+"-installer.uf2")

	if err := uf2Create(base, installer, p); err != nil {
		return err
	}

	return s.serveFile(w, r, installer)
}
