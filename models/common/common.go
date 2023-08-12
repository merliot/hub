package common

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/merliot/dean"
)

//go:embed css js template
var commonFs embed.FS

var pkgTmpl = template.Must(template.ParseFS(commonFs, "template/installer.tmpl"))
var serviceTmpl = template.Must(template.ParseFS(commonFs, "template/service.tmpl"))
var logTmpl = template.Must(template.ParseFS(commonFs, "template/log.tmpl"))

type Common struct {
	dean.Thing
	WebSocket string `json:"-"`
}

func New(id, model, name string) dean.Thinger {
	println("NEW COMMON")
	return &Common{
		Thing: dean.NewThing(id, model, name),
	}
}

func (c *Common) API(fs embed.FS, w http.ResponseWriter, r *http.Request) {
	_, err := fs.Open(strings.TrimPrefix(r.URL.Path, "/"))
	if os.IsNotExist(err) {
		// Not found in fs, try common fs
		http.FileServer(http.FS(commonFs)).ServeHTTP(w, r)
	} else {
		http.FileServer(http.FS(fs)).ServeHTTP(w, r)
	}

}

func (c *Common) Index(indexTmpl *template.Template, w http.ResponseWriter, r *http.Request) {
	id, _, _ := c.Identity()
	c.WebSocket = scheme + r.Host + "/ws/" + id + "/"
	if err := indexTmpl.Execute(w, c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (c *Common) ShowDeploy(deployTmpl *template.Template, w http.ResponseWriter, r *http.Request) {
	if err := deployTmpl.Execute(w, c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func genFile(tmpl *template.Template, name string, values map[string]string) error {
	file , err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()
	return tmpl.Execute(file, values)
}

func (c *Common) _deploy(buildTmpl *template.Template, w http.ResponseWriter, r *http.Request) error {

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
	case "rpi":
		envs = []string{"GOOS=linux", "GOARCH=arm", "GOARM=5"}
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

	// Generate build.go from build.tmpl
	if err := genFile(buildTmpl, "build.go", values); err != nil {
		return err
	}

	// Generate installer.go from installer.tmpl
	if err := genFile(pkgTmpl, "installer.go", values); err != nil {
		return err
	}

	// Generate model.service from service.tmpl
	if err := genFile(serviceTmpl, model + ".service", values); err != nil {
		return err
	}

	// Generate model.conf from log.tmpl
	if err := genFile(logTmpl, model + ".conf", values); err != nil {
		return err
	}

	// Build main.go -> model (binary)

	cmd := exec.Command("go", "mod", "init", model)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdoutStderr)
	}

	cmd = exec.Command("go", "mod", "tidy")
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

	// Make the file executable (e.g., 0755 permission)
	if err := os.Chmod(model, 0755); err != nil {
		return err
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

func (c *Common) Deploy(buildTmpl *template.Template, w http.ResponseWriter, r *http.Request) {
	if err := c._deploy(buildTmpl, w, r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
