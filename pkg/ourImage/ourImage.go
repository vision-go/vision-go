package ourimage

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	histogram "github.com/vision-go/vision-go/pkg/histogram"
	lookUpTable "github.com/vision-go/vision-go/pkg/look-up-table"
)

type OurImage struct {
	widget.BaseWidget
	name               string
	canvasImage        *canvas.Image
	format             string
	brightness         float64
	contrast           float64
	entropy            int
	numberOfColors     int
	minColor, maxColor int
	size               int
	statusBar          *widget.Label
	mainWindow         fyne.Window
	rectangle          image.Rectangle
	ROIcallback        func(*OurImage)
	newImageCallback   func(*OurImage)

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

func (self *OurImage) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(self.canvasImage)
}

func NewFromPath(path, name string, statusBar *widget.Label, w fyne.Window, ROIcallback, newImageCallback func(*OurImage)) (*OurImage, error) {
	img := &OurImage{}
	img.name = name
	img.statusBar = statusBar
	img.mainWindow = w
	img.ROIcallback = ROIcallback
	img.newImageCallback = newImageCallback
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

	img.size = img.canvasImage.Image.Bounds().Dx() * img.canvasImage.Image.Bounds().Dy()
	makeHistogram(img)
	img.minColor, img.maxColor = img.calculateMinAndMaxColor()
	img.brightness = img.calculateBrightness()
	img.contrast = img.calculateContrast(img.brightness)
	img.entropy, img.numberOfColors = img.calculateEntropyAndNumberOfColors()
	return img, nil
}

func (ourImage *OurImage) newFromImage(newImage image.Image, actionForName string) *OurImage {
	img := &OurImage{}
	img.name = ourImage.addOperationToName(actionForName)
	img.statusBar = ourImage.statusBar
	img.mainWindow = ourImage.mainWindow
	img.ROIcallback = ourImage.ROIcallback
	img.newImageCallback = ourImage.newImageCallback
	img.ExtendBaseWidget(img)
	img.canvasImage = canvas.NewImageFromImage(newImage)
	img.canvasImage.FillMode = canvas.ImageFillOriginal
	img.size = img.canvasImage.Image.Bounds().Dx() * img.canvasImage.Image.Bounds().Dy()
	makeHistogram(img)
	img.minColor, img.maxColor = img.calculateMinAndMaxColor()
	img.brightness = img.calculateBrightness()
	img.contrast = img.calculateContrast(img.brightness)
	img.entropy, img.numberOfColors = img.calculateEntropyAndNumberOfColors()
	return img
}

func (img *OurImage) addOperationToName(actionForName string) string {
	if actionForName == "" {
		return img.name
	}
	actionForName = "(" + actionForName + ")"
	pointIndex := strings.LastIndex(img.name, ".")
	if pointIndex == -1 {
		return img.name + actionForName
	}
	return img.name[:pointIndex] + actionForName + img.name[pointIndex:]
}

func (img *OurImage) Save(file *os.File, format string) error {
	if format == "png" {
		return png.Encode(file, img.canvasImage.Image)
	} else if format == "jpeg" || format == "jpg" {
		return jpeg.Encode(file, img.canvasImage.Image, nil)
	}
	return fmt.Errorf("Incorrrect format")
}

// Functions
func (ourimage *OurImage) calculateBrightness() (value float64) {
	for color, count := range ourimage.Histogram.Values {
		value += float64(color * count)
	}
	return value / float64(ourimage.size)
}

func (ourimage *OurImage) calculateContrast(brightness float64) (value float64) {
	for _, count := range ourimage.Histogram.Values {
		value += (float64(count) - brightness) * (float64(count) - brightness)
	}
	return math.Sqrt(value / float64(ourimage.size))
}

func (ourimage *OurImage) calculateMinAndMaxColor() (min, max int) {
	for color, count := range ourimage.Histogram.Values {
		if count != 0 {
			min = color
			break
		}
	}
	for max = ourimage.Histogram.Len(); max > 0; max-- {
		if ourimage.Histogram.At(max) != 0 {
			break
		}
	}
	return
}

func (ourimage *OurImage) calculateEntropyAndNumberOfColors() (int, int) {
	var sum float64
	var numberOfColors int
	for _, count := range ourimage.Histogram.Values {
		if count != 0 {
			numberOfColors++
		}
	}
	for _, count := range ourimage.Histogram.Values {
		if count != 0 {
			probability := 1 / float64(numberOfColors)
			sum += probability * math.Log2(probability)
		}
	}
	return int(math.Ceil(sum * -1)), numberOfColors
}

func (originalImg *OurImage) Negative() *OurImage {
	b := originalImg.canvasImage.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))

	for y := 0; y < originalImg.canvasImage.Image.Bounds().Dy(); y++ {
		for x := 0; x < originalImg.canvasImage.Image.Bounds().Dx(); x++ {
			oldColour := originalImg.canvasImage.Image.At(x, y)
			NewImage.Set(x, y, lookUpTable.RGBA(oldColour, lookUpTable.Negative))
		}
	}
	return originalImg.newFromImage(NewImage, "Negative")
}

func (originalImg *OurImage) Monochrome() *OurImage {
	b := originalImg.canvasImage.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))

	for y := 0; y < originalImg.canvasImage.Image.Bounds().Dy(); y++ {
		for x := 0; x < originalImg.canvasImage.Image.Bounds().Dx(); x++ {
			oldColour := originalImg.canvasImage.Image.At(x, y)
			r, g, b, _ := oldColour.RGBA()
			NewImage.Set(x, y, color.Gray{Y: uint8(0.222*float32(r>>8) + 0.707*float32(g>>8) + 0.071*float32(b>>8))}) // PAL
		}
	}
	return originalImg.newFromImage(NewImage, "Monochrome")
}

func (originalImg *OurImage) ROI(rect image.Rectangle) *OurImage {
	b := rect.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := 0; y < rect.Dy(); y++ {
		for x := 0; x < rect.Dx(); x++ {
			NewImage.Set(x, y, originalImg.canvasImage.Image.At(x+rect.Min.X, y+rect.Min.Y))
		}
	}
	return originalImg.newFromImage(NewImage, "ROI")
}

func (originalImg *OurImage) BrightnessAndContrast(brightness, contrast float64) *OurImage {
	b := originalImg.canvasImage.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))

	A := contrast / originalImg.contrast
	B := brightness - A*originalImg.brightness
	localLookUpTable := make([]color.Gray, 256)
	for colour := range localLookUpTable {
		vOut := A*float64(colour) + B
		if vOut > 255 {
			localLookUpTable[colour] = color.Gray{255}
		} else if vOut < 0 {
			localLookUpTable[colour] = color.Gray{0}
		} else {
			localLookUpTable[colour] = color.Gray{uint8(A*float64(colour) + B)}
		}
	}
	for x := 0; x < originalImg.canvasImage.Image.Bounds().Dx(); x++ {
		for y := 0; y < originalImg.canvasImage.Image.Bounds().Dy(); y++ {
			r, g, b, a := originalImg.canvasImage.Image.At(x, y).RGBA()
			r, g, b = r>>8, g>>8, b>>8
			NewImage.Set(x, y, color.RGBA{R: localLookUpTable[r].Y,
				G: localLookUpTable[g].Y, B: localLookUpTable[b].Y, A: uint8(a)})
		}
	}
	return originalImg.newFromImage(NewImage, "B/C")
}

func (originalImg *OurImage) GammaCorrection(gamma float64) *OurImage {
	b := originalImg.canvasImage.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	localLookUpTable := make([]color.Gray, 256)
	fmt.Println("gamma: ", gamma)
	for colour := range localLookUpTable {
		a := float64(colour) / 255
		b := math.Pow(a, gamma)
		localLookUpTable[colour] = color.Gray{uint8(math.Round(b * 255))}
	}
	for x := 0; x < originalImg.canvasImage.Image.Bounds().Dx(); x++ {
		for y := 0; y < originalImg.canvasImage.Image.Bounds().Dy(); y++ {
			r, g, b, a := originalImg.canvasImage.Image.At(x, y).RGBA()
			r, g, b = r>>8, g>>8, b>>8
			NewImage.Set(x, y, color.RGBA{R: localLookUpTable[r].Y,
				G: localLookUpTable[g].Y, B: localLookUpTable[b].Y, A: uint8(a)})
		}
	}
	return originalImg.newFromImage(NewImage, "Gamma")
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
			r, g, b, a := image.canvasImage.Image.At(i, j).RGBA()
			if a != 0 {
				r, g, b = r>>8, g>>8, b>>8
				image.HistogramR.Values[r] = image.HistogramR.At(int(r)) + 1
				image.HistogramG.Values[g] = image.HistogramG.At(int(r)) + 1
				image.HistogramB.Values[b] = image.HistogramG.At(int(r)) + 1

				grey := 0.222*float64(r) + 0.707*float64(g) + 0.071*float64(b) // PAL
				image.Histogram.Values[int(math.Round(grey))] = image.Histogram.At(int(math.Round(grey))) + 1
			}
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
