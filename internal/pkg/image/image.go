package image

import (
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"

	"strings"

	"github.com/golang/freetype"
	"golang.org/x/image/math/fixed"
)

func Image(userName string) (image.Image, error) {
	initial := getFirstTwoLetters(userName)

	width := 100
	height := 100

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	white := color.RGBA{0, 104, 71, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{white}, image.Point{}, draw.Src)

	fontBytes, err := ioutil.ReadFile("./font/Empire.ttf") // Font faylining yo'lini kiriting
	if err != nil {
		return nil, err
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, err
	}

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(font)
	c.SetFontSize(40)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.White)

	var pt fixed.Point26_6

	if len(initial) == 1 {
		pt = freetype.Pt(33, 60) // Textning boshlang'ich nuqtasi
	} else if len(initial) == 2 {
		pt = freetype.Pt(20, 60) // Textning boshlang'ich nuqtasi
	}
	_, err = c.DrawString(initial, pt)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func getFirstTwoLetters(name string) string {
	if name[0] == 's' && name[1] == 'h' {
		return strings.ToUpper(name[:2])
	} else if name[0] == 'c' && name[1] == 'h' {
		return strings.ToUpper(name[:2])
	}

	return strings.ToUpper(name[:1])
}
