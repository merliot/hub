//go:build !tinygo

package qrcode

import (
	"embed"
	"encoding/base64"
	"html/template"
	"net/http"

	goqr "github.com/skip2/go-qrcode"
)

//go:embed *.go template
var fs embed.FS

func (q *qrcode) generate(w http.ResponseWriter, r *http.Request) {

	content := r.FormValue("Content")
	if content == "" {
		return
	}

	// Generate the QR code
	var png []byte
	png, err := goqr.Encode(content, goqr.Medium, 208)
	if err != nil {
		http.Error(w, "Error generating QR code", http.StatusInternalServerError)
		return
	}

	// Convert the image to base64
	base64Image := base64.StdEncoding.EncodeToString(png)

	// Embed the image in response
	tmpl := template.Must(template.New("qr").Parse(`<img src="data:image/png;base64,{{.}}">`))
	tmpl.Execute(w, base64Image)
}
