package qrcode

import (
	"fmt"
	"os"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/capivara/langs"
	"github.com/atendi9/qrcode/color"
)

func TestEncode(t *testing.T) {
	t.Run("Deve codificar um conteúdo válido com sucesso", func(v *testing.T) {
		t := assert.New(langs.EN_US, v)
		content := "https://github.com/atendi9/capivara"
		qr, err := Encode(content)
		assert.Empty(t, err)
		assert.NotNil(t, qr)
		assert.Equal(t, content, qr.Content())
		filePath := fmt.Sprintf("%s/image.png", v.TempDir())
		err = saveFile(qr, filePath, PNG)
		assert.NoError(t, err)
		filePath = fmt.Sprintf("%s/image.jpg", v.TempDir())
		err = saveFile(qr, filePath, JPG)
		assert.NoError(t, err)
		filePath = fmt.Sprintf("%s/image.jpg", v.TempDir())
		err = saveFile(qr, filePath, Mime("image/webp"))
		assert.ErrorIs(t, err, ErrInvalidMime)
		meta := qr.Metadata()
		assert.Equal(t, 2, meta.Dimensions)

		assert.NotNil(t, qr.Bounds())
		assert.True(t, qr.Bounds().Dx() > 0)
		assert.True(t, qr.Bounds().Dy() > 0)
	})

	t.Run("Deve retornar erro ao tentar codificar conteúdo vazio", func(v *testing.T) {
		t := assert.New(langs.EN_US, v)
		content := ""

		qr, err := Encode(content, color.Blue)
		if err != nil {
			assert.Empty(t, qr)
			assert.NotNil(t, err)
		}
	})
}

func saveFile(qr QRCode, filePath string, mime Mime) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	logo, err := os.Open("atendi9_logo.png")
	if err != nil {
		return err
	}
	defer logo.Close()
	qr.SetIcon(logo)
	if err := qr.Scale(500, 500).Image(f, mime); err != nil {
		return err
	}
	return nil
}
