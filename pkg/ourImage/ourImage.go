package ourimage

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"sync"
)

const weight, height = 320, 200

type ourImage struct {
	// Is a slice because otherwhise encoding would be too slow, this way we avoid making heavy copies
	r    [][]uint32
	g    [][]uint32
	b    [][]uint32
	rect image.Rectangle
}

func NewImage(path string) (ourImage, error) {
	f, err := os.Open(path) // Open file
	if err != nil {
		return ourImage{}, err
	}
	defer f.Close()

	instance := ourImage{} // Initialize slices
	initialize := func() [][]uint32 {
		colorLayer := make([][]uint32, weight)
		for i := range colorLayer {
			colorLayer[i] = make([]uint32, height)
		}
		return colorLayer
	}
	instance.r = initialize()
	instance.g = initialize()
	instance.b = initialize()

	grayImage := image.NewGray(image.Rect(0, 0, weight, height))
	rawImage := make([]byte, weight*height) // Read raw image
	_, err = f.Read(rawImage)
	if err != nil {
		return ourImage{}, err
	}
	grayImage.Pix = rawImage
	instance.rect = grayImage.Bounds()
	for y := 0; y < grayImage.Bounds().Dy(); y++ {
		for x := 0; x < grayImage.Bounds().Dx(); x++ {
			instance.Set(x, y, grayImage.At(x, y))
		}
	}
	return instance, nil
}

func (self *ourImage) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(self.rect)) {
		fmt.Printf("ourImage: Set x: %v, y: %v is outside bounds\n", x, y)
		return
	}
	r, g, b, _ := c.RGBA()
	self.r[x][y] = r >> 8
	self.g[x][y] = g >> 8
	self.b[x][y] = b >> 8
}

// Help to implement image.Image
func (self ourImage) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(self.rect)) {
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
	return self.rect
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

// Functions
// TODO Better performance. Don't make a copy if you are going to overwrite it anyway
func (self ourImage) Negative() ourImage {
  negativeImage := self
  negativeF := func(layer [][]uint32) {
    for x := range layer {
      for y := range layer[x] {
        layer[x][y] = 255 - layer[x][y]
      }
    }
  }
  var wg sync.WaitGroup
  wg.Add(3)
  go func() {
    negativeF(negativeImage.r)
    wg.Done()
  }()
  go func() {
    negativeF(negativeImage.g)
    wg.Done()
  }()
  go func() {
    negativeF(negativeImage.b)
    wg.Done()
  }()
  wg.Wait()
  return negativeImage
}
