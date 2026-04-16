// Package qrcode provides primitives for generating and manipulating QR codes,
// including support for embedding center icons and exporting to multiple formats.
package qrcode

import (
	"image"

	"github.com/atendi9/qrcode/color"
	"github.com/i9si-sistemas/qrcode"
)

// Encode generates a QR Code from the given content.
// The palette parameter is optional; if omitted, [color.DefaultPalette] is used.
func Encode(content string, palette ...color.Color) (QRCode, error) {
	p := color.DefaultPalette

	if len(palette) > 0 {
		p.ForeGround, p.BackGround = palette[0].Pallete()
	}

	q, err := qrcode.New(content, qrcode.High)
	if err != nil {
		return nil, err
	}

	img := q.WithColors(p.ForeGround, p.BackGround).Image(256)

	bounds := img.Bounds()
	rgbaImg := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgbaImg.Set(x, y, img.At(x, y))
		}
	}

	return &qrCode{
		RGBA:    rgbaImg,
		content: content,
	}, nil
}
