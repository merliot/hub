//go:build !tinygo

package device

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	keepBuilds = Getenv("DEBUG_KEEP_BUILDS", "")
)

func gzipFile(src, dst string) error {
	inputFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	outputFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	gzipWriter := gzip.NewWriter(outputFile)
	defer gzipWriter.Close()

	_, err = io.Copy(gzipWriter, inputFile)
	return err
}

func serveFile(w http.ResponseWriter, r *http.Request, fileName string) error {

	// Calculate MD5 checksum
	cmd := exec.Command("md5sum", fileName)
	LogDebug(cmd.String())
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdoutStderr)
	}
	md5sum := bytes.Fields(stdoutStderr)[0]
	md5sumBase64 := base64.StdEncoding.EncodeToString(md5sum)

	// Set the Content-Disposition header to suggest the original filename for download
	downloadName := filepath.Base(fileName)
	w.Header().Set("Content-Disposition", "attachment; filename="+downloadName)
	// Set the MD5 checksum header
	w.Header().Set("Content-MD5", md5sumBase64)

	// Gzip the file
	w.Header().Set("Content-Encoding", "gzip")
	gzipName := fileName + ".gz"
	err = gzipFile(fileName, gzipName)
	if err != nil {
		return err
	}

	LogInfo("Serving download file", "name", gzipName)
	http.ServeFile(w, r, gzipName)
	LogInfo("Done serving download file", "name", gzipName)

	return nil
}

func (d *device) genFile(template string, name string, data any) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()
	return d.renderTmpl(file, template, data)
}

func isLocalhost(referer string) bool {
	url, _ := url.Parse(referer)
	hostname := url.Hostname()
	return hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1"
}

func (d *device) buildLinuxImage(w http.ResponseWriter, r *http.Request, dir string,
	envs []string, target string) error {

	var referer = r.Referer()
	if isLocalhost(referer) {
		return fmt.Errorf("Cannot use localhost for hub address.  Access the hub " +
			"using the hostname or IP address of the host; something that " +
			"is addressable on the network so the device can dial into the hub.")
	}

	var dialurls = strings.Replace(referer, "http", "ws", 1) + "ws"
	var service = d.Model + "-" + d.Id

	// Generate runner.go from device-runner-linux.tmpl
	var runnerGo = filepath.Join(dir, "runner.go")
	if err := d.genFile("device-runner-linux.tmpl", runnerGo, map[string]any{
		"user":     Getenv("USER", ""),
		"passwd":   Getenv("PASSWD", ""),
		"dialurls": dialurls,
		"port":     r.URL.Query().Get("port"),
	}); err != nil {
		return err
	}

	// Generate installer.go from device-installer.tmpl
	var installerGo = filepath.Join(dir, "installer.go")
	if err := d.genFile("device-installer.tmpl", installerGo, map[string]any{
		"service": service,
	}); err != nil {
		return err
	}

	// Generate systemd merliot.target unit from device-merliot-target.tmpl
	var output = filepath.Join(dir, "merliot.target")
	if err := d.genFile("device-merliot-target.tmpl", output, nil); err != nil {
		return err
	}

	// Generate systemd {{service}}.service unit from device-service.tmpl
	output = filepath.Join(dir, service+".service")
	if err := d.genFile("device-service.tmpl", output, map[string]any{
		"service": service,
	}); err != nil {
		return err
	}

	// Generate {{service}}.conf from device-conf.tmpl
	output = filepath.Join(dir, service+".conf")
	if err := d.genFile("device-conf.tmpl", output, map[string]any{
		"service": service,
	}); err != nil {
		return err
	}

	// Build runner binary

	// substitute "-" for "_" in target, ala TinyGo, when using as tag
	var tag = strings.Replace(target, "-", "_", -1)
	var binary = filepath.Join(dir, service)

	cmd := exec.Command("go", "build", "-race", "-ldflags", "-s -w", "-o", binary, "-tags", tag, runnerGo)
	LogDebug(cmd.String())
	cmd.Env = append(cmd.Environ(), envs...)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdoutStderr)
	}

	// Build installer

	var installer = filepath.Join(dir, service+"-installer")

	cmd = exec.Command("go", "build", "-ldflags", "-s -w", "-o", installer, installerGo)
	LogDebug(cmd.String())
	cmd.Env = append(cmd.Environ(), envs...)
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdoutStderr)
	}

	// Serve installer file for download

	LogInfo("Built Linux device image", "installer", installer)
	return serveFile(w, r, installer)
}

func (d *device) buildTinyGoImage(w http.ResponseWriter, r *http.Request, dir, target string) error {

	var referer = r.Referer()
	if isLocalhost(referer) {
		return fmt.Errorf("Cannot use localhost for hub.  Use the hub " +
			"hostname or IP address; something that is addressable so the " +
			"device can dial into the hub.")
	}

	var dialurls = strings.Replace(referer, "http", "ws", 1) + "ws"
	var ssid = r.URL.Query().Get("ssid")
	var wifiAuths = wifiAuths()
	var passphrase = wifiAuths[ssid]

	var p = uf2ParamsBlock{
		MagicStart: uf2Magic,
		uf2Params: uf2Params{
			Ssid:         ssid,
			Passphrase:   passphrase,
			Id:           d.Id,
			Model:        d.Model,
			Name:         d.Name,
			DeployParams: d.DeployParams,
			User:         Getenv("USER", ""),
			Passwd:       Getenv("PASSWD", ""),
			DialURLs:     dialurls,
		},
		MagicEnd: uf2Magic,
	}

	base := filepath.Join("uf2s", d.Model+"-"+target+".uf2")
	installer := filepath.Join(dir, d.Model+"-"+d.Id+"-installer.uf2")

	if err := uf2Create(base, installer, p); err != nil {
		return err
	}

	LogInfo("Built Tinygo device image", "installer", installer)
	return serveFile(w, r, installer)
}

func (d *device) buildImage(w http.ResponseWriter, r *http.Request) error {

	// Create temp build directory
	dir, err := os.MkdirTemp("./", d.Model+"-"+d.Id+"-")
	if err != nil {
		return err
	}

	if keepBuilds != "" {
		LogDebug("Temporary build", "dir", dir)
	} else {
		defer os.RemoveAll(dir)
	}

	target := r.URL.Query().Get("target")

	switch target {
	case "x86-64":
		envs := []string{"CGO_ENABLED=1", "GOOS=linux", "GOARCH=amd64"}
		return d.buildLinuxImage(w, r, dir, envs, target)
	case "rpi":
		// TODO: do we want more targets for GOARM=7|8?
		envs := []string{"CGO_ENABLED=1", "GOOS=linux", "GOARCH=arm", "GOARM=5"}
		return d.buildLinuxImage(w, r, dir, envs, target)
	case "nano-rp2040", "wioterminal", "pyportal":
		return d.buildTinyGoImage(w, r, dir, target)
	default:
		return fmt.Errorf("Target '%s' not supported", target)
	}

	return nil
}

func (d *device) downloadMsgClear(sessionId string) {
	var buf bytes.Buffer
	if err := d.renderTmpl(&buf, "device-download-msg-empty.tmpl", nil); err != nil {
		LogError("Rendering template", "err", err)
		return
	}
	sessionSend(sessionId, string(buf.Bytes()))
}

func (d *device) downloadMsgError(sessionId string, downloadErr error) {
	var buf bytes.Buffer
	if err := d.renderTmpl(&buf, "device-download-msg-error.tmpl", map[string]any{
		"err": "Download error: " + downloadErr.Error(),
	}); err != nil {
		LogError("Rendering template", "err", err)
		return
	}
	sessionSend(sessionId, string(buf.Bytes()))
}

func (d *device) downloadImage(w http.ResponseWriter, r *http.Request) {

	var sessionId = r.PathValue("sessionId")

	d.downloadMsgClear(sessionId)

	if d.IsSet(flagLocked) {
		err := fmt.Errorf("Refusing to download: device is locked")
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

	// Built it!

	if err := d.buildImage(w, r); err != nil {
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
		deviceDirty(root.Id)
		downlinkClose(d.Id)
	}
}
