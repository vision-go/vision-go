package ourimage

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sort"

	"github.com/vision-go/vision-go/pkg/histogram"
	lookUpTable "github.com/vision-go/vision-go/pkg/look-up-table"
)

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
		if points[i].X == points[j].X {
			return points[i].Y > points[j].Y
		}
		return points[i].X < points[j].X
	})
	if points[0].X != 0 {
		points = append([]*histogram.Point{{X: 0, Y: 0}}, points...)
	}
	if points[len(points)-1].X != 255 {
		points = append(points, &histogram.Point{X: 255, Y: 255})
	}
	var localLookUpTable [256]color.Gray
	for i, j := 0, 1; j < len(points); i, j = i+1, j+1 {
		p1, p2 := *points[i], *points[j]
		if p2.X == p1.X {
			p2.X++
		}
		m := float64(p2.Y-p1.Y) / float64(p2.X-p1.X)
		n := float64(p1.Y) - m*float64(p1.X)
		for x := p1.X; x < p2.X; x++ {
			localLookUpTable[x] = color.Gray{uint8(math.Round(m*float64(x) + n))}
		}
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

func (originalImg *OurImage) Equalization() *OurImage {
	b := originalImg.canvasImage.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	size := b.Dx() * b.Dy()
	var lookUpTableArrayR [256]int
	var lookUpTableArrayG [256]int
	var lookUpTableArrayB [256]int

	for i := 0; i < 256; i++ {
		lookUpTableArrayR[i] = int(math.Round(math.Max(0, (float64(originalImg.HistogramAccumulativeR.At(i)*256)/float64(size))-1)))
		lookUpTableArrayG[i] = int(math.Round(math.Max(0, (float64(originalImg.HistogramAccumulativeG.At(i)*256)/float64(size))-1)))
		lookUpTableArrayB[i] = int(math.Round(math.Max(0, (float64(originalImg.HistogramAccumulativeB.At(i)*256)/float64(size))-1)))
	}
	for y := 0; y < originalImg.canvasImage.Image.Bounds().Dy(); y++ {
		for x := 0; x < originalImg.canvasImage.Image.Bounds().Dx(); x++ {
			oldColour := originalImg.canvasImage.Image.At(x, y)
			r, g, b, a := oldColour.RGBA()
			r, g, b = r>>8, g>>8, b>>8

			newColor := color.RGBA{
				R: uint8(lookUpTableArrayR[r]),
				G: uint8(lookUpTableArrayG[g]),
				B: uint8(lookUpTableArrayB[b]),
				A: uint8(a),
			}
			NewImage.Set(x, y, newColor)
		}
	}
	return originalImg.newFromImage(NewImage, "Ecualization")
}

func (originalImg *OurImage) HistogramIgualation(imageIn *OurImage) *OurImage {
	b := originalImg.canvasImage.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	sizeF := float64(b.Dx() * b.Dy())
	var lookUpTableArrayR [256]int
	var lookUpTableArrayG [256]int
	var lookUpTableArrayB [256]int

	M := 256
	PoR := originalImg.HistogramAccumulativeR
	PoG := originalImg.HistogramAccumulativeG
	PoB := originalImg.HistogramAccumulativeB
	PrR := imageIn.HistogramAccumulativeR
	PrG := imageIn.HistogramAccumulativeG
	PrB := imageIn.HistogramAccumulativeB

	for a := range lookUpTableArrayR {
		for j := M - 1; j >= 0; j-- {
			lookUpTableArrayR[a] = j
			if (float64(PoR[a]) / sizeF) > (float64(PrR[j]) / sizeF) {
				break
			}
		}

		for j := M - 1; j >= 0; j-- {
			lookUpTableArrayG[a] = j
			if (float64(PoG[a]) / sizeF) > (float64(PrG[j]) / sizeF) {
				break
			}
		}

		for j := M - 1; j >= 0; j-- {
			lookUpTableArrayB[a] = j
			if (float64(PoB[a]) / sizeF) > (float64(PrB[j]) / sizeF) {
				break
			}
		}
	}
	for y := 0; y < originalImg.canvasImage.Image.Bounds().Dy(); y++ {
		for x := 0; x < originalImg.canvasImage.Image.Bounds().Dx(); x++ {
			oldColour := originalImg.canvasImage.Image.At(x, y)
			r, g, b, a := oldColour.RGBA()
			r, g, b = r>>8, g>>8, b>>8

			newColor := color.RGBA{
				R: uint8(lookUpTableArrayR[r]),
				G: uint8(lookUpTableArrayG[g]),
				B: uint8(lookUpTableArrayB[b]),
				A: uint8(a),
			}
			NewImage.Set(x, y, newColor)
		}
	}
	return originalImg.newFromImage(NewImage, "Histogram Igualated")
}

func (originalImg *OurImage) ImageDiference(imageIn *OurImage) (*OurImage, error) {
	if originalImg.canvasImage.Image.Bounds() != imageIn.canvasImage.Image.Bounds() {
		return nil, fmt.Errorf("images must have the same dimensions")
	}
	b := originalImg.canvasImage.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))

	for y := 0; y < originalImg.canvasImage.Image.Bounds().Dy(); y++ {
		for x := 0; x < originalImg.canvasImage.Image.Bounds().Dx(); x++ {
			oldColour := originalImg.canvasImage.Image.At(x, y)
			r, g, b, a := oldColour.RGBA()
			r, g, b = r>>8, g>>8, b>>8

			newColour := imageIn.canvasImage.Image.At(x, y)
			r2, g2, b2, _ := newColour.RGBA()
			r2, g2, b2 = r2>>8, g2>>8, b2>>8

			newColor := color.RGBA{
				R: uint8(math.Abs(float64(r) - float64(r2))),
				G: uint8(math.Abs(float64(g) - float64(g2))),
				B: uint8(math.Abs(float64(b) - float64(b2))),
				A: uint8(a),
			}
			NewImage.Set(x, y, newColor)
		}
	}
	return originalImg.newFromImage(NewImage, "Image Difference"), nil
}

func (originalImg *OurImage) ChangeMap(imageIn *OurImage) (*OurImage, error) {
	if originalImg.canvasImage.Image.Bounds() != imageIn.canvasImage.Image.Bounds() {
		return nil, fmt.Errorf("images must have the same dimensions")
	}
	b := originalImg.canvasImage.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	T := 30.0

	for y := 0; y < originalImg.canvasImage.Image.Bounds().Dy(); y++ {
		for x := 0; x < originalImg.canvasImage.Image.Bounds().Dx(); x++ {
			oldColour := originalImg.canvasImage.Image.At(x, y)
			r, g, b, a := oldColour.RGBA()
			r, g, b = r>>8, g>>8, b>>8

			newColour := imageIn.canvasImage.Image.At(x, y)
			r2, g2, b2, _ := newColour.RGBA()
			r2, g2, b2 = r2>>8, g2>>8, b2>>8
			grey := 0.222*float64(r) + 0.707*float64(g) + 0.071*float64(b)
			grey2 := 0.222*float64(r2) + 0.707*float64(g2) + 0.071*float64(b2)
			difference := math.Abs(grey2 - grey)

			var newColor color.RGBA
			if difference > T {
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
	return originalImg.newFromImage(NewImage, "Image Difference"), nil
}
