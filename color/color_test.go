package color

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/capivara/langs"
)

func TestPaletteConsistency(t *testing.T) {
	allColors := []Color{
		Black, White, Gray, Blue, Red, Green, Yellow, Magenta, Cyan, Orange,
		Purple, Gold, Silver, Pink, Brown, Navy, Lime, Teal, Indigo, Violet,
		Crimson, Olive, Maroon, SkyBlue, Rose, Emerald,
	}

	for _, color := range allColors {
		t.Run(color.Name(), func(t *testing.T) {
			a := assert.New(langs.EN_US, t)
			palette := NewPalette(Name(color.Name()))

			img := image.NewRGBA(image.Rect(0, 0, 10, 10))
			draw.Draw(img, img.Bounds(), &image.Uniform{palette.ForeGround}, image.Point{}, draw.Src)
			fileName := fmt.Sprintf("%s.png", color.Name())
			f, _ := os.Create(fileName)
			defer f.Close()
			err := png.Encode(f, img)
			assert.NoError(a, err)
			pixelColor := img.At(5, 5)
			defer os.Remove(fileName)
			identifiedName := IdentifyColor(pixelColor, allColors)

			assert.Equal(a, identifiedName, color.Name())
		})
	}
}

func TestDefaultPalette(t *testing.T) {
	palette := NewPalette("non-existent")
	as := assert.New(langs.EN_US, t)

	assert.Equal(as, palette.ForeGround.R, uint8(0))
	assert.Equal(as, palette.ForeGround.G, uint8(0))
	assert.Equal(as, palette.ForeGround.B, uint8(0))
	assert.Equal(as, palette.ForeGround.A, uint8(255))
	assert.Equal(as, palette.BackGround.R, uint8(0xef))
}

func TestPluginInjection(t *testing.T) {
	customColorName := Name("neon-green")
	customPalette := Palette{
		ForeGround: color.RGBA{0x39, 0xff, 0x14, 0xff},
		BackGround: color.RGBA{0x00, 0x00, 0x00, 0xff},
	}
	Register(customColorName, customPalette)

	result := NewPalette(customColorName)
	assert.Equal(assert.New(langs.EN_US, t), result, customPalette)
}

func TestIdentifyWithPlugin(t *testing.T) {
	brandColor := Name("brand-blue")
	palette := Palette{
		ForeGround: color.RGBA{0x12, 0x34, 0x56, 0xff},
		BackGround: color.RGBA{0xff, 0xff, 0xff, 0xff},
	}
	Register(brandColor, palette)
	input := color.RGBA{0x13, 0x35, 0x57, 0xff}

	result := IdentifyColor(input, []Color{brandColor})

	assert.Equal(assert.New(langs.EN_US, t), result, "brand-blue")
}
