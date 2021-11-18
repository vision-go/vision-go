package ourimage

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	histogram "github.com/vision-go/vision-go/pkg/histogram"
	lookUpTable "github.com/vision-go/vision-go/pkg/look-up-table"
)

type OurImage struct {
	widget.BaseWidget
	name        string
	canvasImage *canvas.Image
	format      string
	statusBar   *widget.Label
	mainWindow  fyne.Window
	rectangle   image.Rectangle
	ROIcallback func(*OurImage)

	HistogramR histogram.Histogram
	HistogramG histogram.Histogram
	HistogramB histogram.Histogram
	Histogram  histogram.Histogram

	HistogramAccumulativeR histogram.Histogram
	HistogramAccumulativeG histogram.Histogram
	HistogramAccumulativeB histogram.Histogram
	HistogramAccumulative  histogram.Histogram

	HistogramNormalizedR histogram.HistogramNormalized
	HistogramNormalizedG histogram.HistogramNormalized
	HistogramNormalizedB histogram.HistogramNormalized
	HistogramNormalized  histogram.HistogramNormalized
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

	makeHistogram(img)

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
	makeHistogram(img)
	return img
}

func (img OurImage) Save(file *os.File, format string) error {
	if format == "png" {
		return png.Encode(file, img.canvasImage.Image)
	} else if format == "jpeg" || format == "jpg" {
		return jpeg.Encode(file, img.canvasImage.Image, nil)
	}
	return fmt.Errorf("Incorrrect format")
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

func makeHistogram(image *OurImage) {
	for i := 0; i < 256; i++ {
		image.HistogramR.Values[i] = 0
		image.HistogramG.Values[i] = 0
		image.HistogramB.Values[i] = 0
		image.Histogram.Values[i] = 0

		image.HistogramAccumulativeR.Values[i] = 0
		image.HistogramAccumulativeG.Values[i] = 0
		image.HistogramAccumulativeB.Values[i] = 0
		image.HistogramAccumulative.Values[i] = 0

		image.HistogramNormalized.Values[i] = 0.0
		image.HistogramNormalizedR.Values[i] = 0.0
		image.HistogramNormalizedG.Values[i] = 0.0
		image.HistogramNormalizedB.Values[i] = 0.0
	}
	for i := 0; i < image.canvasImage.Image.Bounds().Dx(); i++ {
		for j := 0; j < image.canvasImage.Image.Bounds().Dy(); j++ {
			r, g, b, _ := image.canvasImage.Image.At(i, j).RGBA()
			
				r, g, b = r>>8, g>>8, b>>8
				image.HistogramR.Values[r] = image.HistogramR.At(int(r)) + 1
				image.HistogramG.Values[g] = image.HistogramG.At(int(g)) + 1
				image.HistogramB.Values[b] = image.HistogramG.At(int(b)) + 1

				grey := 0.222*float64(r) + 0.707*float64(g) + 0.071*float64(b) // PAL
				image.Histogram.Values[int(math.Round(grey))] = image.Histogram.At(int(math.Round(grey))) + 1
			
		}
	}
	for index := range image.Histogram.Values {
		for i := 0; i < index; i++ {
			image.HistogramAccumulativeR.Values[index] += image.HistogramR.At(i)
			image.HistogramAccumulativeG.Values[index] += image.HistogramG.At(i)
			image.HistogramAccumulativeB.Values[index] += image.HistogramB.At(i)
			image.HistogramAccumulative.Values[index] += image.Histogram.At(i)
		}
	}
	for i := 0; i < 256; i++ {
		image.HistogramNormalized.Values[i] = float64(image.Histogram.Values[i]) / float64(image.canvasImage.Image.Bounds().Dx()*image.canvasImage.Image.Bounds().Dy())
		image.HistogramNormalizedR.Values[i] = float64(image.HistogramR.Values[i]) / float64(image.canvasImage.Image.Bounds().Dx()*image.canvasImage.Image.Bounds().Dy())
		image.HistogramNormalizedG.Values[i] = float64(image.HistogramG.Values[i]) / float64(image.canvasImage.Image.Bounds().Dx()*image.canvasImage.Image.Bounds().Dy())
		image.HistogramNormalizedB.Values[i] = float64(image.HistogramB.Values[i]) / float64(image.canvasImage.Image.Bounds().Dx()*image.canvasImage.Image.Bounds().Dy())
	}
}

func (originalImg *OurImage) Equalization() *OurImage { // TODO it makes a copy
	b := originalImg.canvasImage.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	size := b.Dx() * b.Dy()
	var lookUpTableArrayR [256]int
	var lookUpTableArrayG [256]int
	var lookUpTableArrayB [256]int

	
	for i := 0; i < 256; i++{
		lookUpTableArrayR[i] = Max(0,int(math.Round((float64(originalImg.HistogramAccumulativeR.At(i)*256)/float64(size))-1)))
		lookUpTableArrayG[i] = Max(0,int(math.Round((float64(originalImg.HistogramAccumulativeG.At(i)*256)/float64(size))-1)))
		lookUpTableArrayB[i] = Max(0,int(math.Round((float64(originalImg.HistogramAccumulativeB.At(i)*256)/float64(size))-1)))
	}
	for y := 0; y < originalImg.canvasImage.Image.Bounds().Dy(); y++ {
		for x := 0; x < originalImg.canvasImage.Image.Bounds().Dx(); x++ {
			oldColour := originalImg.canvasImage.Image.At(x, y)
			r,g,b,a := oldColour.RGBA()
			r,g,b = r >> 8, g >> 8, b >> 8
			
			newColor := color.RGBA{
				R: uint8(lookUpTableArrayR[r]),
				G: uint8(lookUpTableArrayG[g]),
				B: uint8(lookUpTableArrayB[b]),
				A: uint8(a),
			}
			NewImage.Set(x, y, newColor)
		}
	}
	return  newFromImage(originalImg, NewImage, "Ecualization")
}

func (originalImg *OurImage) HistogramIgualation(imageIn *OurImage) *OurImage { // TODO it makes a copy
	b := originalImg.canvasImage.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	size := b.Dx() * b.Dy()
	var lookUpTableArrayR [256]int
	var lookUpTableArrayG [256]int
	var lookUpTableArrayB [256]int
	sizeF := float64(size)

	M := 256
	PoR := originalImg.HistogramAccumulativeR.Values
	PoG := originalImg.HistogramAccumulativeG.Values
	PoB := originalImg.HistogramAccumulativeB.Values
	PrR := imageIn.HistogramAccumulativeR.Values
	PrG := imageIn.HistogramAccumulativeG.Values
	PrB := imageIn.HistogramAccumulativeB.Values
	
	for a, _ := range lookUpTableArrayR{
		
		for j := M - 1 ; j >= 0; j--{
			lookUpTableArrayR[a] = j
			if (float64(PoR[a])/sizeF) > (float64(PrR[j])/sizeF) {
				break
			}
		}

		for j := M - 1 ; j >= 0 ; j--{
			lookUpTableArrayG[a] = j
			if (float64(PoG[a])/sizeF) > (float64(PrG[j])/sizeF) {
				break
			}
		}

		for j := M - 1 ; j >= 0; j--{
			lookUpTableArrayB[a] = j
			if (float64(PoB[a])/sizeF) > (float64(PrB[j])/sizeF) {
				break
			}
		}
	}
	for y := 0; y < originalImg.canvasImage.Image.Bounds().Dy(); y++ {
		for x := 0; x < originalImg.canvasImage.Image.Bounds().Dx(); x++ {
			oldColour := originalImg.canvasImage.Image.At(x, y)
			r,g,b,a := oldColour.RGBA()
			r,g,b = r >> 8, g >> 8, b >> 8
			
			newColor := color.RGBA{
				R: uint8(lookUpTableArrayR[r]),
				G: uint8(lookUpTableArrayR[g]),
				B: uint8(lookUpTableArrayR[b]),
				A: uint8(a),
			}
			NewImage.Set(x, y, newColor)
		}
	}
	return  newFromImage(originalImg, NewImage, "Histogram Igualated")
}

func (originalImg *OurImage) ImageDiference(imageIn *OurImage) *OurImage { // TODO it makes a copy
	b := originalImg.canvasImage.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))

	for y := 0; y < originalImg.canvasImage.Image.Bounds().Dy(); y++ {
		for x := 0; x < originalImg.canvasImage.Image.Bounds().Dx(); x++ {
			oldColour := originalImg.canvasImage.Image.At(x, y)
			r,g,b,a := oldColour.RGBA()
			r,g,b = r >> 8, g >> 8, b >> 8

			newColour := imageIn.canvasImage.Image.At(x,y)
			r2,g2,b2,_ := newColour.RGBA()
			r2,g2,b2 = r2 >> 8, g2 >> 8, b2 >> 8
			
			newColor := color.RGBA{
				R: uint8(math.Abs(float64(r)-float64(r2))),
				G: uint8(math.Abs(float64(g)-float64(g2))),
				B: uint8(math.Abs(float64(b)-float64(b2))),
				A: uint8(a),
			}
			NewImage.Set(x, y, newColor)
		}
	}
	return  newFromImage(originalImg, NewImage, "Image Difference")
}

func (originalImg *OurImage) ChangeMap(imageIn *OurImage) *OurImage { // TODO it makes a copy
	b := originalImg.canvasImage.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	T := 20.0

	for y := 0; y < originalImg.canvasImage.Image.Bounds().Dy(); y++ {
		for x := 0; x < originalImg.canvasImage.Image.Bounds().Dx(); x++ {
			oldColour := originalImg.canvasImage.Image.At(x, y)
			r,g,b,a := oldColour.RGBA()
			r,g,b = r >> 8, g >> 8, b >> 8

			newColour := imageIn.canvasImage.Image.At(x,y)
			r2,g2,b2,_ := newColour.RGBA()
			r2,g2,b2 = r2 >> 8, g2 >> 8, b2 >> 8
			grey := 0.222*float64(r) + 0.707*float64(g) + 0.071*float64(b)
			grey2 := 0.222*float64(r2) + 0.707*float64(g2) + 0.071*float64(b2)
			difference := math.Abs(grey2-grey)

			var newColor color.RGBA
			if difference > T{
			newColor = color.RGBA{
				R: 255,
				G: 0,
				B: 0,
				A: uint8(a),
			}
		} else {
			newColor = color.RGBA{
				R: uint8(r),
				G: uint8(g),
				B: uint8(b),
				A: uint8(a),
			}
		}
		
			NewImage.Set(x, y, newColor)
		}
	}
	return  newFromImage(originalImg, NewImage, "Image Difference")
}

func Max(x, y int) int {
	if x < y {
			return y
	}
	return x
}