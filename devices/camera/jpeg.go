package camera

import (
        "bytes"
        "fmt"
        "image"
        "image/color"
        "image/draw"
        "image/jpeg"
        "io/ioutil"
        "os"
        "path/filepath"
        "sort"
        "time"

        "golang.org/x/image/font"
        "golang.org/x/image/font/basicfont"
)

type fileInfo struct {
        name    string
        modTime time.Time
}

type ByModTime []fileInfo

func (a ByModTime) Len() int           { return len(a) }
func (a ByModTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByModTime) Less(i, j int) bool { return a[i].modTime.After(a[j].modTime) } // Newest first

// getJPEGFileWithWatermark adds a timestamp watermark to jpeg file at index
func getJPEGFileWithWatermark(index int) (jpegData []byte, prev, next int, err error) {
        name, prev, next := getFile(index)

        if name == "" {
                return nil, 0, 0, fmt.Errorf("file not found at index %d", index)
        }

        filePath := name

        imageData, err := ioutil.ReadFile(filePath)
        if err != nil {
                return nil, 0, 0, fmt.Errorf("error reading file %s: %w", filePath, err)
        }

        img, err := jpeg.Decode(nil, imageData)
        if err != nil {
                return nil, 0, 0, fmt.Errorf("error decoding JPEG file %s: %w", filePath, err)
        }

        // Add watermark
        timestamp := time.Now().Format("2006-01-02 15:04:05")
        addWatermark(img, timestamp)

        // Encode the image back to JPEG with the watermark
        buf := new(bytes.Buffer) // Use a buffer to store the encoded JPEG data
        err = jpeg.Encode(buf, img, nil) // You can adjust the JPEG quality here (e.g., &jpeg.Options{Quality: 80})
        if err != nil {
                return nil, 0, 0, fmt.Errorf("error encoding JPEG: %w", err)
        }

        jpegData = buf.Bytes()

        return jpegData, prev, next, nil
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

// GetFile retrieves information about a JPG file in the current directory,
// sorted by creation time (newest first).
func GetFile(index int) (name string, prev, next int) {
        files, err := ioutil.ReadDir(".")
        if err != nil {
                fmt.Println("Error reading directory:", err)
                return "", 0, 0 // Or handle the error differently
        }

        var jpgFiles []fileInfo
        for _, file := range files {
                if !file.IsDir() && filepath.Ext(file.Name()) == ".jpg" {
                        info, err := file.Info()
                        if err != nil {
                                fmt.Println("Error getting file info:", err)
                                continue // Skip files with errors
                        }
                        jpgFiles = append(jpgFiles, fileInfo{name: file.Name(), modTime: info.ModTime()})
                }
        }

        sort.Sort(ByModTime(jpgFiles))

        if index < 0 || index >= len(jpgFiles) {
                return "", 0, 0 // Index out of range
        }

        name = jpgFiles[index].name

        if index > 0 {
                prev = index - 1
        }

        if index < len(jpgFiles)-1 {
                next = index + 1
        }

        return name, prev, next
}
