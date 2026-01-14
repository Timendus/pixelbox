package protocol

import (
	"fmt"
	"image"
	"math"
)

type Color [3]byte

func (c1 *Color) equals(c2 Color) bool {
	return c1[0] == c2[0] &&
		c1[1] == c2[1] &&
		c1[2] == c2[2]
}

func convertImage(image *image.RGBA) ([]byte, []byte, error) {
	if image.Bounds().Size().X != 16 || image.Bounds().Size().Y != 16 {
		return nil, nil, fmt.Errorf("image needs to be 16x16, got: %dx%d", image.Bounds().Size().X, image.Bounds().Size().Y)
	}

	// Collect all the unique colors to form a palette and create the indexed
	// image data in one go
	width := image.Bounds().Size().X
	height := image.Bounds().Size().Y
	paletteData := make([]Color, 0)
	imageData := make([]int, 0)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			index := (y*width + x) * 4
			color := Color{
				image.Pix[index+0],
				image.Pix[index+1],
				image.Pix[index+2],
			}
			new := true
			for i, c := range paletteData {
				if c.equals(color) {
					new = false
					imageData = append(imageData, i)
					break
				}
			}
			if new {
				paletteData = append(paletteData, color)
				imageData = append(imageData, len(paletteData)-1)
			}
		}
	}

	// Convert the palette data to the required byte array, which is just a 24
	// bits [ R, G, B ] per color.
	paletteBytes := make([]byte, len(paletteData)*3)
	for i := range paletteData {
		paletteBytes[i*3+0] = paletteData[i][0]
		paletteBytes[i*3+1] = paletteData[i][1]
		paletteBytes[i*3+2] = paletteData[i][2]
	}

	// Calculate bits needed per pixel to index the whole palette
	bpp := int(math.Ceil(math.Log2(float64(len(paletteData)))))
	// Calculate total number of bytes needed for the image
	totalBytes := int(math.Ceil(float64(bpp * width * height / 8)))

	// Convert the indexed image data to the required bitstream. Which is a bit
	// weird in that the top left pixel starts at the least significant `bpp`
	// bits of the first byte instead of the most significant side. This repeats
	// for every byte, in essense reading each byte "backwards", but the bytes
	// in order.
	imageBytes := make([]byte, totalBytes)
	offset := 0
	index := 0
	for _, colorIndex := range imageData {
		imageBytes[index] = imageBytes[index] | byte(colorIndex<<offset)
		if index+1 < len(imageBytes) {
			imageBytes[index+1] = imageBytes[index+1] | byte(colorIndex>>(8-offset))
		}
		offset += bpp
		if offset > 8 {
			offset -= 8
			index++
		}
	}

	return paletteBytes, imageBytes, nil
}
