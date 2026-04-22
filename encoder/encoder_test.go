package encoder

import (
	"bytes"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/capivara/langs"
)

func TestNewQRCode(t *testing.T) {
	a := assert.New(langs.EN_US, t)

	content := "https://example.com"
	level := RecoveryLevel(0)

	q, err := NewQRCode(content, level)

	assert.Equal(a, err, nil)
	assert.Equal(a, q.Content, content)
	assert.Equal(a, q.Level, level)
}

func TestNewWithForcedVersion(t *testing.T) {
	a := assert.New(langs.EN_US, t)

	content := "Forced Version Test Data"
	level := RecoveryLevel(0)
	version := 10

	q, err := NewWithForcedVersion(content, version, level)

	assert.Equal(a, err, nil)
	assert.Equal(a, q.Content, content)
	assert.Equal(a, q.VersionNumber, version)
}

func TestEncode(t *testing.T) {
	a := assert.New(langs.EN_US, t)

	content := "Encode PNG Test"
	level := RecoveryLevel(0)
	size := 256

	pngBytes, err := Encode(content, level, size)

	assert.Equal(a, err, nil)

	expectedMagic := []byte{137, 80, 78, 71, 13, 10, 26, 10}
	var actualMagic []byte
	if len(pngBytes) >= 8 {
		actualMagic = pngBytes[:8]
	}
	assert.Equal(a, string(actualMagic), string(expectedMagic))
}

func TestWriteFile(t *testing.T) {
	a := assert.New(langs.EN_US, t)

	content := "Write File Test"
	level := RecoveryLevel(0)
	size := 256

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test_qr.png")

	err := WriteFile(content, level, size, filename)
	assert.Equal(a, err, nil)

	_, err = os.Stat(filename)
	assert.Equal(a, err, nil)
}

func TestWriteColorFile(t *testing.T) {
	a := assert.New(langs.EN_US, t)

	content := "Write Color File Test"
	level := RecoveryLevel(0)
	size := 256
	bg := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	fg := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test_qr_color.png")

	err := WriteColorFile(content, level, size, bg, fg, filename)
	assert.Equal(a, err, nil)

	_, err = os.Stat(filename)
	assert.Equal(a, err, nil)
}

func TestQRCode_Bitmap(t *testing.T) {
	a := assert.New(langs.EN_US, t)

	q, _ := NewQRCode("Bitmap Test Data", RecoveryLevel(0))
	bitmap := q.Bitmap()

	isMatrixPopulated := len(bitmap) > 0 && len(bitmap[0]) > 0
	assert.Equal(a, isMatrixPopulated, true)
}

func TestQRCode_Image(t *testing.T) {
	a := assert.New(langs.EN_US, t)

	q, _ := NewQRCode("Image Size Test", RecoveryLevel(0))
	img := q.Image(256)

	bounds := img.Bounds()
	assert.Equal(a, bounds.Dx(), 256)
	assert.Equal(a, bounds.Dy(), 256)
}

func TestQRCode_PNG(t *testing.T) {
	a := assert.New(langs.EN_US, t)

	q, _ := NewQRCode("Struct PNG Test", RecoveryLevel(0))
	pngBytes := q.PNG(256)

	expectedMagic := []byte{137, 80, 78, 71, 13, 10, 26, 10}
	var actualMagic []byte
	if len(pngBytes) >= 8 {
		actualMagic = pngBytes[:8]
	}
	assert.Equal(a, string(actualMagic), string(expectedMagic))
}

func TestQRCode_Write(t *testing.T) {
	a := assert.New(langs.EN_US, t)

	q, _ := NewQRCode("io.Writer Test", RecoveryLevel(0))
	var buf bytes.Buffer

	err := q.Write(256, &buf)
	assert.Equal(a, err, nil)

	isBufferPopulated := buf.Len() > 0
	assert.Equal(a, isBufferPopulated, true)
}
func TestQRCode_WriteFile(t *testing.T) {
	a := assert.New(langs.EN_US, t)

	q, _ := NewQRCode("Struct WriteFile Test", RecoveryLevel(0))

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test_obj_write.png")

	err := q.WriteFile(256, filename)
	assert.Equal(a, err, nil)

	_, err = os.Stat(filename)
	assert.Equal(a, err, nil)
}

func TestQRCode_WriteFileWithoutSize(t *testing.T) {
	a := assert.New(langs.EN_US, t)

	q, _ := NewQRCode("Default Size WriteFile Test", RecoveryLevel(0))

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test_obj_write_default.png")

	err := q.WriteFileWithoutSize(filename)
	assert.Equal(a, err, nil)

	_, err = os.Stat(filename)
	assert.Equal(a, err, nil)
}

func TestQRCode_WithColors(t *testing.T) {
	a := assert.New(langs.EN_US, t)

	q, _ := NewQRCode("Fluent Colors Test", RecoveryLevel(0))

	bg := color.RGBA{R: 10, G: 20, B: 30, A: 255}
	fg := color.RGBA{R: 40, G: 50, B: 60, A: 255}

	q = q.WithColors(fg, bg)

	assert.Equal(a, fmt.Sprint(q.ForegroundColor), fmt.Sprint(fg))
	assert.Equal(a, fmt.Sprint(q.BackgroundColor), fmt.Sprint(bg))
}

func TestQRCode_WithNoBorder(t *testing.T) {
	a := assert.New(langs.EN_US, t)

	q, _ := NewQRCode("Fluent No Border Test", RecoveryLevel(0))

	q = q.WithNoBorder()

	assert.Equal(a, q.DisableBorder, true)
}

func TestQRCode_ToString(t *testing.T) {
	a := assert.New(langs.EN_US, t)

	q, _ := NewQRCode("Terminal ToString Test", RecoveryLevel(0))

	str := q.ToString(false)

	isStrPopulated := len(str) > 0
	assert.Equal(a, isStrPopulated, true)
}

func TestQRCode_ToSmallString(t *testing.T) {
	a := assert.New(langs.EN_US, t)

	q, _ := NewQRCode("Terminal ToSmallString Test", RecoveryLevel(0))

	str := q.ToSmallString(false)

	isStrPopulated := len(str) > 0
	assert.Equal(a, isStrPopulated, true)
}
