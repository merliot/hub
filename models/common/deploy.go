//go:build !tinygo

package common

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

func genFile(templates *template.Template, template string, name string,
	values map[string]string) error {

	tmpl := templates.Lookup(template)
	if tmpl == nil {
		return fmt.Errorf("Template '%s' not found", template)
	}

	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, values)
}

func (c *Common) deployGo(dir string, values map[string]string, envs []string,
	templates *template.Template, w http.ResponseWriter, r *http.Request) error {

	// Generate build.go from server.tmpl
	if err := genFile(templates, "server.tmpl", dir+"/build.go", values); err != nil {
		return err
	}

	// Generate installer.go from installer.tmpl
	if err := genFile(templates, "installer.tmpl", dir+"/installer.go", values); err != nil {
		return err
	}

	// Generate model.service from service.tmpl
	if err := genFile(templates, "service.tmpl", dir+"/"+c.Model+".service", values); err != nil {
		return err
	}

	// Generate model.conf from log.tmpl
	if err := genFile(templates, "log.tmpl", dir+"/"+c.Model+".conf", values); err != nil {
		return err
	}

	// Build build.go -> model (binary)

	// substitute "-" for "_" in target, ala TinyGo, when using as tag
	target := strings.Replace(values["target"], "-", "_", -1)

	cmd := exec.Command("go", "build", "-o", dir+"/"+c.Model, "-tags", target, dir+"/build.go")
	println(cmd.String())
	cmd.Env = append(cmd.Environ(), envs...)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdoutStderr)
	}

	// Build installer and serve as download-able file

	installer := c.Id + "-installer"
	cmd = exec.Command("go", "build", "-o", dir+"/"+installer, dir+"/installer.go")
	println(cmd.String())
	cmd.Env = append(cmd.Environ(), envs...)
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdoutStderr)
	}

	// Calculate MD5 checksum of installer
	cmd = exec.Command("md5sum", dir+"/"+installer)
	println(cmd.String())
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdoutStderr)
	}
	md5sum := bytes.Fields(stdoutStderr)[0]
	md5sumBase64 := base64.StdEncoding.EncodeToString(md5sum)

	// Set the Content-Disposition header to suggest the original filename for download
	w.Header().Set("Content-Disposition", "attachment; filename="+installer)
	// Set the MD5 checksum header
	w.Header().Set("Content-MD5", md5sumBase64)

	http.ServeFile(w, r, dir+"/"+installer)

	return nil
}

func (c *Common) deployTinyGo(dir string, values map[string]string, envs []string,
	templates *template.Template, w http.ResponseWriter, r *http.Request) error {

	// Generate build.go from runner.tmpl
	if err := genFile(templates, "runner.tmpl", dir+"/build.go", values); err != nil {
		return err
	}

	// Build build.go -> uf2 binary

	installer := c.Id + "-installer.uf2"
	target := values["target"]

	cmd := exec.Command("tinygo", "build", "-target", target, "-stack-size", "8kb",
		"-o", dir+"/"+installer, dir+"/build.go")
	println(cmd.String())
	cmd.Env = append(cmd.Environ(), envs...)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdoutStderr)
	}

	// Calculate MD5 checksum of installer
	cmd = exec.Command("md5sum", dir+"/"+installer)
	println(cmd.String())
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdoutStderr)
	}
	md5sum := bytes.Fields(stdoutStderr)[0]
	md5sumBase64 := base64.StdEncoding.EncodeToString(md5sum)

	// Set the Content-Disposition header to suggest the original filename for download
	w.Header().Set("Content-Disposition", "attachment; filename="+installer)
	// Set the MD5 checksum header
	w.Header().Set("Content-MD5", md5sumBase64)

	http.ServeFile(w, r, dir+"/"+installer)

	return nil
}

func (c *Common) buildValues(r *http.Request) (map[string]string, error) {

	var values = make(map[string]string)

	// Squash request params down to first value for each key.  The resulting
	// map[string]string is much nicer to pass to html/template as data value.

	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			values[k] = v[0]
		}
	}

	values["deployParams"] = c.DeployParams
	values["id"] = c.Id
	values["model"] = c.Model
	values["modelStruct"] = strings.Title(c.Model)
	values["name"] = c.Name

	if ssid, ok := values["ssid"]; ok {
		values["passphrase"] = c.WifiAuth[ssid]
	}

	values["hub"] = wsScheme + r.Host + "/ws/?ping-period=4"

	if values["backuphub"] != "" {
		u, err := url.Parse(values["backuphub"])
		if err != nil {
			return nil, err
		}
		scheme := "ws://"
		if u.Scheme == "https" {
			scheme = "wss://"
		}
		values["backupHub"] = scheme + u.Host + "/ws/?ping-period=4"
	}

	if user, passwd, ok := r.BasicAuth(); ok {
		values["user"] = user
		values["passwd"] = passwd
	}

	return values, nil
}

func (c *Common) buildEnvs(values map[string]string) []string {
	envs := []string{}
	switch values["target"] {
	case "demo", "x86-64":
		envs = []string{"CGO_ENABLED=0", "GOOS=linux", "GOARCH=amd64"}
	case "rpi":
		// TODO: do we want more targets for GOARM=7|8?
		envs = []string{"CGO_ENABLED=0", "GOOS=linux", "GOARCH=arm", "GOARM=5"}
	}
	return envs
}

func (c *Common) _deploy(templates *template.Template, w http.ResponseWriter, r *http.Request) error {

	values, err := c.buildValues(r)
	if err != nil {
		return err
	}

	envs := c.buildEnvs(values)

	// Create temp build directory
	dir, err := os.MkdirTemp("./", c.Id+"-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	//println(dir)

	switch values["target"] {
	case "demo", "x86-64", "rpi":
		return c.deployGo(dir, values, envs, templates, w, r)
	case "nano-rp2040", "wioterminal", "pyportal":
		return c.deployTinyGo(dir, values, envs, templates, w, r)
	default:
		return errors.New("Target not supported")
	}

	return nil
}

func (c *Common) deploy(templates *template.Template, w http.ResponseWriter, r *http.Request) {
	c.DeployParams = r.URL.RawQuery
	if err := c._deploy(templates, w, r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	c.Save()
}
