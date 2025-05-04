//go:build !tinygo

package device

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	localhost     = "localhost"
	localhostIPv4 = "127.0.0.1"
	localhostIPv6 = "::1"
	localhostAny  = "0.0.0.0"
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
	url, err := url.Parse(referer)
	if err != nil {
		return false
	}
	hostname := url.Hostname()
	return hostname == localhost || hostname == localhostIPv4 || hostname == localhostIPv6 || hostname == localhostAny
}

// createSFX concatenates the tar ball to the SFX installer script, making the
// SFX installer
func createSFX(dir, sfxFile, tarFile, installerFile string) error {

	script, err := os.Open(filepath.Join(dir, sfxFile))
	if err != nil {
		return fmt.Errorf("%w: opening sfx script file", err)
	}
	defer script.Close()

	archive, err := os.Open(filepath.Join(dir, tarFile))
	if err != nil {
		return fmt.Errorf("%w: opening archive file", err)
	}
	defer archive.Close()

	installer, err := os.Create(filepath.Join(dir, installerFile))
	if err != nil {
		return fmt.Errorf("%w: creating installer file", err)
	}
	defer installer.Close()

	// Copy the script file to the installer file
	if _, err := io.Copy(installer, script); err != nil {
		return fmt.Errorf("%w: writing script to installer file", err)
	}

	// Copy the archive file to the installer file
	if _, err := io.Copy(installer, archive); err != nil {
		return fmt.Errorf("%w: writing archive to installer file", err)
	}

	return nil
}

// createTarBall creates a gzipped tar archive containing the specified files and directories.
// It excludes the tar file itself from the archive to prevent recursion.
func (s *server) createTarBall(dir, tarFile string, binFiles []string) error {
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
	return nil
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
}

func (s *server) downloadMsgClear(d *device, sessionId string) {
	var buf bytes.Buffer
	if err := d.renderTmpl(&buf, "device-download-msg-empty.tmpl", nil); err != nil {
		s.logError("Rendering template", "err", err)
		return
	}
	s.sessions.send(sessionId, buf.String())
}

func (s *server) downloadMsgError(d *device, sessionId string, downloadErr error) {
	var buf bytes.Buffer
	if err := d.renderTmpl(&buf, "device-download-msg-error.tmpl", map[string]any{
		"err": "Download error: " + downloadErr.Error(),
	}); err != nil {
		s.logError("Rendering template", "err", err)
		return
	}
	s.sessions.send(sessionId, buf.String())
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

func (s *server) _downloadImage(d *device, w http.ResponseWriter, r *http.Request) error {

	if d.isSet(flagLocked) {
		return fmt.Errorf("Refusing to download: device is locked")
	}

	if isLocalhost(r.Referer()) {
		return fmt.Errorf("Cannot use localhost for hub address.  Access the hub " +
			"using the hostname or IP address of the host; something that " +
			"is addressable on the network so the device can dial into the hub.")
	}

	// The r.URL values are passed in from the download <form>.  These
	// values are the proposed new device config, and should decode into
	// the device.  If accepted, the device is updated and the config is
	// stored in DeployParams.

	changed, err := d.formConfig(r.URL.RawQuery)
	if err != nil {
		return err
	}

	// Built it!

	if err := s.buildImage(d, w, r); err != nil {
		return err
	}

	// If the device config has changed, kick the downlink device offline.
	// It will try to reconnect, but fail, because the DeployParams now
	// don't match this (uplink) device.  Once the downlink device is
	// updated (with the image we created above) the downlink device
	// will connect.

	if changed {
		if err := s.save(); err != nil {
			return err
		}
		s.downlinks.linkClose(d.Id)
	}

	return nil
}

func (s *server) downloadImage(w http.ResponseWriter, r *http.Request) {

	var id = r.PathValue("id")
	var sessionId = r.PathValue("sessionId")

	d, exists := s.devices.get(id)
	if !exists {
		err := fmt.Errorf("Can't download image: unknown device id '%s'", id)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.downloadMsgClear(d, sessionId)

	err := s._downloadImage(d, w, r)
	if err != nil {
		if sessionId == "" {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			s.downloadMsgError(d, sessionId, err)
			w.WriteHeader(http.StatusNoContent)
		}
		return
	}

	// Send a downloaded msg up so uplinks can update their DeployParams

	msg := msgDownloaded{d.DeployParams}
	d.newPacket().SetPath("downloaded").Marshal(&msg).BroadcastUp()
}
