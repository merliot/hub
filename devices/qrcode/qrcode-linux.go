//go:build !tinygo

package qrcode

import (
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"

	"github.com/merliot/hub/pkg/device"
	goqr "github.com/skip2/go-qrcode"
)

//go:embed *.go images template
var fs embed.FS

type qrcode struct {
	Content string `schema:"required,desc=QR code content"`
}

func (q qrcode) Desc() string {
	return "QR code"
}

func (q *qrcode) GetConfig() device.Config {
	return device.Config{
		Model:   "qrcode",
		Parents: []string{"hub"},
		State:   q,
		FS:      &fs,
		Targets: []string{"pyportal"},
		BgColor: "magenta",
		FgColor: "black",
		PacketHandlers: device.PacketHandlers{
			"/update": &device.PacketHandler[qrcode]{q.update},
		},
		APIs: device.APIs{
			"POST /generate":    q.generate,
			"GET /edit-content": q.editContent,
		},
		FuncMap: device.FuncMap{
			"png": q.png,
		},
	}
}

func (q *qrcode) Setup() error { return nil }

func (q *qrcode) update(pkt *device.Packet) {
	pkt.Unmarshal(q).BroadcastUp()
}

func (q *qrcode) png(content string, size int) (template.URL, error) {
	if content == "" {
		content = "missing content?"
	}

	// Generate the QR code as PNG image
	var png []byte
	png, err := goqr.Encode(content, goqr.Medium, size)
	if err != nil {
		return "", err
	}
	// Convert the image to base64
	base64Image := base64.StdEncoding.EncodeToString(png)
	url := fmt.Sprintf("data:image/png;base64,%s", base64Image)
	// Return it as template-safe url to use with <img src={{.}}>
	return template.URL(url), nil
}

func (q *qrcode) generate(w http.ResponseWriter, r *http.Request) {

	content := r.FormValue("Content")

	url, err := q.png(content, -5)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tmpl := template.Must(template.New("qr").Parse(`<img src="{{.}}">`))
	tmpl.Execute(w, url)
}

var tmpl = `<input class="m-4" type="text"
	name="Content" value="%s" autofocus required
	hx-post="/device/%s/update"
	hx-swap="none">`

func (q *qrcode) editContent(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	html := fmt.Sprintf(tmpl, q.Content, id)
	w.Write([]byte(html))
}
