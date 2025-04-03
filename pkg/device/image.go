//go:build !tinygo

package device

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tpkg "github.com/merliot/hub/pkg/target"
)

func (s *server) setContentMd5(w http.ResponseWriter, fileName string) error {

	// Calculate MD5 checksum
	cmd := exec.Command("md5sum", fileName)
	s.logDebug(cmd.String())
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdoutStderr)
	}
	md5sum := bytes.Fields(stdoutStderr)[0]
	md5sumBase64 := base64.StdEncoding.EncodeToString(md5sum)

	// Set the MD5 checksum header
	w.Header().Set("Content-MD5", md5sumBase64)
	return nil
}

func (s *server) serveFile(w http.ResponseWriter, r *http.Request, fileName string) error {

	if err := s.setContentMd5(w, fileName); err != nil {
		return err
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(fileName))
	w.Header().Set("Content-Type", "application/octet-stream")

	s.logInfo("Serving download file", "name", filepath.Base(fileName))
	http.ServeFile(w, r, fileName)

	return nil
}

func (d *device) genFile(dir, template, name string, data any) error {
	filePath := filepath.Join(dir, name)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	return d.renderTmpl(file, template, data)
}

func isLocalhost(referer string) bool {
	url, _ := url.Parse(referer)
	hostname := url.Hostname()
	return hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1" || hostname == "0.0.0.0"
}

// createSFX concatenates the tar ball to the SFX installer script, making the
// SFX installer
func createSFX(dir, sfxFile, tarFile, installerFile string) error {

	script, err := os.Open(filepath.Join(dir, sfxFile))
	if err != nil {
		return fmt.Errorf("error opening sfx script file: %v", err)
	}
	defer script.Close()

	archive, err := os.Open(filepath.Join(dir, tarFile))
	if err != nil {
		return fmt.Errorf("error opening archive file: %v", err)
	}
	defer archive.Close()

	installer, err := os.Create(filepath.Join(dir, installerFile))
	if err != nil {
		return fmt.Errorf("error creating installer file: %v", err)
	}
	defer installer.Close()

	// Copy the script file to the installer file
	if _, err := io.Copy(installer, script); err != nil {
		return fmt.Errorf("error writing script to installer file: %v", err)
	}

	// Copy the archive file to the installer file
	if _, err := io.Copy(installer, archive); err != nil {
		return fmt.Errorf("error writing archive to installer file: %v", err)
	}

	return nil
}

func (s *server) buildLinuxImage(d *device, w http.ResponseWriter, r *http.Request, dir, target string) error {

	var referer = r.Referer()
	var service = d.Model + "-" + d.Id
	var dialurls = strings.Replace(referer, "http", "ws", 1) + "ws"

	// Generate environment variable file.  The service will load env vars
	// from this file.
	if err := d.genFile(dir, "device-env.tmpl", "env", map[string]any{
		"port":       r.URL.Query().Get("port"),
		"user":       s.user,
		"passwd":     s.passwd,
		"dialurls":   dialurls,
		"logLevel":   s.logLevel,
		"pingPeriod": s.wsxPingPeriod,
		"background": s.background,
		"autoSave":   s.isSet(flagAutoSave),
	}); err != nil {
		return err
	}

	// Generate systemd merliot.target unit from
	// device-merliot-target.tmpl.  This will be the parent unit of all
	// device units.
	targetFile := "merliot.target"
	if err := d.genFile(dir, "device-merliot-target.tmpl", targetFile, nil); err != nil {
		return err
	}

	// Generate systemd {{service}}.service unit from device-service.tmpl
	serviceFile := service + ".service"
	if err := d.genFile(dir, "device-service.tmpl", serviceFile, map[string]any{
		"service": service,
	}); err != nil {
		return err
	}

	// Generate {{service}}.conf from device-conf.tmpl.  This sets up
	// logging service for the device.  Logs are available at
	// /var/log/{{.service}}.log.
	confFile := service + ".conf"
	if err := d.genFile(dir, "device-conf.tmpl", confFile, map[string]any{
		"service": service,
	}); err != nil {
		return err
	}

	// Generate SelF-eXtracting (SFX) installer script
	sfxFile := "sfx.sh"
	if err := d.genFile(dir, "device-sfx.tmpl", sfxFile, map[string]any{
		"service": service,
	}); err != nil {
		return err
	}

	// Generate service install script
	if err := d.genFile(dir, "device-install.tmpl", "install.sh", map[string]any{
		"service": service,
	}); err != nil {
		return err
	}

	// Make a devices.json file
	if err := fileWriteJSON(filepath.Join(dir, "devices.json"), d.familyTree()); err != nil {
		return err
	}

	// Figure out which binaries to include in installer such that this
	// device can reproduce any child model device

	var binFiles []string
	childModels := s.childModels(d)
	if childModels.length() == 0 {
		// Sterile device only needs the bin/device-<target> binary
		binFiles = append(binFiles, "-C", ".", "./bin/device-"+target)
	} else {
		// Copy over the binaries needed to produce children devices
		binFiles = append(binFiles, "-C", ".", "./bin/device-rpi")
		binFiles = append(binFiles, "-C", ".", "./bin/device-x86-64")
		childModels.drange(func(name string, model *Model) bool {
			// Copy the UF2 files for the model (all targets)
			for _, t := range tpkg.TinyGoTargets(model.Config.Targets) {
				binFiles = append(binFiles, "-C", ".", "./bin/"+name+"-"+t+".uf2")
			}
			return true
		})
	}

	// Create a gzipped tar ball with everything inside need to
	// install/uninstall the device

	tarFile := service + ".tar.gz"
	tarFilePath := filepath.Join(dir, tarFile)

	args := []string{"--exclude", tarFile, "-czf", tarFilePath}
	args = append(args, binFiles...)
	args = append(args, "-C", dir, ".")

	cmd := exec.Command("tar", args...)
	s.logDebug(cmd.String())
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdoutStderr)
	}

	// Create final SFX image as the installer

	installer := service + "-installer"
	if err := createSFX(dir, sfxFile, tarFile, installer); err != nil {
		return err
	}

	// Serve installer file for download

	return s.serveFile(w, r, filepath.Join(dir, installer))
}

func (s *server) buildTinyGoImage(d *device, w http.ResponseWriter, r *http.Request, dir, target string) error {

	var referer = r.Referer()
	var dialurls = strings.Replace(referer, "http", "ws", 1) + "ws"
	var ssid = r.URL.Query().Get("ssid")

	if len(s.wifiSsids) != len(s.wifiPassphrases) {
		return errors.New("Wifi SSIDS and Passphrases don't match")
	}

	var passphrase = ""
	var found bool
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

func (s *server) buildImage(d *device, w http.ResponseWriter, r *http.Request) error {

	// Create temp build directory
	dir, err := os.MkdirTemp("", d.Model+"-"+d.Id+"-")
	if err != nil {
		return err
	}

	if s.isSet(flagDebugKeepBuilds) {
		s.logDebug("Temporary build", "dir", dir)
	} else {
		defer os.RemoveAll(dir)
	}

	target := r.URL.Query().Get("target")

	switch target {
	case "x86-64", "rpi":
		return s.buildLinuxImage(d, w, r, dir, target)
	case "nano-rp2040", "wioterminal", "pyportal":
		return s.buildTinyGoImage(d, w, r, dir, target)
	default:
		return fmt.Errorf("Target '%s' not supported", target)
	}

	return nil
}

func (s *server) downloadMsgClear(d *device, sessionId string) {
	var buf bytes.Buffer
	if err := d.renderTmpl(&buf, "device-download-msg-empty.tmpl", nil); err != nil {
		s.logError("Rendering template", "err", err)
		return
	}
	s.sessions.send(sessionId, string(buf.Bytes()))
}

func (s *server) downloadMsgError(d *device, sessionId string, downloadErr error) {
	var buf bytes.Buffer
	if err := d.renderTmpl(&buf, "device-download-msg-error.tmpl", map[string]any{
		"err": "Download error: " + downloadErr.Error(),
	}); err != nil {
		s.logError("Rendering template", "err", err)
		return
	}
	s.sessions.send(sessionId, string(buf.Bytes()))
}

type msgDownloaded struct {
	DeployParams string
}

func (s *server) handleDownloaded(pkt *Packet) {
	var msg msgDownloaded

	d, exists := s.devices.get(pkt.Dst)
	if !exists {
		s.logError("Handling downloaded", "err", deviceNotFound(pkt.Dst))
		return
	}

	pkt.Unmarshal(&msg)
	d.formConfig(string(msg.DeployParams))
	pkt.BroadcastUp()
}

func (s *server) downloadImage(w http.ResponseWriter, r *http.Request) {

	var referer = r.Referer()
	var id = r.PathValue("id")
	var sessionId = r.PathValue("sessionId")

	d, exists := s.devices.get(id)
	if !exists {
		err := fmt.Errorf("Can't download image: unknown device id '%s'", id)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.downloadMsgClear(d, sessionId)

	if d.isSet(flagLocked) {
		err := fmt.Errorf("Refusing to download: device is locked")
		s.downloadMsgError(d, sessionId, err)
		http.Error(w, err.Error(), http.StatusLocked)
		return
	}

	if isLocalhost(referer) {
		err := fmt.Errorf("Cannot use localhost for hub address.  Access the hub " +
			"using the hostname or IP address of the host; something that " +
			"is addressable on the network so the device can dial into the hub.")
		s.downloadMsgError(d, sessionId, err)
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	// The r.URL values are passed in from the download <form>.  These
	// values are the proposed new device config, and should decode into
	// the device.  If accepted, the device is updated and the config is
	// stored in DeployParams.

	changed, err := d.formConfig(r.URL.RawQuery)
	if err != nil {
		s.downloadMsgError(d, sessionId, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Built it!

	if err := s.buildImage(d, w, r); err != nil {
		s.downloadMsgError(d, sessionId, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// If the device config has changed, kick the downlink device offline.
	// It will try to reconnect, but fail, because the DeployParams now
	// don't match this (uplink) device.  Once the downlink device is
	// updated (with the image we created above) the downlink device
	// will connect.

	if changed {
		s.save()
		s.downlinks.linkClose(d.Id)
	}

	// Send a /downloaded msg up so uplinks can update their DeployParams

	msg := msgDownloaded{d.DeployParams}
	d.newPacket().SetPath("/downloaded").Marshal(&msg).RouteUp()
}
