//go:build !tinygo

package common

import (
	"fmt"
	"html/template"
	"net/http"
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

	file , err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, values)
}

func (c *Common) _deploy(templates *template.Template, w http.ResponseWriter, r *http.Request) error {

	var values = make(map[string]string)

	// Squash request params down to first value for each key.  The resulting
	// map[string]string is much nicer to pass to html/template as data value.

	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			values[k] = v[0]
		}
	}

	id, model, name := c.Identity()

	values["id"] = id
	values["model"] = model
	values["modelStruct"] = strings.Title(model)
	values["name"] = name
	values["hub"] = r.Host
	values["scheme"] = "wss://"
	if r.TLS == nil {
		values["scheme"] = "ws://"
	}

	if user, passwd, ok := r.BasicAuth(); ok {
		values["user"] = user
		values["passwd"] = passwd
	}

	envs := []string{}
	switch values["target"] {
	case "x86_64":
		envs = []string{"CGO_ENABLED=0", "GOOS=linux", "GOARCH=amd64"}
	case "rpi":
		envs = []string{"CGO_ENABLED=0", "GOOS=linux", "GOARCH=arm", "GOARM=5"}
	}

	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(wd)

	// Create temp build directory in /tmp
	dir, err := os.MkdirTemp("", id + "-")
	if err != nil {
		return err
	}
//	defer os.RemoveAll(dir)
	println(dir)

	// Change the working directory to temp build directory
	if err = os.Chdir(dir); err != nil {
		return err
	}

	// Generate build.go from server.tmpl
	if err := genFile(templates, "server.tmpl", "build.go", values); err != nil {
		return err
	}

	// Generate installer.go from installer.tmpl
	if err := genFile(templates, "installer.tmpl", "installer.go", values); err != nil {
		return err
	}

	// Generate model.service from service.tmpl
	if err := genFile(templates, "service.tmpl", model + ".service", values); err != nil {
		return err
	}

	// Generate model.conf from log.tmpl
	if err := genFile(templates, "log.tmpl", model + ".conf", values); err != nil {
		return err
	}

	// Build main.go -> model (binary)

	cmd := exec.Command("go", "mod", "init", model)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdoutStderr)
	}

	cmd = exec.Command("go", "mod", "tidy", "-e")
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdoutStderr)
	}

	cmd = exec.Command("go", "build", "-o", model, "build.go")
	cmd.Env = append(cmd.Environ(), envs...)
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdoutStderr)
	}

	// Build installer and serve as download-able file

	installer := id + "-installer"
	cmd = exec.Command("go", "build", "-o", installer, "installer.go")
	cmd.Env = append(cmd.Environ(), envs...)
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdoutStderr)
	}

	// Set the Content-Disposition header to suggest the original filename for download
	w.Header().Set("Content-Disposition", "attachment; filename="+installer)

	http.ServeFile(w, r, installer)

	return nil
}

func (c *Common) deploy(templates *template.Template, w http.ResponseWriter, r *http.Request) {
	if err := c._deploy(templates, w, r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
