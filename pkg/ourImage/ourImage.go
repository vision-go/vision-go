package ourimage

import (
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	histogram "github.com/vision-go/vision-go/pkg/histogram"
)

type OurImage struct {
	widget.BaseWidget
	Image      *canvas.Image
	format     string
	statusBar  *widget.Label

	HistogramR histogram.Histogram
	HistogramG histogram.Histogram
	HistogramB histogram.Histogram
	Histogram  histogram.Histogram

	HistogramAccumulativeR histogram.Histogram
	HistogramAccumulativeG histogram.Histogram
	HistogramAccumulativeB histogram.Histogram
	HistogramAccumulative histogram.Histogram

	HistogramNormalizedR histogram.HistogramNormalized
	HistogramNormalizedG histogram.HistogramNormalized
	HistogramNormalizedB histogram.HistogramNormalized
	HistogramNormalized histogram.HistogramNormalized
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

	makeHistogram(&img)

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
	makeHistogram(&newOurImage)
	return newOurImage
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
	for i := 0; i < image.Image.Image.Bounds().Dx(); i++ {
		for j := 0; j < image.Image.Image.Bounds().Dy(); j++ {
			r, g, b, a := image.Image.Image.At(i, j).RGBA()
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
	for index, _ := range image.Histogram.Values{
		for i := 0; i < index; i++{	
			image.HistogramAccumulativeR.Values[index] += image.HistogramR.At(i)
			image.HistogramAccumulativeG.Values[index] += image.HistogramG.At(i)
			image.HistogramAccumulativeB.Values[index] += image.HistogramB.At(i)
			image.HistogramAccumulative.Values[index] += image.Histogram.At(i)
		}
	}
		for i := 0; i < 256; i++{	
			image.HistogramNormalized.Values[i] /= float64(image.Image.Image.Bounds().Dx() * image.Image.Image.Bounds().Dy())
			image.HistogramNormalizedR.Values[i] /= float64(image.Image.Image.Bounds().Dx() * image.Image.Image.Bounds().Dy())
			image.HistogramNormalizedG.Values[i] /= float64(image.Image.Image.Bounds().Dx() * image.Image.Image.Bounds().Dy())
			image.HistogramNormalizedB.Values[i] /= float64(image.Image.Image.Bounds().Dx() * image.Image.Image.Bounds().Dy())
		}

}
