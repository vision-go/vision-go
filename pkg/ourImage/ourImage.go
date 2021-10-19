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
)

type OurImage struct {
	widget.BaseWidget
	Image     *canvas.Image
	format    string
	statusBar *widget.Label
}

func (self *OurImage) MouseIn(mouse *desktop.MouseEvent) {
	if self.statusBar != nil {
		r, g, b, a := self.Image.Image.At(int(mouse.Position.X), int(mouse.Position.Y)).RGBA()
		self.statusBar.SetText("R: " + strconv.Itoa(int(r>>8)) + " || G: " + strconv.Itoa(int(g>>8)) + " || B: " + strconv.Itoa(int(b>>8)) + " || A: " + strconv.Itoa(int(a>>8)))
	}
}

// MouseMoved is a hook that is called if the mouse pointer moved over the element.
func (self *OurImage) MouseMoved(mouse *desktop.MouseEvent) {
	if self.statusBar != nil {
		r, g, b, a := self.Image.Image.At(int(mouse.Position.X), int(mouse.Position.Y)).RGBA()
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
	return widget.NewSimpleRenderer(self.Image)
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
	img.Image = canvas.NewImageFromImage(inputImg)
	img.Image.FillMode = canvas.ImageFillOriginal
	return img, nil
}

func Negative(originalImg OurImage) OurImage { // TODO it makes a copy
	b := originalImg.Image.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(NewImage, NewImage.Bounds(), originalImg.Image.Image, b.Min, draw.Src)

	for y := 0; y < originalImg.Image.Image.Bounds().Dy(); y++ {
		for x := 0; x < originalImg.Image.Image.Bounds().Dx(); x++ {
			col := originalImg.Image.Image.At(x, y)
			r, g, b, a := col.RGBA()
			newCol := color.RGBA{uint8(255 - r), uint8(255 - g), uint8(255 - b), uint8(a)}
			NewImage.Set(x, y, newCol)
		}
	}
	newOurImage := originalImg
	newOurImage.Image = canvas.NewImageFromImage(NewImage)
	return newOurImage
}
