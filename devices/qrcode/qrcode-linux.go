//go:build !tinygo

package qrcode

import (
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"

	"github.com/merliot/hub"
	goqr "github.com/skip2/go-qrcode"
)

//go:embed *.go images template
var fs embed.FS

type qrcode struct {
	Content string
}

func (q *qrcode) GetConfig() hub.Config {
	return hub.Config{
		Model:   "qrcode",
		State:   q,
		FS:      &fs,
		Targets: []string{"wioterminal", "pyportal"},
		BgColor: "magenta",
		FgColor: "black",
		PacketHandlers: hub.PacketHandlers{
			"/update": &hub.PacketHandler[qrcode]{q.update},
		},
		APIs: hub.APIs{
			"POST /generate":    q.generate,
			"GET /edit-content": q.editContent,
		},
		FuncMap: template.FuncMap{
			"png": q.png,
		},
	}
}

func (q *qrcode) Setup() error { return nil }

func (q *qrcode) update(pkt *hub.Packet) {
	pkt.Unmarshal(q).RouteUp()
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

func (q *qrcode) editContent(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if err := hub.RenderTemplate(w, id, "edit-content.tmpl", nil); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
