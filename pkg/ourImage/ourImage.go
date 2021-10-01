package ourimage

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

const weigh, heigh = 320, 200

type ourImage struct {
	// r [weigh][heigh]uint32
	// g [weigh][heigh]uint32
	// b [weigh][heigh]uint32
	r    [][]uint32 // Is an slice because otherwhise encoding is too slow
	g    [][]uint32
	b    [][]uint32
	Rect image.Rectangle
}

func NewImage(path string) (ourImage, error) {
	f, err := os.Open(path)
	instance := ourImage{}
	defer f.Close()
	if err != nil {
		return ourImage{}, err
	}
	rawImage := make([]byte, weigh*heigh)
	_, err = f.Read(rawImage)
	if err != nil {
		return ourImage{}, err
	}
	instance.r = make([][]uint32, weigh)
	for i := range instance.r {
		instance.r[i] = make([]uint32, heigh)
	}
	instance.g = make([][]uint32, weigh)
	for i := range instance.g {
		instance.g[i] = make([]uint32, heigh)
	}
	instance.b = make([][]uint32, weigh)
	for i := range instance.b {
		instance.b[i] = make([]uint32, heigh)
	}
	grayImage := image.NewGray(image.Rect(0, 0, weigh, heigh))
	grayImage.Pix = rawImage
	instance.Rect = grayImage.Bounds()
	for y := 0; y < grayImage.Bounds().Dy(); y++ {
		for x := 0; x < grayImage.Bounds().Dx(); x++ {
			instance.Set(x, y, grayImage.At(x, y))
		}
	}
	return instance, nil
}

func (self *ourImage) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(self.Rect)) {
		fmt.Printf("ourImage: Set x: %v, y: %v is outside bounds\n", x, y)
		return
	}
	self.r[x][y], self.g[x][y], self.b[x][y], _ = c.RGBA()
	self.r[x][y] >>= 8
	self.g[x][y] >>= 8
	self.b[x][y] >>= 8
}

// Help to implement image.Image
func (self ourImage) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(self.Rect)) {
		fmt.Printf("ourImage: At x: %v, y: %v is outside bounds\n", x, y)
		return color.Gray{}
	}
	return color.Gray{Y: uint8(self.r[x][y])}
}

// Help to implement image.Image
func (self ourImage) ColorModel() color.Model {
	return color.GrayModel
}

// Help to implement image.Image
func (self ourImage) Bounds() image.Rectangle {
	return self.Rect
}

func (self ourImage) Encode(path string) error {
	outfile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outfile.Close()
	png.Encode(outfile, self)
	return nil
}
