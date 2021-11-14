package ourimage

import (
  "image"
  "fyne.io/fyne/v2/canvas"
)

func (img *OurImage) Name() string {
	return img.name
}

func (img *OurImage) Format() string {
	return img.format
}

func (img *OurImage) Dimensions() image.Point {
	return img.canvasImage.Image.Bounds().Size()
}

func (img *OurImage) Brightness() float64 {
	return img.brightness
}

func (img *OurImage) Contrast() float64 {
	return img.contrast
}

func (img *OurImage) EntropyAndNumberOfColors() (int, int) {
	return img.entropy, img.numberOfColors
}

func (img *OurImage) CanvasImage() *canvas.Image {
  return img.canvasImage
}

func (img *OurImage) MinAndMaxColor() (int, int) {
  return img.minColor, img.maxColor
}

// [4,8] -> 4, 5, 6, 7, 8 (5)
func (img *OurImage) Range() int {
  return img.maxColor - img.minColor + 1
}
