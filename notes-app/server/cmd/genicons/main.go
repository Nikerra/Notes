package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

func main() {
	createIcon(192, "web/icon-192.png")
	createIcon(512, "web/icon-512.png")
}

func createIcon(size int, filename string) {
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			ratio := float64(x+y) / float64(size*2)
			r := uint8(102 + float64(118-102)*ratio)
			g := uint8(126 + float64(75-126)*ratio)
			b := uint8(234 + float64(162-234)*ratio)
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	pad := size / 8
	lineHeight := size / 15
	linePad := size / 6

	white := color.RGBA{255, 255, 255, 255}
	for y := pad; y < size-pad; y++ {
		for x := pad; x < size-pad; x++ {
			line1Y := pad + linePad
			line2Y := line1Y + linePad
			line3Y := line2Y + linePad
			line4Y := line3Y + linePad

			radius := lineHeight / 2

			if y >= line1Y-radius && y <= line1Y+radius && x >= pad && x <= size-pad {
				img.Set(x, y, white)
			}
			if y >= line2Y-radius && y <= line2Y+radius && x >= pad && x <= size-pad*3 {
				img.Set(x, y, white)
			}
			if y >= line3Y-radius && y <= line3Y+radius && x >= pad && x <= size-pad*2 {
				img.Set(x, y, white)
			}
			if y >= line4Y-radius && y <= line4Y+radius && x >= pad && x <= size-pad*4 {
				img.Set(x, y, white)
			}
		}
	}

	f, _ := os.Create(filename)
	png.Encode(f, img)
	f.Close()
}
