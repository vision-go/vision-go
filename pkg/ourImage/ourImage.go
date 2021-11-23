package ourimage

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"golang.org/x/image/tiff"
	"math"
	"os"
	"strings"

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
	entropy            float64
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
    fmt.Println("No debería")
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

func makeHistogram(image *OurImage) {
	for i := 0; i < image.canvasImage.Image.Bounds().Dx(); i++ {
		for j := 0; j < image.canvasImage.Image.Bounds().Dy(); j++ {
			r, g, b, a := image.canvasImage.Image.At(i, j).RGBA()
			if a != 0 {
				r, g, b = r>>8, g>>8, b>>8
				image.HistogramR[r] = image.HistogramR.At(int(r)) + 1
				image.HistogramG[g] = image.HistogramG.At(int(g)) + 1
				image.HistogramB[b] = image.HistogramG.At(int(b)) + 1

				grey := 0.222*float64(r) + 0.707*float64(g) + 0.071*float64(b) // PAL
				image.Histogram[int(math.Round(grey))] = image.Histogram.At(int(math.Round(grey))) + 1
			}
		}
	}
	for index := range image.Histogram {
		for i := 0; i < index; i++ {
			image.HistogramAccumulativeR[index] += image.HistogramR.At(i)
			image.HistogramAccumulativeG[index] += image.HistogramG.At(i)
			image.HistogramAccumulativeB[index] += image.HistogramB.At(i)
			image.HistogramAccumulative[index] += image.Histogram.At(i)
		}
	}
	for i := 0; i < 256; i++ {
		image.HistogramNormalized[i] = float64(image.Histogram[i]) / float64(image.size)
		image.HistogramNormalizedR[i] = float64(image.HistogramR[i]) / float64(image.size)
		image.HistogramNormalizedG[i] = float64(image.HistogramG[i]) / float64(image.size)
		image.HistogramNormalizedB[i] = float64(image.HistogramB[i]) / float64(image.size)
	}
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
	} else if format == "tif" || format == "tiff" {
    return tiff.Encode(file, img.canvasImage.Image, nil)
  }
	return fmt.Errorf("incorrrect format")
}

func (ourimage *OurImage) calculateBrightness() (value float64) {
	for color, count := range ourimage.Histogram {
		value += float64(color * count)
	}
	return value / float64(ourimage.size)
}

func (ourimage *OurImage) calculateContrast(brightness float64) (value float64) {
	for color, count := range ourimage.Histogram {
		value += float64(count) * (float64(color) - brightness) * (float64(color) - brightness)
	}
	return math.Sqrt(value / float64(ourimage.size))
}

func (ourimage *OurImage) calculateMinAndMaxColor() (min, max int) {
	for color, count := range ourimage.Histogram {
		if count != 0 {
			min = color
			break
		}
	}
	for max = ourimage.Histogram.Len() - 1; max >= 0; max-- {
		if ourimage.Histogram.At(max) != 0 {
			break
		}
	}
	return
}

func (ourimage *OurImage) calculateEntropyAndNumberOfColors() (float64, int) {
	var sum float64
	var numberOfColors int
	for _, count := range ourimage.Histogram {
		if count != 0 {
			numberOfColors++
		}
	}
	for _, count := range ourimage.Histogram {
		if count != 0 {
			probability := 1 / float64(numberOfColors)
			sum += probability * math.Log2(probability)
		}
	}
	return sum * -1, numberOfColors
}
