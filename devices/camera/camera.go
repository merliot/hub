package camera

import (
	"embed"
	"fmt"
        "image"
        "image/color"
        "image/draw"
        "image/jpeg"

	"html/template"
	"time"

	"github.com/merliot/hub/devices/camera/cache"
	"github.com/merliot/hub/pkg/device"

        "golang.org/x/image/font"
        "golang.org/x/image/font/basicfont"
)

//go:embed *.go images template
var embedFS embed.FS

type camera struct {
	*cache.Cache
}

type msgGetImage struct {
	Index uint32
}

type msgImage struct {
	Jpeg []byte
	Prev uint32
	Next uint32
}

func NewModel() device.Devicer {
	return &camera{Cache: cache.New(maxMemoryFiles, maxFiles)}
}

func (c *camera) GetConfig() device.Config {
	return device.Config{
		Model:      "camera",
		Parents:    []string{"hub"},
		State:      c,
		FS:         &embedFS,
		Targets:    []string{"rpi", "x86-64"},
		BgColor:    "blue",
		FgColor:    "black",
		PollPeriod: 5 * time.Second,
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
	msgImage.Jpeg, msgImage.Prev, msgImage.Next, err = c.Cache.GetJpeg(msgGet.Index)
	if err == nil {
		pkt.SetPath("/image").Marshal(&msgImage).RouteUp()
	} else {
		println(err.Error())
	}
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
	go c.captureSave()
}

func (c *capture) captureSave() {

	// Command to execute libcamera-still
	cmd := exec.Command(
		"libcamera-still",
		"-t", "0", // Continuous capture
		"--timelapse", "5s", // Capture every 5 seconds
		"-o", "-", // Output to stdout
	)

	// Capture output from stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to get stdout pipe: %v", err)
	}

	// Start the libcamera-still process
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start libcamera-still: %v", err)
	}

	// Process the JPEG stream
	var buffer bytes.Buffer
	var chunk = make([]byte, 100 * 1024)

	for {
		// Read chunks from stdout
		n, err := stdout.Read(chunk)
		if err != nil {
			log.Fatalf("Error reading from stdout: %v", err)
		}

		// Append the chunk to the buffer
		buffer.Write(chunk[:n])
		println("n=", n)

		// Look for the EOI marker (0xFF 0xD9)
		for {
			data := buffer.Bytes()
			eoiIndex := bytes.Index(data, []byte{0xFF, 0xD9})
			if eoiIndex == -1 {
				// EOI marker not found, wait for more data
				break
			}
			println("found EOI")

			// Extract a complete JPEG image (SOI to EOI)
			jpegData := data[:eoiIndex+2]
			buffer.Next(eoiIndex + 2) // Remove the processed image from the buffer

			// Process the JPEG image
			img, _, err := image.Decode(bytes.NewReader(jpegData))
			if err != nil {
				log.Printf("Error decoding JPEG image: %v", err)
				continue
			}

			// Add a timestamp watermark
			timestampedImg := addTimestamp(img)

			// Save the image to disk
			c.Cache.SaveJpeg(timestampedImg)
		}
	}
}

func addWatermark(img image.Image, text string) {
        bounds := img.Bounds()
        // Create a new RGBA image for drawing (important for text rendering)
        rgba := image.NewRGBA(bounds)
        draw.Draw(rgba, bounds, img, image.Point{0, 0}) // Copy original image

        // Set font and color
        col := color.RGBA{255, 255, 255, 255} // White color
        // face := inconsolata.Regular // Example font (you might need to install a font)
        face := basicfont.Face // Use basic font
        // Calculate text position (example: bottom-right corner)
        point := image.Pt(bounds.Max.X-200, bounds.Max.Y-20) // Adjust position as needed

        // Draw the text
        d := &font.Drawer{
                Dst:  rgba,
                Src:  image.NewUniform(col),
                Face: face,
                // ... other font drawing options
        }

        d.DrawString(text, point)

        // Draw the new image with the watermark over the original
        draw.Draw(img, bounds, rgba, image.Point{0, 0})
}

// addTimestamp adds a timestamp watermark to the image
func addTimestamp(img image.Image) image.Image {
	const fontSize = 24

	// Create a new context with the image
	bounds := img.Bounds()
	dc := gg.NewContext(bounds.Dx(), bounds.Dy())
	dc.DrawImage(img, 0, 0)

	// Add the timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	dc.SetRGB(1, 1, 1) // White text
	dc.DrawStringAnchored(timestamp, float64(bounds.Dx())-10,
		float64(bounds.Dy())-10, 1, 1) // Bottom-right corner

	return dc.Image()
}

func (c *camera) Poll(pkt *device.Packet) {}
func (c *camera) DemoSetup() error            { return c.Setup() }
func (c *camera) DemoPoll(pkt *device.Packet) { c.Poll(pkt) }
