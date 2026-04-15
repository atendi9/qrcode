// Package qrcode provides primitives for generating and manipulating QR codes,
// including support for embedding center icons and exporting to multiple formats.
package qrcode

import (
	"errors"
	"image"
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
	SetIcon(icon io.Reader)
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

// SetIcon reads an image from iconReader and draws it centered on the QR code.
func (q *qrCode) SetIcon(iconReader io.Reader) {
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

	draw.Draw(q.RGBA, rect, image.Black, image.Point{}, draw.Src) 

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
