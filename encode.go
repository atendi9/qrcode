// Package qrcode provides primitives for generating and manipulating QR codes,
// including support for embedding center icons and exporting to multiple formats.
package qrcode

import (
	"image"

	"github.com/atendi9/qrcode/color"
	"github.com/atendi9/qrcode/encoder"
	"golang.org/x/image/draw"
)

// Encode generates a high-recovery (High) QR Code from the given content.
// The options parameter is variadic and optional, allowing customization
// of the QR code's size and color palette.
//
// If no options are provided, or if the specified size is less than 256,
// it defaults to a size of 1024x1024 pixels. If no custom color is specified
// within the options, [color.DefaultPalette] is used for the foreground and background.
//
// It returns a QRCode interface containing the generated RGBA image and the original
// content, or an error if the generation fails.
func Encode(content string, options ...Options) (QRCode, error) {
	p := color.DefaultPalette
	size := 1024
	if len(options) > 0 {
		cfg := options[0]
		if cfg.Color != nil {
			p.ForeGround, p.BackGround = cfg.Color.Palette()
		}
		if cfg.Size >= 256 {
			size = cfg.Size
		}
	}

	q, err := encoder.NewQRCode(content, encoder.High)
	if err != nil {
		return nil, err
	}

	img := q.WithColors(p.ForeGround, p.BackGround).Image(size)

	bounds := img.Bounds()
	rgbaImg := image.NewRGBA(bounds)

	draw.Draw(rgbaImg, bounds, img, bounds.Min, draw.Src)

	return &qrCode{
		RGBA:    rgbaImg,
		content: content,
	}, nil
}
