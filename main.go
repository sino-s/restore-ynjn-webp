package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path/filepath"

	_ "golang.org/x/image/webp"
)

func main() {
	const (
		rows = 4
		cols = 4
	)

	if len(os.Args) < 2 {
		log.Fatalf("usage: %s <input_file>", filepath.Base(os.Args[0]))
	}
	inputFileName := os.Args[1]

	file, err := os.Open(inputFileName)
	if err != nil {
		log.Fatalf("Cannot open input file.: %v", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatalf("Image decoding failed.: %v", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	partWidth := width / cols
	partHeight := height / rows

	if partWidth == 0 || partHeight == 0 {
		log.Fatalf("Image width or height is less than the number of segments (%dx%d).", cols, rows)
	}

	parts := make([][]image.Image, rows)
	for r := 0; r < rows; r++ {
		parts[r] = make([]image.Image, cols)
		for c := 0; c < cols; c++ {
			x0 := c * partWidth
			y0 := r * partHeight
			rect := image.Rect(x0, y0, x0+partWidth, y0+partHeight)
			subImgInterface, ok := img.(interface {
				SubImage(r image.Rectangle) image.Image
			})
			if !ok {
				log.Fatalf("Image does not support the SubImage method.")
			}
			subImg := subImgInterface.SubImage(rect)
			parts[r][c] = subImg
		}
	}

	newImg := image.NewRGBA(bounds)
	for newR := 0; newR < rows; newR++ {
		for newC := 0; newC < cols; newC++ {
			oldR := newC
			oldC := newR
			if oldR >= rows || oldC >= cols {
				log.Fatalf("Index is out of range.: oldR=%d, oldC=%d", oldR, oldC)
			}
			subImg := parts[oldR][oldC]

			x0 := newC * partWidth
			y0 := newR * partHeight
			destRect := image.Rect(x0, y0, x0+partWidth, y0+partHeight)

			draw.Draw(newImg, destRect, subImg, subImg.Bounds().Min, draw.Over)
		}
	}

	outputFileName := changeExtensionToPNG(inputFileName)

	outFile, err := os.Create(outputFileName)
	if err != nil {
		log.Fatalf("Unable to create output file.: %v", err)
	}
	defer outFile.Close()

	if err = png.Encode(outFile, newImg); err != nil {
		log.Fatalf("Failed to save in PNG format.: %v", err)
	}

	fmt.Printf("Image saved in %s\n", outputFileName)
}

func changeExtensionToPNG(filename string) string {
	return filename[:len(filename)-len(filepath.Ext(filename))] + ".png"
}
