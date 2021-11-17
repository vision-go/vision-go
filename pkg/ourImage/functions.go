package ourimage

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/vision-go/vision-go/pkg/histogram"
	lookUpTable "github.com/vision-go/vision-go/pkg/look-up-table"
)

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

func (ourimage *OurImage) calculateEntropyAndNumberOfColors() (int, int) {
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
			r, g, b, a := oldColour.RGBA()
			if a == 0 {
				continue
			}
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
			localLookUpTable[colour] = color.Gray{uint8(vOut)}
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

func (ourimage *OurImage) LinearTransformation(points []*histogram.Point) *OurImage {
	sort.Slice(points, func(i, j int) bool {
		return points[i].X_ < points[j].X_
	})
	if points[0].X_ != 0 {
		points = append([]*histogram.Point{{X_: 0, Y_: 0}}, points...)
	}
	if points[len(points)-1].X_ != 255 {
		points = append(points, &histogram.Point{X_: 255, Y_: 255})
	}
	var localLookUpTable [256]color.Gray
	calculate := func(p1, p2 histogram.Point) {
		m := float64(p2.Y_-p1.Y_) / float64(p2.X_-p1.X_)
		n := float64(p1.Y_) - m*float64(p1.X_)
		for x := p1.X_; x < p2.X_; x++ {
			localLookUpTable[x] = color.Gray{uint8(math.Round(m*float64(x) + n))}
		}
	}
	i := 0
	for j := 1; i < len(points)-1; i, j = i+2, j+2 {
		calculate(*points[i], *points[j])
	}
	if i == len(points)-1 { // Even number of points
		calculate(*points[len(points)-2], *points[len(points)-1])
	}
	b := ourimage.canvasImage.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for x := 0; x < ourimage.canvasImage.Image.Bounds().Dx(); x++ {
		for y := 0; y < ourimage.canvasImage.Image.Bounds().Dy(); y++ {
			r, g, b, a := ourimage.canvasImage.Image.At(x, y).RGBA()
			r, g, b = r>>8, g>>8, b>>8
			NewImage.Set(x, y, color.RGBA{R: localLookUpTable[r].Y,
				G: localLookUpTable[g].Y, B: localLookUpTable[b].Y, A: uint8(a)})
		}
	}
	return ourimage.newFromImage(NewImage, "LinearTrans")
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
