package ourimage

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"image/draw"
  "image/color"
)

func NewImage(path string) (draw.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
  img, format, err := image.Decode(f)
  if err != nil {
    return nil, err
  }
  dimg, ok := img.(draw.Image)
  if !ok {
    return nil, fmt.Errorf("%v, is not a valid format", format)
  }
  return dimg, nil
}

func Negative(img draw.Image) draw.Image {
  NewImage := img
  for y := 0; y < img.Bounds().Dy(); y++ {
    for x := 0; x < img.Bounds().Dx(); x++ {
      col := img.At(x, y)
      r, g, b, a := col.RGBA()
      newCol := color.RGBA{uint8(255-r), uint8(255-g), uint8(255-b), uint8(a)}
      NewImage.Set(x, y, newCol)
    }
  }
  return NewImage
}
