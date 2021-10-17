package ourimage

import (
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
  inputImg, _, err := image.Decode(f)
  if err == image.ErrFormat {
    pixels := make([]byte, 320*200) // TODO dynamic size?
    f.Seek(0, 0)
    _, err = f.Read(pixels)
    if err != nil {
      return nil, err
    }
    grayImg := image.NewGray(image.Rect(0, 0, 320, 200))
    grayImg.Pix = pixels
    inputImg = grayImg
  } else if err != nil {
    return nil, err
  }
  b := inputImg.Bounds()
  img := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
  draw.Draw(img, img.Bounds(), inputImg, b.Min, draw.Src)
  return img, nil
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
