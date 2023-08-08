package common

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/merliot/dean"
)

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

func parseTemplate(data any, tmpls *template.Template, w http.ResponseWriter, name string) {
	err := tmpls.ExecuteTemplate(w, name, data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func (c *Common) API(embedFs embed.FS, tmpls *template.Template, w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "", "/":
		id, _, _ := c.Identity()
		c.WebSocket = scheme + r.Host + "/ws/" + id + "/"
		parseTemplate(c, tmpls, w, "index.html")
	default:
		http.FileServer(http.FS(embedFs)).ServeHTTP(w, r)
	}
}

func (c *Common) Deploy(tmpls *template.Template, w http.ResponseWriter, r *http.Request) error {

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
	values["name"] = name

	// Create temp build directory in /tmp
	dir, err := os.MkdirTemp("", id + "-")
	if err != nil {
		return err
	}
//	defer os.RemoveAll(dir)
	println(dir)

	// Generate build.go from build.tmpl, passing in request params
	destination := filepath.Join(dir, "build.go")
	outputFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	err = tmpls.ExecuteTemplate(outputFile, "build.tmpl", values)
	if err != nil {
		return err
	}

	// Generate pkg.go from pkg.tmpl, passing in request params
	destination = filepath.Join(dir, "pkg.go")
	outputFile, err = os.Create(destination)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	err = tmpls.ExecuteTemplate(outputFile, "pkg.tmpl", values)
	if err != nil {
		return err
	}

	// Build main.go -> model (binary)

	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(wd)

	// Change the working directory to temp build directory
	if err = os.Chdir(dir); err != nil {
		return err
	}

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
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdoutStderr)
	}

	cmd = exec.Command("go", "build", "-o", id, "pkg.go")
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdoutStderr)
	}

	return nil
}
