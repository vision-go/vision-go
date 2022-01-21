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
	NewImage := BrightnessAndContrastPreview(originalImg.canvasImage.Image, originalImg.brightness, originalImg.contrast, brightness, contrast)
	return originalImg.newFromImage(NewImage, "B/C")
}

func BrightnessAndContrastPreview(img image.Image, oldbr, oldctr, newbr, newctr float64) image.Image {
	A := newctr / oldctr
	B := newbr - A*oldbr
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
	b := img.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for x := 0; x < img.Bounds().Dx(); x++ {
		for y := 0; y < img.Bounds().Dy(); y++ {
			r, g, b, a := img.At(x, y).RGBA()
			r, g, b = r>>8, g>>8, b>>8
			NewImage.Set(x, y, color.RGBA{R: localLookUpTable[r].Y,
				G: localLookUpTable[g].Y, B: localLookUpTable[b].Y, A: uint8(a)})
		}
	}
	return NewImage
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
	b2 := imageIn.canvasImage.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	sizeF := float64(b.Dx() * b.Dy())
	sizeF2 := float64(b2.Dx() * b2.Dy())
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
			if (float64(PoR[a]) / sizeF) > (float64(PrR[j]) / sizeF2) {
				break
			}
		}

		for j := M - 1; j >= 0; j-- {
			lookUpTableArrayG[a] = j
			if (float64(PoG[a]) / sizeF) > (float64(PrG[j]) / sizeF2) {
				break
			}
		}

		for j := M - 1; j >= 0; j-- {
			lookUpTableArrayB[a] = j
			if (float64(PoB[a]) / sizeF) > (float64(PrB[j]) / sizeF2) {
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

func (originalImg *OurImage) ChangeMap(imageIn *OurImage, colour color.Color, T int) (*OurImage, error) {
	if originalImg.canvasImage.Image.Bounds() != imageIn.canvasImage.Image.Bounds() {
		return nil, fmt.Errorf("images must have the same dimensions")
	}
	b := originalImg.canvasImage.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := 0; y < originalImg.canvasImage.Image.Bounds().Dy(); y++ {
		for x := 0; x < originalImg.canvasImage.Image.Bounds().Dx(); x++ {
			oldColour := originalImg.canvasImage.Image.At(x, y)
			r, g, b, _ := oldColour.RGBA()
			r, g, b = r>>8, g>>8, b>>8

			r2, g2, b2, _ := imageIn.canvasImage.Image.At(x, y).RGBA()
			r2, g2, b2 = r2>>8, g2>>8, b2>>8
			grey := 0.222*float64(r) + 0.707*float64(g) + 0.071*float64(b)
			grey2 := 0.222*float64(r2) + 0.707*float64(g2) + 0.071*float64(b2)
			difference := math.Abs(grey2 - grey)

			if difference > float64(T) {
				oldColour = colour
			}
			NewImage.Set(x, y, oldColour)
		}
	}
	return originalImg.newFromImage(NewImage, "Image Difference"), nil
}

func (originalImg *OurImage) HorizontalMirror() *OurImage {
	b := originalImg.canvasImage.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))

	for y := 0; y < originalImg.canvasImage.Image.Bounds().Dy(); y++ {
		for x := 0; x < originalImg.canvasImage.Image.Bounds().Dx(); x++ {
			oldColour := originalImg.canvasImage.Image.At(originalImg.canvasImage.Image.Bounds().Dx()-1-x, y)
			NewImage.Set(x, y, oldColour)
		}
	}
	return originalImg.newFromImage(NewImage, "Horizontal-Mirror")
}

func (originalImg *OurImage) VerticalMirror() *OurImage {
	b := originalImg.canvasImage.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))

	for y := 0; y < originalImg.canvasImage.Image.Bounds().Dy(); y++ {
		for x := 0; x < originalImg.canvasImage.Image.Bounds().Dx(); x++ {
			oldColour := originalImg.canvasImage.Image.At(x, originalImg.canvasImage.Image.Bounds().Dy()-1-y)
			NewImage.Set(x, y, oldColour)
		}
	}
	return originalImg.newFromImage(NewImage, "Vertical-Mirror")
}

func (originalImg *OurImage) RotateRight() *OurImage {
	b := originalImg.canvasImage.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dy(), b.Dx()))

	for y := 0; y < originalImg.canvasImage.Image.Bounds().Dy(); y++ {
		for x := 0; x < originalImg.canvasImage.Image.Bounds().Dx(); x++ {
			oldColour := originalImg.canvasImage.Image.At(x, y)
			NewImage.Set(originalImg.canvasImage.Image.Bounds().Dy()-1-y, x, oldColour)
		}
	}
	return originalImg.newFromImage(NewImage, "Rotate-Right")
}

func (originalImg *OurImage) RotateLeft() *OurImage {
	b := originalImg.canvasImage.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dy(), b.Dx()))

	for y := 0; y < originalImg.canvasImage.Image.Bounds().Dy(); y++ {
		for x := 0; x < originalImg.canvasImage.Image.Bounds().Dx(); x++ {
			oldColour := originalImg.canvasImage.Image.At(x, y)
			NewImage.Set(y, originalImg.canvasImage.Image.Bounds().Dx()-1-x, oldColour)
		}
	}
	return originalImg.newFromImage(NewImage, "Rotate-Right")
}

func (originalImg *OurImage) Transpose() *OurImage {
	b := originalImg.canvasImage.Image.Bounds()
	NewImage := image.NewRGBA(image.Rect(0, 0, b.Dy(), b.Dx()))

	for y := 0; y < originalImg.canvasImage.Image.Bounds().Dy(); y++ {
		for x := 0; x < originalImg.canvasImage.Image.Bounds().Dx(); x++ {
			oldColour := originalImg.canvasImage.Image.At(x, y)
			NewImage.Set(y, x, oldColour)
		}
	}
	return originalImg.newFromImage(NewImage, "Transpose")
}

func (originalImg *OurImage) Rescaling(rescalingFactor float64, VMP bool) *OurImage {
	b := originalImg.canvasImage.Image.Bounds()
	width := int(math.Round(float64(b.Dx()) * (rescalingFactor)))
	height := int(math.Round(float64(b.Dy()) * (rescalingFactor)))
	NewImage := image.NewRGBA(image.Rect(0, 0, width, height))

	var Colour color.Color
	for y := 0; y <= height; y++ {
		for x := 0; x <= width; x++ {
			cordX := float64(x) / (rescalingFactor)
			cordY := float64(y) / (rescalingFactor)
			if VMP {
				indexI := int(math.Round(cordX))
				indexJ := int(math.Round(cordY))

				Colour = originalImg.canvasImage.Image.At(indexI, indexJ)
			} else {
				indexICeil := int(math.Ceil(cordX))
				indexIFloor := int(math.Floor(cordX))
				indexJCeil := int(math.Ceil(cordY))
				indexJFloor := int(math.Floor(cordY))

				p := uint32(cordX - math.Floor(cordX))
				q := uint32(cordY - math.Floor(cordY))
				A := originalImg.canvasImage.Image.At(indexIFloor, indexJCeil)
				D := originalImg.canvasImage.Image.At(indexICeil, indexJCeil)
				C := originalImg.canvasImage.Image.At(indexIFloor, indexJFloor)
				B := originalImg.canvasImage.Image.At(indexICeil, indexJFloor)

				ra, ga, ba, aa := A.RGBA()
				rb, gb, bb, ab := B.RGBA()
				rc, gc, bc, ac := C.RGBA()
				rd, gd, bd, ad := D.RGBA()
				ra, rb, rc, rd = ra>>8, rb>>8, rc>>8, rd>>8
				ga, gb, gc, gd = ga>>8, gb>>8, gc>>8, gd>>8
				ba, bb, bc, bd = ba>>8, bb>>8, bc>>8, bd>>8
				aa, ab, ac, ad = aa>>8, ab>>8, ac>>8, ad>>8

				rG := rc + uint32((rd-rc)*p) + uint32((ra-rc)*q) + uint32((rb+rc-ra-rd)*p*q)
				gG := gc + uint32((gd-gc)*p) + uint32((ga-gc)*q) + uint32((gb+gc-ga-gd)*p*q)
				bG := bc + uint32((bd-bc)*p) + uint32((ba-bc)*q) + uint32((bb+bc-ba-bd)*p*q)
				aG := ac + uint32((ad-ac)*p) + uint32((aa-ac)*q) + uint32((ab+ac-aa-ad)*p*q)
				Colour = color.RGBA{
					R: uint8(rG),
					G: uint8(gG),
					B: uint8(bG),
					A: uint8(aG),
				}
			}
			NewImage.Set(x, y, Colour)
		}
	}
	return originalImg.newFromImage(NewImage, "Rescaling")
}

type Point struct {
	X, Y float64
}

func (originalImg *OurImage) RotateAndPrint (angle float64) *OurImage {
	b := originalImg.canvasImage.Image.Bounds()
	angleRadian := -angle * math.Pi / 180
	rotateX := func(x, y int) float64 {
		return float64(x)*math.Cos(angleRadian) - float64(y)*math.Sin(angleRadian)
	}
	rotateY := func(x, y int) float64 {
		return float64(x)*math.Sin(angleRadian) + float64(y)*math.Cos(angleRadian)
	}
	A := Point{X: rotateX(0, 0), Y: rotateY(0, 0)}
	B := Point{X: rotateX(b.Dx(), 0), Y: rotateY(b.Dx(), 0)}
	C := Point{X: rotateX(0, b.Dy()), Y: rotateY(0, b.Dy())}
	D := Point{X: rotateX(b.Dx(), b.Dy()), Y: rotateY(b.Dx(), b.Dy())}
	minX := math.Min(math.Min(A.X, B.X), math.Min(C.X, D.X))
	maxX := math.Max(math.Max(A.X, B.X), math.Max(C.X, D.X))
	minY := math.Min(math.Min(A.Y, B.Y), math.Min(C.Y, D.Y))
	maxY := math.Max(math.Max(A.Y, B.Y), math.Max(C.Y, D.Y))

	newImage := image.NewRGBA(image.Rect(0, 0, int(math.Ceil(math.Abs(maxX-minX))), int(math.Ceil(math.Abs(maxY-minY)))))
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			newImage.Set(int(math.Round(rotateX(x, y)+math.Abs(minX))), int(math.Round(rotateY(x, y)+math.Abs(minY))), originalImg.canvasImage.Image.At(x, y))
		}
	}
	return originalImg.newFromImage(newImage, "Rotate and print")
}

func (originalImg *OurImage) Rotate(angle float64, selection int) *OurImage {
  if (selection == 0) {
    return originalImg.RotateVMP(angle)
  }
  return originalImg.RotateBilineal(angle)
}

func (originalImg *OurImage) RotateVMP (angle float64) *OurImage {
	b := originalImg.canvasImage.Image.Bounds()
	angleRadian := -angle * math.Pi / 180
	rotateX_TD := func(x, y int) float64 {
		return float64(x)*math.Cos(angleRadian) - float64(y)*math.Sin(angleRadian)
	}
	rotateY_TD := func(x, y int) float64 {
		return float64(x)*math.Sin(angleRadian) + float64(y)*math.Cos(angleRadian)
	}
	A := Point{X: rotateX_TD(0, 0), Y: rotateY_TD(0, 0)}
	B := Point{X: rotateX_TD(b.Dx(), 0), Y: rotateY_TD(b.Dx(), 0)}
	C := Point{X: rotateX_TD(0, b.Dy()), Y: rotateY_TD(0, b.Dy())}
	D := Point{X: rotateX_TD(b.Dx(), b.Dy()), Y: rotateY_TD(b.Dx(), b.Dy())}
	minX := math.Min(math.Min(A.X, B.X), math.Min(C.X, D.X))
	maxX := math.Max(math.Max(A.X, B.X), math.Max(C.X, D.X))
	minY := math.Min(math.Min(A.Y, B.Y), math.Min(C.Y, D.Y))
	maxY := math.Max(math.Max(A.Y, B.Y), math.Max(C.Y, D.Y))

	rotateX_TI := func(x, y int) float64 {
		return float64(x)*math.Cos(-angleRadian) - float64(y)*math.Sin(-angleRadian)
	}
	rotateY_TI := func(x, y int) float64 {
		return float64(x)*math.Sin(-angleRadian) + float64(y)*math.Cos(-angleRadian)
	}
	newImage := image.NewRGBA(image.Rect(0, 0, int(math.Ceil(math.Abs(maxX-minX))), int(math.Ceil(math.Abs(maxY-minY)))))
  getDistance := func(x1, y1, x2, y2 int) float64 {
    return math.Sqrt(float64((x2-x1)*(x2-x1)) + float64((y2-y1)*(y2-y1)))
  }
  for y := 0; y < newImage.Rect.Dy(); y++ {
    for x := 0; x < newImage.Rect.Dx(); x++ {
      rotatedX := math.Round(rotateX_TI(x-int(math.Abs(minX)), y-int(math.Abs(minY))))
      rotatedY := math.Round(rotateY_TI(x-int(math.Abs(minX)), y-int(math.Abs(minY))))
      points := [4][2]int{{int(rotatedX), int(rotatedY+1)},
        {int(rotatedX+1), int(rotatedY+1)},
        {int(rotatedX), int(rotatedY)},
        {int(rotatedX+1), int(rotatedY)}}
      var distances [4]float64
      for i, point := range points {
        distances[i] = getDistance(x, y, point[0], point[1])
      }
      nearestDistanceIndex := 0
      for i, distance := range distances[1:] {
        if distance < distances[nearestDistanceIndex] {
          nearestDistanceIndex = i
        }
      }
      if rotatedX >= 0 && rotatedX < float64(b.Dx()) && rotatedY >= 0 && rotatedY < float64(b.Dy()) {
        newImage.Set(x, y, originalImg.canvasImage.Image.At(int(rotatedX), int(rotatedY)))
      }
		}
	}
	return originalImg.newFromImage(newImage, "Rotate(VMP)")
}

func (originalImg *OurImage) RotateBilineal (angle float64) *OurImage {
	b := originalImg.canvasImage.Image.Bounds()
	angleRadian := -angle * math.Pi / 180
	rotateX_TD := func(x, y int) float64 {
		return float64(x)*math.Cos(angleRadian) - float64(y)*math.Sin(angleRadian)
	}
	rotateY_TD := func(x, y int) float64 {
		return float64(x)*math.Sin(angleRadian) + float64(y)*math.Cos(angleRadian)
	}
	A := Point{X: rotateX_TD(0, 0), Y: rotateY_TD(0, 0)}
	B := Point{X: rotateX_TD(b.Dx(), 0), Y: rotateY_TD(b.Dx(), 0)}
	C := Point{X: rotateX_TD(0, b.Dy()), Y: rotateY_TD(0, b.Dy())}
	D := Point{X: rotateX_TD(b.Dx(), b.Dy()), Y: rotateY_TD(b.Dx(), b.Dy())}
	minX := math.Min(math.Min(A.X, B.X), math.Min(C.X, D.X))
	maxX := math.Max(math.Max(A.X, B.X), math.Max(C.X, D.X))
	minY := math.Min(math.Min(A.Y, B.Y), math.Min(C.Y, D.Y))
	maxY := math.Max(math.Max(A.Y, B.Y), math.Max(C.Y, D.Y))

	rotateX_TI := func(x, y int) float64 {
		return float64(x)*math.Cos(-angleRadian) - float64(y)*math.Sin(-angleRadian)
	}
	rotateY_TI := func(x, y int) float64 {
		return float64(x)*math.Sin(-angleRadian) + float64(y)*math.Cos(-angleRadian)
	}
	newImage := image.NewRGBA(image.Rect(0, 0, int(math.Ceil(math.Abs(maxX-minX))), int(math.Ceil(math.Abs(maxY-minY)))))
 
  for y := 0; y < newImage.Rect.Dy(); y++ {
    for x := 0; x < newImage.Rect.Dx(); x++ {
      rotatedX := math.Round(rotateX_TI(x-int(math.Abs(minX)), y-int(math.Abs(minY))))
      rotatedY := math.Round(rotateY_TI(x-int(math.Abs(minX)), y-int(math.Abs(minY))))
			
			indexICeil := int(math.Ceil(rotatedX))
				indexIFloor := int(math.Floor(rotatedX))
				indexJCeil := int(math.Ceil(rotatedY))
				indexJFloor := int(math.Floor(rotatedY))

				p := uint32(rotatedX - math.Floor(rotatedX))
				q := uint32(rotatedY - math.Floor(rotatedY))
				A := originalImg.canvasImage.Image.At(indexIFloor, indexJCeil)
				D := originalImg.canvasImage.Image.At(indexICeil, indexJCeil)
				C := originalImg.canvasImage.Image.At(indexIFloor, indexJFloor)
				B := originalImg.canvasImage.Image.At(indexICeil, indexJFloor)

				ra, ga, ba, aa := A.RGBA()
				rb, gb, bb, ab := B.RGBA()
				rc, gc, bc, ac := C.RGBA()
				rd, gd, bd, ad := D.RGBA()
				ra, rb, rc, rd = ra>>8, rb>>8, rc>>8, rd>>8
				ga, gb, gc, gd = ga>>8, gb>>8, gc>>8, gd>>8
				ba, bb, bc, bd = ba>>8, bb>>8, bc>>8, bd>>8
				aa, ab, ac, ad = aa>>8, ab>>8, ac>>8, ad>>8

				rG := rc + uint32((rd-rc)*p) + uint32((ra-rc)*q) + uint32((rb+rc-ra-rd)*p*q)
				gG := gc + uint32((gd-gc)*p) + uint32((ga-gc)*q) + uint32((gb+gc-ga-gd)*p*q)
				bG := bc + uint32((bd-bc)*p) + uint32((ba-bc)*q) + uint32((bb+bc-ba-bd)*p*q)
				aG := ac + uint32((ad-ac)*p) + uint32((aa-ac)*q) + uint32((ab+ac-aa-ad)*p*q)
				Colour := color.RGBA{
					R: uint8(rG),
					G: uint8(gG),
					B: uint8(bG),
					A: uint8(aG),
				}
      if rotatedX >= 0 && rotatedX < float64(b.Dx()) && rotatedY >= 0 && rotatedY < float64(b.Dy()) {
        newImage.Set(x, y, Colour)
      }
		}
	}
  return originalImg.newFromImage(newImage, "Rotate (Bilineal)")
}
