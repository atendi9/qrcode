// Package qrcode provides primitives for generating and manipulating QR codes,
// including support for embedding center icons and exporting to multiple formats.
package qrcode

import (
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"

	"golang.org/x/image/draw"
)

// QRCode represents a QR code image with its associated content and metadata.
// It implements the image.Image interface.
type QRCode interface {
	image.Image
	// Content returns the raw string data encoded in the QR code.
	Content() string
	// Metadata returns technical details about the QR code structure.
	Metadata() Metadata
	// SetIcon overlays an image (logo) in the center of the QR code.
	// It handles the icon scaling and optional background box rendering.
	SetIcon(icon io.Reader, box ...IconBox)
	// Image encodes the QR code into a specific Mime format and writes it to w.
	Image(w io.Writer, m Mime) error
	// Scale resizes the current QR Code to a new width and height.
	Scale(width, height int) QRCode
}

// Metadata holds structural information about the generated QR code.
type Metadata struct {
	Dimensions int
}

// qrCode is the internal implementation of the QRCode interface.
type qrCode struct {
	*image.RGBA
	content string
}

// Content returns the string content of the QR code.
func (q *qrCode) Content() string { return q.content }

// Mime defines the supported image output formats.
type Mime string

const (
	// PNG represents the Portable Network Graphics format.
	PNG Mime = "image/png"
	// JPG represents the Joint Photographic Experts Group format.
	JPG Mime = "image/jpeg"
)

// ErrInvalidMime is returned when an unsupported Mime type is provided.
var ErrInvalidMime = errors.New("invalid mime type")

// Image encodes the QR code using the specified Mime type.
// For JPG format, it uses a high-quality setting (99) to ensure readability.
func (q *qrCode) Image(w io.Writer, m Mime) error {
	switch m {
	case PNG:
		return png.Encode(w, q.RGBA)
	case JPG:
		return jpeg.Encode(w, q.RGBA, &jpeg.Options{
			Quality: 99,
		})
	default:
		return ErrInvalidMime
	}
}

// Metadata returns the QR code specifications.
func (q *qrCode) Metadata() Metadata { return Metadata{Dimensions: 2} }

// IconBox defines the interface for the background container of the icon.
// Any type implementing this interface can be used to provide the background image.
type IconBox interface {
	// Image returns the image.Image to be used as the icon's background.
	Image() image.Image
}

// CarbonFiberBox represents a box that generates a carbon fiber texture pattern.
type CarbonFiberBox struct {
    size int
}

// NewCarbonFiberBox creates and returns a new CarbonFiberBox with the specified size.
// It returns an IconBox interface implementation.
func NewCarbonFiberBox(size int) IconBox {
    return &CarbonFiberBox{size: size}
}

// Image generates and returns a new image.Image containing a simulated
// carbon fiber texture. The pattern is created using alternating shades of 
// dark gray blocks based on the configured size of the CarbonFiberBox.
func (c *CarbonFiberBox) Image() image.Image {
    img := image.NewRGBA(image.Rect(0, 0, c.size, c.size))
    draw.Draw(img, img.Bounds(), image.NewUniform(color.Black), image.Point{}, draw.Src)

    darkGray := color.RGBA{15, 15, 15, 255}
    vDarkGray := color.RGBA{10, 10, 10, 255}

    patternSize := 4
    for y := 0; y < c.size; y++ {
        for x := 0; x < c.size; x++ {
            if (x/patternSize+y/patternSize)%2 == 0 {
                img.Set(x, y, darkGray)
            } else {
                img.Set(x, y, vDarkGray)
            }
        }
    }

    return img
}

// SetIcon reads an image from iconReader and draws it centered on the QR code.
// It scales the icon to 20% of the QR code dimensions and applies an optional background
// provided by the IconBox interface.
func (q *qrCode) SetIcon(iconReader io.Reader, box ...IconBox) {
	icon, _, err := image.Decode(iconReader)
	if err != nil {
		return
	}

	bounds := q.Bounds()
	qrWidth := bounds.Dx()
	qrHeight := bounds.Dy()

	logoWidth := qrWidth / 5
	logoHeight := qrHeight / 5

	xOffset := (qrWidth - logoWidth) / 2
	yOffset := (qrHeight - logoHeight) / 2

	rect := image.Rect(xOffset, yOffset, xOffset+logoWidth, yOffset+logoHeight)

	// Define the background image. Defaults to image.Black if no IconBox is provided.
	var img image.Image = image.Black
	if len(box) > 0 {
		img = box[0].Image()
	}
	draw.Draw(q.RGBA, rect, img, image.Point{}, draw.Src)
	draw.ApproxBiLinear.Scale(q.RGBA, rect, icon, icon.Bounds(), draw.Over, nil)
}

// Scale resizes the current QR Code to a new width and height.
// Returns a new QR Code instance with the resized image.
func (q *qrCode) Scale(width, height int) QRCode {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.NearestNeighbor.Scale(dst, dst.Bounds(), q.RGBA, q.Bounds(), draw.Over, nil)
	return &qrCode{
		RGBA:    dst,
		content: q.content,
	}
}
