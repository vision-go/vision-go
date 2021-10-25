package ourimage

import (
	"fmt"
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

func deepCopy(originalImg image.Image) image.Image {
	b := originalImg.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(NewImage, NewImage.Bounds(), originalImg, b.Min, draw.Src)
  return NewImage
}

type SubImager interface {
    SubImage(r image.Rectangle) image.Image
}

type OurImage struct {
	widget.BaseWidget
	canvasImage *canvas.Image
	format      string
	statusBar   *widget.Label
  mainWindow fyne.Window
  // rectangle *canvas.Rectangle
  rectangle image.Rectangle
  ROIcallback func(OurImage)
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
	} else {
	    fmt.Println("Wrong")
	  }
}

// MouseOut is a hook that is called if the mouse pointer leaves the element.
func (ourimage *OurImage) MouseOut() {
	if ourimage.statusBar != nil {
		ourimage.statusBar.SetText("")
	}
  // ourimage.rectangle.Resize(fyne.NewSize(0, 0))
  ourimage.rectangle.Min = image.Pt(0, 0)
  ourimage.rectangle.Max = image.Pt(0, 0)
}

// // desktop.Mouseable
func (ourimage *OurImage) MouseDown(mouseEvent *desktop.MouseEvent) {
  // ourimage.rectangle.Resize(fyne.NewSize(0, 0))
  // ourimage.rectangle.Move(mouseEvent.AbsolutePosition.Subtract(fyne.NewPos(5, 74)))
  ourimage.rectangle.Min = image.Pt(int(mouseEvent.Position.X), int(mouseEvent.Position.Y))
}

func (ourimage *OurImage) MouseUp(mouseEvent *desktop.MouseEvent) {
  ourimage.rectangle.Max = image.Pt(int(mouseEvent.Position.X), int(mouseEvent.Position.Y))
  newImage := ourimage.canvasImage.Image
  fmt.Println("Size before crop: ", newImage.Bounds())
  cropped := newImage.(SubImager).SubImage(ourimage.rectangle)
  fmt.Println("Size after crop: ", cropped.Bounds())
  ourimage.ROIcallback(newOurImageFromImage(*ourimage, cropped))
}

// func (ourimage *OurImage) Tapped(mouseEvent *fyne.PointEvent) {
//   ourimage.rectangle.Resize(fyne.NewSize(0, 0))
//   ourimage.rectangle.Refresh()
// }

func (self *OurImage) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(self.canvasImage)
}

func NewImage(path string, statusBar *widget.Label, w fyne.Window, ROIcallback func(OurImage)) (OurImage, error) {
	var img OurImage
	img.statusBar = statusBar
  img.mainWindow = w
  img.ROIcallback = ROIcallback
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

func newOurImageFromImage(ourImage OurImage, newImage image.Image) OurImage {
  newSize := fyne.NewSize(float32(newImage.Bounds().Dx()), float32(newImage.Bounds().Dy()))
  fmt.Println("It should have: ", newSize)
  ourImage.canvasImage = canvas.NewImageFromImage(newImage)
  ourImage.canvasImage.Resize(newSize)
  ourImage.canvasImage.FillMode = canvas.ImageFillOriginal
  fmt.Println("It has: ", ourImage.canvasImage.Size())
  return ourImage
}

func (img OurImage) Format() string {
	return img.format
}

func (img OurImage) Dimensions() image.Point {
	return img.canvasImage.Image.Bounds().Size()
}

// Functions
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
  return newOurImageFromImage(*originalImg, NewImage)
}

func (originalImg *OurImage) Monochrome() OurImage { // TODO it makes a copy
	b := originalImg.canvasImage.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(NewImage, NewImage.Bounds(), originalImg.canvasImage.Image, b.Min, draw.Src)

	for y := 0; y < originalImg.canvasImage.Image.Bounds().Dy(); y++ {
		for x := 0; x < originalImg.canvasImage.Image.Bounds().Dx(); x++ {
			oldColour := originalImg.canvasImage.Image.At(x, y)
			r, g, b, _ := oldColour.RGBA()
			NewImage.Set(x, y, color.Gray{Y: uint8(0.222*float32(r>>8) + 0.707*float32(g>>8) + 0.071*float32(b>>8))})
		}
	}
  return newOurImageFromImage(*originalImg, NewImage)
}
