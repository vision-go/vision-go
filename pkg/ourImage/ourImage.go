package ourimage

import (
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	lookUpTable "github.com/vision-go/vision-go/pkg/look-up-table"
)

type OurImage struct {
	widget.BaseWidget
	canvasImage     *canvas.Image
	format    string
	statusBar *widget.Label
}

func (self *OurImage) MouseIn(mouse *desktop.MouseEvent) {
	if self.statusBar != nil {
		r, g, b, a := self.canvasImage.Image.At(int(mouse.Position.X), int(mouse.Position.Y)).RGBA()
		self.statusBar.SetText("R: " + strconv.Itoa(int(r>>8)) + " || G: " + strconv.Itoa(int(g>>8)) + " || B: " + strconv.Itoa(int(b>>8)) + " || A: " + strconv.Itoa(int(a>>8)))
	}
}

// MouseMoved is a hook that is called if the mouse pointer moved over the element.
func (self *OurImage) MouseMoved(mouse *desktop.MouseEvent) {
	if self.statusBar != nil {
		r, g, b, a := self.canvasImage.Image.At(int(mouse.Position.X), int(mouse.Position.Y)).RGBA()
		self.statusBar.SetText("R: " + strconv.Itoa(int(r>>8)) + " || G: " + strconv.Itoa(int(g>>8)) + " || B: " + strconv.Itoa(int(b>>8)) + " || A: " + strconv.Itoa(int(a>>8)))
	}
}

// MouseOut is a hook that is called if the mouse pointer leaves the element.
func (self *OurImage) MouseOut() {
	if self.statusBar != nil {
		self.statusBar.SetText("")
	}
}

func (self *OurImage) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(self.canvasImage)
}

func NewImage(path string, statusBar *widget.Label) (OurImage, error) {
	var img OurImage
	img.statusBar = statusBar
	img.ExtendBaseWidget(&img)
	f, err := os.Open(path)
	if err != nil {
		return img, err
	}
	defer f.Close()
	inputImg, format, err := image.Decode(f)
	img.format = format
	if err == image.ErrFormat {
		img.format = "tfe(no format)"
		pixels := make([]byte, 320*200) // TODO dynamic size?
		f.Seek(0, 0)
		_, err = f.Read(pixels)
		if err != nil {
			return img, err
		}
		grayImg := image.NewGray(image.Rect(0, 0, 320, 200))
		grayImg.Pix = pixels
		inputImg = grayImg
	} else if err != nil {
		return img, err
	}
	img.canvasImage = canvas.NewImageFromImage(inputImg)
	img.canvasImage.FillMode = canvas.ImageFillOriginal
	return img, nil
}

func (originalImg *OurImage) Negative() OurImage { // TODO it makes a copy
	b := originalImg.canvasImage.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(NewImage, NewImage.Bounds(), originalImg.canvasImage.Image, b.Min, draw.Src)

	for y := 0; y < originalImg.canvasImage.Image.Bounds().Dy(); y++ {
		for x := 0; x < originalImg.canvasImage.Image.Bounds().Dx(); x++ {
			oldColour := originalImg.canvasImage.Image.At(x, y)
      NewImage.Set(x, y, lookUpTable.RGBA(oldColour, lookUpTable.Negative))
		}
	}
	newOurImage := *originalImg
	newOurImage.canvasImage = canvas.NewImageFromImage(NewImage)
	return newOurImage
}

func (originalImg *OurImage) Monochrome() OurImage { // TODO it makes a copy
	b := originalImg.canvasImage.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(NewImage, NewImage.Bounds(), originalImg.canvasImage.Image, b.Min, draw.Src)

	for y := 0; y < originalImg.canvasImage.Image.Bounds().Dy(); y++ {
		for x := 0; x < originalImg.canvasImage.Image.Bounds().Dx(); x++ {
			oldColour := originalImg.canvasImage.Image.At(x, y)
      r, g, b, _ := oldColour.RGBA()
      NewImage.Set(x, y, color.Gray{Y: uint8(0.222 * float32(r>>8) + 0.707 * float32(g>>8) + 0.071 * float32(b>>8))})
		}
	}
	newOurImage := *originalImg
	newOurImage.canvasImage = canvas.NewImageFromImage(NewImage)
	return newOurImage
}
