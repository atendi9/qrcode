# QrCode Encoder

<img src='./logo.png' width=300/>

- provides primitives for generating and manipulating QR codes, including support for embedding center icons and exporting to multiple formats.

## Usage

```go
package main

import (
	"os"

	"github.com/atendi9/qrcode"
)

func main() {
	content := "https://github.com/atendi9/capivara"
	qr, err := qrcode.Encode(content)
	if err != nil {
		panic(err)
	}
	f, err := os.Create("qrcode.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	logo, err := os.Open("logo.png")
	if err != nil {
		panic(err)
	}
	defer logo.Close()
	qr.SetIcon(logo)
	mime := qrcode.PNG
	if err := qr.Scale(500, 500).Image(f, mime); err != nil {
		panic(err)
	}
}
```