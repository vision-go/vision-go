package ourimage

import (
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	lookUpTable "github.com/vision-go/vision-go/pkg/look-up-table"
)

type OurImage struct {
	widget.BaseWidget
  name string
	canvasImage *canvas.Image
	format      string
	statusBar   *widget.Label
	mainWindow  fyne.Window
	rectangle   image.Rectangle
	ROIcallback func(*OurImage)
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
func (ourimage *OurImage) MouseOut() {
	if ourimage.statusBar != nil {
		ourimage.statusBar.SetText("")
	}
}

// desktop.Mouseable
func (ourimage *OurImage) MouseDown(mouseEvent *desktop.MouseEvent) {
	ourimage.rectangle.Min = image.Pt(int(math.Round(float64(mouseEvent.Position.X))), int(math.Round(float64(mouseEvent.Position.Y))))
}

func (ourimage *OurImage) MouseUp(mouseEvent *desktop.MouseEvent) {
	ourimage.rectangle.Max = image.Pt(int(math.Round(float64(mouseEvent.Position.X))), int(math.Round(float64(mouseEvent.Position.Y))))
  if ourimage.rectangle.Dx() > 10 && ourimage.rectangle.Dy() > 10 {
    ourimage.ROIcallback(ourimage.ROI(ourimage.rectangle))
  }
}

func (self *OurImage) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(self.canvasImage)
}

func NewFromPath(path, name string, statusBar *widget.Label, w fyne.Window, ROIcallback func(*OurImage)) (*OurImage, error) {
	img := &OurImage{}
  img.name = name
	img.statusBar = statusBar
	img.mainWindow = w
	img.ROIcallback = ROIcallback
	img.ExtendBaseWidget(img)
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

func (img OurImage) addOperationToName(actionForName string) string {
  actionForName = "(" + actionForName + ")"
  pointIndex := strings.LastIndex(img.name, ".")
  if pointIndex == -1 {
    return img.name + actionForName
  }
  return img.name[:pointIndex] + actionForName + img.name[pointIndex:]
}

func newFromImage(ourImage *OurImage, newImage image.Image, actionForName string) *OurImage {
	img := &OurImage{}
  img.name = ourImage.addOperationToName(actionForName)
	img.statusBar = ourImage.statusBar
	img.mainWindow = ourImage.mainWindow
	img.ROIcallback = ourImage.ROIcallback
	img.ExtendBaseWidget(img)
	img.canvasImage = canvas.NewImageFromImage(newImage)
	img.canvasImage.FillMode = canvas.ImageFillOriginal
	return img
}

func (img OurImage) Name() string {
  return img.name
}

func (img OurImage) Format() string {
	return img.format
}

func (img OurImage) Dimensions() image.Point {
	return img.canvasImage.Image.Bounds().Size()
}

// Functions
func (originalImg *OurImage) Negative() *OurImage { // TODO it makes a copy
	b := originalImg.canvasImage.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))

	for y := 0; y < originalImg.canvasImage.Image.Bounds().Dy(); y++ {
		for x := 0; x < originalImg.canvasImage.Image.Bounds().Dx(); x++ {
			oldColour := originalImg.canvasImage.Image.At(x, y)
			NewImage.Set(x, y, lookUpTable.RGBA(oldColour, lookUpTable.Negative))
		}
	}
  return newFromImage(originalImg, NewImage, "Negative")
}

func (originalImg *OurImage) Monochrome() *OurImage { // TODO it makes a copy
	b := originalImg.canvasImage.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))

	for y := 0; y < originalImg.canvasImage.Image.Bounds().Dy(); y++ {
		for x := 0; x < originalImg.canvasImage.Image.Bounds().Dx(); x++ {
			oldColour := originalImg.canvasImage.Image.At(x, y)
			r, g, b, _ := oldColour.RGBA()
			NewImage.Set(x, y, color.Gray{Y: uint8(0.222*float32(r>>8) + 0.707*float32(g>>8) + 0.071*float32(b>>8))}) // PAL
		}
	}
  return newFromImage(originalImg, NewImage, "Monochrome")
}

func (originalImg *OurImage) ROI(rect image.Rectangle) *OurImage {
	b := rect.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := 0; y < rect.Dy(); y++ {
		for x := 0; x < rect.Dx(); x++ {
			NewImage.Set(x, y, originalImg.canvasImage.Image.At(x+rect.Min.X, y+rect.Min.Y))
		}
	}
  return newFromImage(originalImg, NewImage, "ROI")
}
