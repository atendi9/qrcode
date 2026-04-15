// Package qrcode provides primitives for generating and manipulating QR codes,
// including support for embedding center icons and exporting to multiple formats.
package qrcode

import (
	"image"

	"github.com/i9si-sistemas/qrcode"
)

func Encode(content string) (QRCode, error) {
	q, err := qrcode.New(content, qrcode.High)
	if err != nil {
		return nil, err
	}

	img := q.Image(256) 
	bounds := img.Bounds()
	grayImg := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			grayImg.Set(x, y, img.At(x, y))
		}
	}

	return &qrCode{
		RGBA:    grayImg,
		content: content,
	}, nil
}
