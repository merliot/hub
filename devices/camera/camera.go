package camera

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"os/exec"

	"github.com/merliot/hub/pkg/device"
)

//go:embed *.go images template
var embedFS embed.FS

type camera struct {
	getJpeg func(int) ([]byte, error)
}

type msgGetImage struct {
	Index int
}

type msgImage struct {
	Jpeg []byte
}

func NewModel() device.Devicer {
	return &camera{}
}

func (c *camera) GetConfig() device.Config {
	return device.Config{
		Model:   "camera",
		State:   c,
		FS:      &embedFS,
		Targets: []string{"rpi"},
		BgColor: "almond-creme",
		FgColor: "black",
		PacketHandlers: device.PacketHandlers{
			"/get-image": &device.PacketHandler[msgGetImage]{c.getImage},
			"/image":     &device.PacketHandler[msgImage]{device.RouteUp},
		},
		FuncMap: template.FuncMap{
			"jpeg": c.jpeg,
		},
	}
}

func (c *camera) getImage(pkt *device.Packet) {
	var msgGet msgGetImage
	var msgImage msgImage
	var err error

	pkt.Unmarshal(&msgGet)
	msgImage.Jpeg, err = c.getJpeg(msgGet.Index)
	if err == nil {
		pkt.SetPath("/image").Marshal(&msgImage).RouteUp()
	} else {
		println(err.Error())
	}
}

func (c *camera) jpeg(raw string) (template.URL, error) {
	// Convert the image to base64
	//base64Image := base64.StdEncoding.EncodeToString([]byte(raw))
	url := fmt.Sprintf("data:image/jpeg;base64,%s", raw)
	// Return it as template-safe url to use with <img src={{.}}>
	return template.URL(url), nil
}

func (c *camera) rawJpeg(index int) ([]byte, error) {
	// Define the command to capture the image using `raspistill`
	// -o - : Outputs the image to stdout
	// -t 1 : Takes the picture immediately
	cmd := exec.Command("rpicam-still", "-o", "-", "-t", "1", "--width", "640", "--height", "480")

	// Create a buffer to store the image
	var out bytes.Buffer
	cmd.Stdout = &out

	// Run the command
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to capture image: %w", err)
	}

	// Return the captured image as a byte slice
	return out.Bytes(), nil
}

func (c *camera) Setup() error {
	c.getJpeg = c.rawJpeg
	return nil
}

func (c *camera) Poll(pkt *device.Packet) {}
