package image

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/golang/freetype"
	"golang.org/x/image/math/fixed"
)

func Image(userName string, fileName string) {

	initial := getFirstTwoLetters(userName)

	width := 100
	height := 100

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	white := color.RGBA{0, 104, 71, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{white}, image.Point{}, draw.Src)

	fontBytes, err := ioutil.ReadFile("./font/Empire.ttf") // Font faylining yo'lini kiriting
	if err != nil {
		log.Println(err)
		return
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println(err)
		return
	}

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(font)
	c.SetFontSize(40)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.White)

	var pt fixed.Point26_6

	if len(initial) == 1{
		pt = freetype.Pt(33, 60) // Textning boshlang'ich nuqtasi
		_, err = c.DrawString(initial, pt)
		if err != nil {
			log.Println(err)
			return
		}
    }else if len(initial) == 2{
		pt = freetype.Pt(20, 60) // Textning boshlang'ich nuqtasi
		_, err = c.DrawString(initial, pt)
		if err != nil {
			log.Println(err)
			return
		}
	}
	uploadDir := "./avatar"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, os.ModePerm)
	}

	path := fmt.Sprintf("./avatar/%s.png", fileName)
	
	outFile, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	err = png.Encode(outFile, img)
	if err != nil {
		log.Fatal(err)
	}
}
func getFirstTwoLetters(name string) string {
		if name[0] == 's' && name[1] == 'h' {
			return strings.ToUpper(name[:2])
		} else if name[0] == 'c' && name[1] == 'h' {
			return strings.ToUpper(name[:2])
		}

	return strings.ToUpper(name[:1])
}