package ourimage

import (
	"fmt"
	"image"
	"image/color"
  "image/png"
	"os"
	"sync"
)

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

  img, _, err := image.Decode(f)
  if err == image.ErrFormat {
    
  } else if err != nil {
    return ourImage{}, err
  }
  // fmt.Println("format", format)
  // fmt.Printf("tipo de imagen: %T\n", img)

	instance := ourImage{} 
  instance.rect = img.Bounds()
	initialize := func() [][]uint32 { // Initialize slices
		colorLayer := make([][]uint32, img.Bounds().Dx())
		for i := range colorLayer {
			colorLayer[i] = make([]uint32, img.Bounds().Dy())
		}
		return colorLayer
	}
	instance.r = initialize()
	instance.g = initialize()
	instance.b = initialize()

  for y := 0; y < img.Bounds().Dy(); y++ {
    for x := 0; x < img.Bounds().Dx(); x++ {
      instance.Set(x, y, img.At(x, y))
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
  self.r[x][y] = r
  self.g[x][y] = g
  self.b[x][y] = b
}

// Help to implement image.Image
func (self ourImage) At(x, y int) color.Color {
  if !(image.Point{x, y}.In(self.rect)) {
    fmt.Printf("ourImage: At x: %v, y: %v is outside bounds\n", x, y)
    return color.NRGBA{}
  }
  return color.NRGBA{R: uint8(self.r[x][y]), G: uint8(self.g[x][y]), B: uint8(self.b[x][y]), A: 255}
}

// Help to implement image.Image
func (self ourImage) ColorModel() color.Model {
  return color.NRGBAModel
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
  // jpeg.Encode(outfile, self, nil)
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
