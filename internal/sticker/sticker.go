package sticker

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"

	"github.com/fogleman/gg"
	"github.com/kolesa-team/go-webp/webp"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
)

//go:embed template.png
var templateBytes []byte

type StickerMaker struct {
	templateImage image.Image
	fontFace      font.Face
}

func NewStickerMaker() (*StickerMaker, error) {
	r := bytes.NewReader(templateBytes)
	templateImage, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}

	font, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil, err
	}

	fontFace, err := opentype.NewFace(font, &opentype.FaceOptions{
		Size: 48,
		DPI:  72,
	})
	if err != nil {
		return nil, err
	}

	result := &StickerMaker{templateImage, fontFace}
	return result, nil
}

func (m *StickerMaker) MakeSticker(text string) ([]byte, error) {
	canvas := gg.NewContextForImage(m.templateImage)

	canvas.SetColor(color.White)
	canvas.SetFontFace(m.fontFace)
	canvas.DrawStringWrapped(
		text,
		180, 80, // x, y
		0.0, 0.7, // anchorX, anchorY; mostly found by trial & error
		500, 1.0, // width, lineSpacing
		gg.AlignLeft,
	)

	buf := &bytes.Buffer{}
	err := webp.Encode(buf, canvas.Image(), nil)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
