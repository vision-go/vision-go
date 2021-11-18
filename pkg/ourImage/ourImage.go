package ourimage

import (
	"image"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	histogram "github.com/vision-go/vision-go/pkg/histogram"
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

	ROIcallback       func(*OurImage)
	closeTabsCallback func(int)

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

func (ourimage *OurImage) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(ourimage.canvasImage)
}

func NewFromPath(path, name string, statusBar *widget.Label, w fyne.Window, ROIcallback func(*OurImage), closeTabsCallback func(int)) (*OurImage, error) {
	img := &OurImage{}
	img.name = name
	img.statusBar = statusBar
	img.mainWindow = w
	img.ROIcallback = ROIcallback
	img.closeTabsCallback = closeTabsCallback
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
	img.closeTabsCallback = ourImage.closeTabsCallback
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
