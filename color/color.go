// Package color provides tools for defining and generating high-contrast
// color palettes suitable for QR codes and other visual elements.
package color

import (
	"image/color"
	"math"
	"sync"
)

// Name defines a custom type for palette identification.
type Name string

// Name returns the string representation of the color name.
// This is part of the Color interface implementation.
func (n Name) Name() string {
	return string(n)
}

// Palette returns the foreground and background RGBA values for the specific color name.
// This is part of the Color interface implementation.
func (n Name) Palette() (foreground color.RGBA, background color.RGBA) {
	p := NewPalette(n)
	return p.ForeGround, p.BackGround
}

const (
	Black   Name = "black"
	White   Name = "white"
	Gray    Name = "gray"
	Blue    Name = "blue"
	Red     Name = "red"
	Green   Name = "green"
	Yellow  Name = "yellow"
	Magenta Name = "magenta"
	Cyan    Name = "cyan"
	Orange  Name = "orange"
	Purple  Name = "purple"
	Gold    Name = "gold"
	Silver  Name = "silver"
	Pink    Name = "pink"
	Brown   Name = "brown"
	Navy    Name = "navy"
	Lime    Name = "lime"
	Teal    Name = "teal"
	Indigo  Name = "indigo"
	Violet  Name = "violet"
	Crimson Name = "crimson"
	Olive   Name = "olive"
	Maroon  Name = "maroon"
	SkyBlue Name = "skyblue"
	Rose    Name = "rose"
	Emerald Name = "emerald"
)

// Palette groups the foreground and background colors for the palette.
type Palette struct {
	ForeGround color.RGBA
	BackGround color.RGBA
}

// DefaultPalette defines the default system colors (black on light gray).
var DefaultPalette = Palette{
	ForeGround: color.RGBA{0x00, 0x00, 0x00, 0xff}, // Black
	BackGround: color.RGBA{0xef, 0xef, 0xef, 0xff}, // Light Gray
}

var (
	// mu protects access to the registry map.
	mu sync.RWMutex
	// registry stores all available palettes.
	registry = make(map[Name]Palette)
)

func init() {
	// Initialize the registry with default values.
	Register(Blue, Palette{ForeGround: color.RGBA{0x00, 0x5a, 0x9c, 0xff}, BackGround: color.RGBA{0xff, 0xff, 0xff, 0xff}})
	Register(Red, Palette{ForeGround: color.RGBA{0xc4, 0x1e, 0x3a, 0xff}, BackGround: color.RGBA{0xff, 0xff, 0xff, 0xff}})
	Register(Green, Palette{ForeGround: color.RGBA{0x00, 0x87, 0x3e, 0xff}, BackGround: color.RGBA{0xff, 0xff, 0xff, 0xff}})
	Register(Yellow, Palette{ForeGround: color.RGBA{0xfd, 0xd8, 0x35, 0xff}, BackGround: color.RGBA{0x00, 0x00, 0x00, 0xff}})
	Register(Magenta, Palette{ForeGround: color.RGBA{0xff, 0x00, 0xff, 0xff}, BackGround: color.RGBA{0xff, 0xff, 0xff, 0xff}})
	Register(Cyan, Palette{ForeGround: color.RGBA{0x00, 0xff, 0xff, 0xff}, BackGround: color.RGBA{0x00, 0x00, 0x00, 0xff}})
	Register(Orange, Palette{ForeGround: color.RGBA{0xff, 0x8c, 0x00, 0xff}, BackGround: color.RGBA{0xff, 0xff, 0xff, 0xff}})
	Register(Purple, Palette{ForeGround: color.RGBA{0x80, 0x00, 0x80, 0xff}, BackGround: color.RGBA{0xff, 0xff, 0xff, 0xff}})
	Register(Gold, Palette{ForeGround: color.RGBA{0xd4, 0xaf, 0x37, 0xff}, BackGround: color.RGBA{0x00, 0x00, 0x00, 0xff}})
	Register(Silver, Palette{ForeGround: color.RGBA{0xc0, 0xc0, 0xc0, 0xff}, BackGround: color.RGBA{0x00, 0x00, 0x00, 0xff}})
	Register(Black, Palette{ForeGround: color.RGBA{0x00, 0x00, 0x00, 0xff}, BackGround: color.RGBA{0xff, 0xff, 0xff, 0xff}})
	Register(White, Palette{ForeGround: color.RGBA{0xff, 0xff, 0xff, 0xff}, BackGround: color.RGBA{0x00, 0x00, 0x00, 0xff}})
	Register(Gray, Palette{ForeGround: color.RGBA{0x80, 0x80, 0x80, 0xff}, BackGround: color.RGBA{0xff, 0xff, 0xff, 0xff}})
	Register(Pink, Palette{ForeGround: color.RGBA{0xff, 0xc0, 0xcb, 0xff}, BackGround: color.RGBA{0x00, 0x00, 0x00, 0xff}})
	Register(Brown, Palette{ForeGround: color.RGBA{0x8b, 0x45, 0x13, 0xff}, BackGround: color.RGBA{0xff, 0xff, 0xff, 0xff}})
	Register(Navy, Palette{ForeGround: color.RGBA{0x00, 0x00, 0x80, 0xff}, BackGround: color.RGBA{0xff, 0xff, 0xff, 0xff}})
	Register(Lime, Palette{ForeGround: color.RGBA{0x32, 0xcd, 0x32, 0xff}, BackGround: color.RGBA{0x00, 0x00, 0x00, 0xff}})
	Register(Teal, Palette{ForeGround: color.RGBA{0x00, 0x80, 0x80, 0xff}, BackGround: color.RGBA{0xff, 0xff, 0xff, 0xff}})
	Register(Indigo, Palette{ForeGround: color.RGBA{0x4b, 0x00, 0x82, 0xff}, BackGround: color.RGBA{0xff, 0xff, 0xff, 0xff}})
	Register(Violet, Palette{ForeGround: color.RGBA{0xee, 0x82, 0xee, 0xff}, BackGround: color.RGBA{0x00, 0x00, 0x00, 0xff}})
	Register(Crimson, Palette{ForeGround: color.RGBA{0xdc, 0x14, 0x3c, 0xff}, BackGround: color.RGBA{0xff, 0xff, 0xff, 0xff}})
	Register(Olive, Palette{ForeGround: color.RGBA{0x80, 0x80, 0x00, 0xff}, BackGround: color.RGBA{0xff, 0xff, 0xff, 0xff}})
	Register(Maroon, Palette{ForeGround: color.RGBA{0x80, 0x00, 0x00, 0xff}, BackGround: color.RGBA{0xff, 0xff, 0xff, 0xff}})
	Register(SkyBlue, Palette{ForeGround: color.RGBA{0x87, 0xce, 0xeb, 0xff}, BackGround: color.RGBA{0x00, 0x00, 0x00, 0xff}})
	Register(Rose, Palette{ForeGround: color.RGBA{0xff, 0x00, 0x7f, 0xff}, BackGround: color.RGBA{0xff, 0xff, 0xff, 0xff}})
	Register(Emerald, Palette{ForeGround: color.RGBA{0x50, 0xc8, 0x78, 0xff}, BackGround: color.RGBA{0x00, 0x00, 0x00, 0xff}})
}

// Register adds a new palette to the registry.
// This allows external packages to inject new color cases as plugins.
func Register(name Name, p Palette) {
	mu.Lock()
	defer mu.Unlock()
	registry[name] = p
}

// NewPalette returns a Palette based on a Name.
// It searches the registry for the name; if not found, it returns the DefaultPalette.
func NewPalette(name Name) Palette {
	mu.RLock()
	defer mu.RUnlock()
	if p, ok := registry[name]; ok {
		return p
	}
	return DefaultPalette
}


// DefaultColors returns a list of all built-in Color objects currently registered.
func DefaultColors() []Color {
	mu.RLock()
	defer mu.RUnlock()

	colors := make([]Color, 0, len(registry))
	for name := range registry {
		colors = append(colors, name)
	}
	return colors
}


// Color defines the behavior for color objects.
type Color interface {
	// Name returns the descriptive name of the color.
	Name() string
	// Palette returns the foreground and background RGBA values.
	Palette() (foreground color.RGBA, background color.RGBA)
}

// IdentifyColor finds the closest color from a list of Color objects using Euclidean distance.
func IdentifyColor(input color.Color, colors []Color) string {
	r1, g1, b1, _ := input.RGBA()

	// Convert 16-bit color components to 8-bit (0-255) for comparison.
	r1h, g1h, b1h := uint8(r1>>8), uint8(g1>>8), uint8(b1>>8)

	var closestColor string
	minDistance := math.MaxFloat64

	for _, color := range colors {
		foreground, _ := color.Palette()

		// Euclidean distance formula: sqrt((r2-r1)^2 + (g2-g1)^2 + (b2-b1)^2)
		dist := math.Sqrt(
			math.Pow(float64(foreground.R)-float64(r1h), 2) +
				math.Pow(float64(foreground.G)-float64(g1h), 2) +
				math.Pow(float64(foreground.B)-float64(b1h), 2),
		)

		if dist < minDistance {
			minDistance = dist
			closestColor = color.Name()
		}
	}

	return closestColor
}
