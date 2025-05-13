package utils

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"time"

	"github.com/golang/freetype"
	"golang.org/x/image/font"
)

func GenerateTimestampImage(timestamp time.Time, outputPath string) error {
	text := timestamp.Format("02/01/2006 15:04:05")

	// Font loading
	fontBytes, err := os.ReadFile("./fonts/dejavu-sans-bold.ttf")
	if err != nil {
		return err
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return err
	}

	const fontSize = 30
	fgColor := color.RGBA{R: 198, G: 23, B: 187, A: 255} // #c617bb

	// Estimate text bounds with a more generous width
	width := len(text) * (fontSize / 1.5)
	height := int(fontSize * 1.5)

	// Transparent background
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(rgba, rgba.Bounds(), &image.Uniform{color.Transparent}, image.Point{}, draw.Src)

	// Draw the text
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(f)
	c.SetFontSize(fontSize)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(&image.Uniform{fgColor})
	c.SetHinting(font.HintingFull)

	// Add left padding to center the text better
	leftPadding := fontSize
	pt := freetype.Pt(leftPadding, int(c.PointToFixed(fontSize)>>6))
	if _, err := c.DrawString(text, pt); err != nil {
		return err
	}

	// Save PNG
	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	if err := png.Encode(outFile, rgba); err != nil {
		return err
	}
	return nil
}
