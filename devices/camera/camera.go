package camera

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"image"
	"image/draw"
	"image/jpeg"
	"os"
	"path/filepath"
	"time"

	"github.com/merliot/hub/devices/camera/cache"
	"github.com/merliot/hub/pkg/device"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

//go:embed *.go images template
var embedFS embed.FS

type camera struct {
	*cache.Cache
}

type msgGetImage struct {
	Index uint32 `mcp:"required,desc=Image index"`
}

func (m msgGetImage) Desc() string {
	return "Get the camera image"
}

type msgImage struct {
	Jpeg []byte
	Prev uint32
	Next uint32
	Err  string
}

func NewModel() device.Devicer {
	return &camera{Cache: cache.New(maxMemoryFiles, maxFiles)}
}

func (c *camera) GetConfig() device.Config {
	return device.Config{
		Model:   "camera",
		Parents: []string{"hub"},
		State:   c,
		FS:      &embedFS,
		Targets: []string{"rpi", "x86-64"},
		BgColor: "blue",
		FgColor: "black",
		PacketHandlers: device.PacketHandlers{
			"/get-image": &device.PacketHandler[msgGetImage]{c.getImage},
			"image":      &device.PacketHandler[msgImage]{device.RouteUp},
		},
		FuncMap: device.FuncMap{
			"jpeg": c.jpeg,
		},
	}
}

func (c *camera) getImage(pkt *device.Packet) {
	var msgGet msgGetImage
	var msgImage msgImage
	var err error

	pkt.Unmarshal(&msgGet)
	msgImage.Jpeg, msgImage.Prev, msgImage.Next, err = c.Cache.GetJpeg(msgGet.Index)
	if err != nil {
		msgImage.Err = err.Error()
	}
	pkt.SetPath("image").Marshal(&msgImage).RouteUp()
}

func (c *camera) jpeg(raw string) (template.URL, error) {
	url := fmt.Sprintf("data:image/jpeg;base64,%s", raw)
	// Return it as template-safe url to use with <img src={{.}}>
	return template.URL(url), nil
}

func (c *camera) Setup() error {
	if err := c.Cache.Preload(); err != nil {
		return err
	}
	go c.captureAndSave()
	return nil
}

func (c *camera) captureAndSave() {

	var rawSpec = "raw-%d.jpg"
	var rawGlob = "raw-*.jpg"

	// Clean up any old raw files
	matches, err := filepath.Glob(rawGlob)
	if err == nil {
		for _, fileName := range matches {
			os.Remove(fileName)
		}
	}

	// Start the capture process in the background, continuously capturing
	// camera images to raw files
	if err := startCapture(rawSpec); err != nil {
		println("Failed to start camera capture process:", err)
		return
	}

	// When a captured raw file shows up, add watermark, save to cache, and
	// delete raw file
	for {
		time.Sleep(time.Second)
		matches, err := filepath.Glob(rawGlob)
		if err != nil {
			println("Failed to find any raw files:", err)
			continue
		}
		time.Sleep(time.Second)
		for _, fileName := range matches {
			fileInfo, err := os.Stat(fileName)
			if err == nil {
				jpeg, err := watermark(fileName, fileInfo.ModTime())
				if err == nil {
					c.Cache.SaveJpeg(jpeg)
				} else {
					println("Failed to add watermark",
						"fileName", fileName, "err", err)
				}
			}
			os.Remove(fileName)
		}
	}
}

func watermark(fileName string, modTime time.Time) ([]byte, error) {

	// Open the image file
	imgFile, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("error opening image: %w", err)
	}
	defer imgFile.Close()

	// Decode the image
	img, err := jpeg.Decode(imgFile)
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %w", err)
	}

	// Create a new RGBA image for drawing
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, image.Point{0, 0}, draw.Src)

	// Draw a timestamp watermark in the lower-right
	d := &font.Drawer{
		Dst:  rgba,
		Src:  image.White,
		Face: basicfont.Face7x13,
	}
	d.Dot = fixed.Point26_6{
		X: fixed.I(bounds.Max.X - 200),
		Y: fixed.I(bounds.Max.Y - 20),
	}
	timestamp := modTime.Format("2006-01-02 15:04:05")
	d.DrawString(timestamp)

	// Encode the image to JPEG bytes
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, rgba, &jpeg.Options{Quality: 90})
	if err != nil {
		return nil, fmt.Errorf("error encoding image: %w", err)
	}

	return buf.Bytes(), nil
}

func (c *camera) Poll(pkt *device.Packet)     {}
func (c *camera) DemoSetup() error            { return c.Setup() }
func (c *camera) DemoPoll(pkt *device.Packet) { c.Poll(pkt) }
