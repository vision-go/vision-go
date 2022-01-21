package userinterface

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"sort"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/disintegration/imaging"
	"github.com/dustin/go-humanize"
	"github.com/vision-go/vision-go/pkg/histogram"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"

	ourimage "github.com/vision-go/vision-go/pkg/ourImage"
)

func (ui *UI) equializationOp() {
	if ui.tabs.SelectedIndex() == -1 {
		dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
		return
	}
	ui.newImage(ui.tabsElements[ui.tabs.SelectedIndex()].Equalization()) // TODO Improve name
}

func (ui *UI) negativeOp() {
	currentImage, err := ui.getCurrentImage()
	if err != nil {
		dialog.ShowError(err, ui.MainWindow)
		return
	}
	ui.newImage(currentImage.Negative())
}

func (ui *UI) monochromeOp() {
	currentImage, err := ui.getCurrentImage()
	if err != nil {
		dialog.ShowError(err, ui.MainWindow)
		return
	}
	ui.newImage(currentImage.Monochrome())
}

func (ui *UI) adjustBrightnessAndContrastOp() {
	currentImage, err := ui.getCurrentImage()
	if err != nil {
		dialog.ShowError(err, ui.MainWindow)
		return
	}
	originalPreview := imaging.Resize(currentImage.CanvasImage().Image, 500, 0, imaging.NearestNeighbor)
	previewImg := canvas.NewImageFromImage(originalPreview)
	previewImg.SetMinSize(fyne.NewSize(500, 500)) // TODO dynamic size
	brightnessValue, contrastValue := binding.NewFloat(), binding.NewFloat()
	brightnessLabel, contrastLabel :=
		widget.NewLabelWithData(binding.FloatToStringWithFormat(brightnessValue, "%v")),
		widget.NewLabelWithData(binding.FloatToStringWithFormat(contrastValue, "%v"))
	brightnessSlider, contrastSlider :=
		widget.NewSliderWithData(0, 255, brightnessValue),
		widget.NewSliderWithData(0, 255, contrastValue)
	brightnessSlider.SetValue(currentImage.Brightness())
	contrastSlider.SetValue(currentImage.Contrast())
	brightnessSlider.OnChanged = func(value float64) {
		brightnessValue.Set(value)
		newBrightness, _ := brightnessValue.Get()
		newContrast, _ := contrastValue.Get()
		previewImg.Image = ourimage.BrightnessAndContrastPreview(originalPreview, currentImage.Brightness(), currentImage.Contrast(), newBrightness, newContrast)
		previewImg.Refresh()
	}
	contrastSlider.OnChanged = func(value float64) {
		contrastValue.Set(value)
		newBrightness, _ := brightnessValue.Get()
		newContrast, _ := contrastValue.Get()
		previewImg.Image = ourimage.BrightnessAndContrastPreview(originalPreview, currentImage.Brightness(), currentImage.Contrast(), newBrightness, newContrast)
		previewImg.Refresh()
	}
	content := container.NewGridWithColumns(2, container.NewGridWithRows(4, container.NewCenter(brightnessLabel), brightnessSlider, container.NewCenter(contrastLabel), contrastSlider), previewImg)
	dialog.ShowCustomConfirm("Adjust Brightness and Contrast", "Ok", "Cancel", content,
		func(choice bool) {
			if !choice {
				return
			}
			brightness, err := brightnessValue.Get()
			if err != nil {
				dialog.ShowError(err, ui.MainWindow)
			}
			contrast, err := contrastValue.Get()
			if err != nil {
				dialog.ShowError(err, ui.MainWindow)
			}
			ui.newImage(currentImage.BrightnessAndContrast(brightness, contrast))
		},
		ui.MainWindow)
}

func (ui *UI) gammaCorrectionOp() {
	currentImage, err := ui.getCurrentImage()
	if err != nil {
		dialog.ShowError(err, ui.MainWindow)
		return
	}
	entry := widget.NewEntry()
	entry.Validator = func(value string) error {
		valueFloat, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		if valueFloat < 0.05 || valueFloat > 20 {
			return fmt.Errorf("gamma must be between values 0.05 and 20")
		}
		return nil
	}
	form := []*widget.FormItem{widget.NewFormItem("Gamma", entry)}
	dialog.ShowForm("Select gamma", "Ok", "Cancel", form,
		func(choice bool) {
			if !choice {
				return
			}
			gamma, _ := strconv.ParseFloat(entry.Text, 64) // No need to check thanks to validator
			ui.newImage(currentImage.GammaCorrection(gamma))
		},
		ui.MainWindow)
}

func (ui *UI) linearTransformationOp() {
	currentImage, err := ui.getCurrentImage()
	if err != nil {
		dialog.ShowError(err, ui.MainWindow)
		return
	}
	entryForNumberOfPoints := widget.NewEntry()
	entryForNumberOfPoints.Validator = func(value string) error {
		valueInt, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		if valueInt < 2 {
			return fmt.Errorf("the number of points must greater than two")
		}
		return nil
	}
	dialog.ShowForm("How many points?", "OK", "Cancel",
		[]*widget.FormItem{widget.NewFormItem("", entryForNumberOfPoints)},
		func(choice bool) {
			if !choice {
				return
			}
			var canvasPoints []fyne.CanvasObject
			var points []*histogram.Point
			pointsN, _ := strconv.Atoi(entryForNumberOfPoints.Text) // Error already checked in Validator
			graph := canvas.NewImageFromResource(theme.ViewFullScreenIcon())
			graph.SetMinSize(fyne.NewSize(500, 500))
			updateGraph := func() {
				validatedPoints := make([]histogram.Point, 0, len(points))
				for _, point := range points {
					if err := point.Validate(); err != nil {
						continue
					}
					validatedPoints = append(validatedPoints, *point)
				}
				sort.Slice(validatedPoints, func(i, j int) bool {
					return validatedPoints[i].X < validatedPoints[j].X
				})
				if len(validatedPoints) >= 1 {
					if validatedPoints[0].X != 0 {
						validatedPoints = append([]histogram.Point{{X: 0, Y: 0}}, validatedPoints...)
					}
					if validatedPoints[len(validatedPoints)-1].X != 255 {
						validatedPoints = append(validatedPoints, histogram.Point{X: 255, Y: 255})
					}
				} else {
					validatedPoints = append([]histogram.Point{{X: 0, Y: 0}}, validatedPoints...)
					validatedPoints = append(validatedPoints, histogram.Point{X: 255, Y: 255})
				}
				newGraph, err := createGraph(validatedPoints)
				if err != nil {
					dialog.ShowError(err, ui.MainWindow)
					return
				}
				graph.Resource = nil
				graph.Image = newGraph
				graph.Refresh()
			}
			for i := 0; i < pointsN; i++ {
				rawPoint, canvasPoint := histogram.NewPoint(i, updateGraph)
				points = append(points, rawPoint)
				canvasPoints = append(canvasPoints, canvasPoint)
			}
			content := container.NewHBox(container.NewVScroll(container.NewCenter(container.NewVBox(canvasPoints...))), graph)
			dialog.ShowCustomConfirm("Linear Transformation", "OK", "Cancel", content,
				func(choice bool) {
					if !choice {
						return
					}
					for _, point := range points {
						if point.Validate() != nil {
							dialog.ShowError(point.Validate(), ui.MainWindow)
							return
						}
					}
					ui.newImage(currentImage.LinearTransformation(points))
				},
				ui.MainWindow)
		},
		ui.MainWindow)
}

func (ui *UI) infoView() {
	currentImage, err := ui.getCurrentImage()
	if err != nil {
		dialog.ShowError(err, ui.MainWindow)
		return
	}
	format := currentImage.Format()
	size := currentImage.Dimensions()
	message := fmt.Sprintf("Format: %v\n Size: %v bytes (%v x %v)\n", format, humanize.Bytes(uint64(size.X*size.Y)), size.X, size.Y)
	minColor, maxColor := currentImage.MinAndMaxColor()
	message += fmt.Sprintf("Range: [%v, %v]", minColor, maxColor)
	message += "\nBrightness: " + fmt.Sprintf("%f", currentImage.Brightness())
	message += "\nContrast: " + fmt.Sprintf("%f", currentImage.Contrast())
	entropy, numberOfColors := currentImage.EntropyAndNumberOfColors()
	message += "\nEntropy: " + fmt.Sprintf("%f", entropy) + " with " + strconv.Itoa(numberOfColors) + " diferent colors"
	dialog.ShowInformation("Information", message, ui.MainWindow)
}

func (ui *UI) histogram() {
	if ui.tabs.SelectedIndex() == -1 {
		dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
		return
	}
	a := ui.App.NewWindow(ui.tabs.Selected().Text + " || (Histogram)")
	a.Resize(fyne.NewSize(500, 500))

	image := ui.calculateHistogramGraph(convertToFloat(ui.tabsElements[ui.tabs.SelectedIndex()].Histogram[:]), drawing.ColorBlack)
	imageR := ui.calculateHistogramGraph(convertToFloat(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramR[:]), drawing.ColorRed)
	imageG := ui.calculateHistogramGraph(convertToFloat(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramG[:]), drawing.ColorGreen)
	imageB := ui.calculateHistogramGraph(convertToFloat(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramB[:]), drawing.ColorBlue)

	image1 := canvas.NewImageFromImage(image)
	image2 := canvas.NewImageFromImage(imageR)
	image3 := canvas.NewImageFromImage(imageG)
	image4 := canvas.NewImageFromImage(imageB)

	content := container.New(layout.NewAdaptiveGridLayout(2), image1, image2, image3, image4)

	a.SetContent(content)
	a.Show()
}

func (ui *UI) accumulativeHistogram() {
	if ui.tabs.SelectedIndex() == -1 {
		dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
		return
	}
	a := ui.App.NewWindow(ui.tabs.Selected().Text + " || (AccumulativeHistogram)")
	a.Resize(fyne.NewSize(500, 500))

	image := ui.calculateHistogramGraph(convertToFloat(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramAccumulative[:]), drawing.ColorBlack)
	imageR := ui.calculateHistogramGraph(convertToFloat(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramAccumulativeR[:]), drawing.ColorRed)
	imageG := ui.calculateHistogramGraph(convertToFloat(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramAccumulativeG[:]), drawing.ColorGreen)
	imageB := ui.calculateHistogramGraph(convertToFloat(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramAccumulativeB[:]), drawing.ColorBlue)

	image1 := canvas.NewImageFromImage(image)
	image2 := canvas.NewImageFromImage(imageR)
	image3 := canvas.NewImageFromImage(imageG)
	image4 := canvas.NewImageFromImage(imageB)

	content := container.New(layout.NewAdaptiveGridLayout(2), image1, image2, image3, image4)

	a.SetContent(content)
	a.Show()
}

func (ui *UI) normalizedHistogram() {
	currentImage, err := ui.getCurrentImage()
	if err != nil {
		dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
		return
	}
	a := ui.App.NewWindow(ui.tabs.Selected().Text + " || (NormalizedHistogram)")
	a.Resize(fyne.NewSize(500, 500))

	image := ui.calculateHistogramGraph(currentImage.HistogramNormalized[:], drawing.ColorBlack)
	imageR := ui.calculateHistogramGraph(currentImage.HistogramNormalizedR[:], drawing.ColorRed)
	imageG := ui.calculateHistogramGraph(currentImage.HistogramNormalizedG[:], drawing.ColorGreen)
	imageB := ui.calculateHistogramGraph(currentImage.HistogramNormalizedB[:], drawing.ColorBlue)

	image1 := canvas.NewImageFromImage(image)
	image2 := canvas.NewImageFromImage(imageR)
	image3 := canvas.NewImageFromImage(imageG)
	image4 := canvas.NewImageFromImage(imageB)

	content := container.New(layout.NewAdaptiveGridLayout(2), image1, image2, image3, image4)

	a.SetContent(content)
	a.Show()
}

func (ui *UI) calculateHistogramGraph(valuesY []float64, color drawing.Color) image.Image {
	var indexValues []float64
	strokeColor := color
	fillColor := color
	fillColor.A = 128

	for i := 0; i < 256; i++ {
		indexValues = append(indexValues, float64(i))
	}
	graph := chart.Chart{
		Series: []chart.Series{
			chart.ContinuousSeries{
				Style: chart.Style{
					StrokeColor: strokeColor,
					FillColor:   fillColor,
				},
				XValues: indexValues,
				YValues: valuesY,
			},
		},
	}
	collector := &chart.ImageWriter{}
	graph.Render(chart.PNG, collector)

	image, err := collector.Image()
	if err != nil {
		log.Fatal(err)
	}
	return image
}

func (ui *UI) histogramEqual() {
	currentImage, err := ui.getCurrentImage()
	if err != nil {
		dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
		return
	}
	dialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, ui.MainWindow)
			return
		}
		if reader == nil {
			return
		}
		img, err := ourimage.NewFromPath(reader.URI().Path(), reader.URI().Name(),
			ui.label, ui.MainWindow, ui.ROIcallback, ui.closeTabsCallback)
		if err != nil {
			dialog.ShowError(err, ui.MainWindow)
		}
		ui.newImage(currentImage.HistogramIgualation(img))
	}, ui.MainWindow)
	dialog.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpeg", ".jpg", ".tfe", ".tfi"}))
	dialog.Show()

}

func (ui *UI) imgDifference() {
	currentImage, err := ui.getCurrentImage()
	if err != nil {
		dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
		return
	}
	dialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, ui.MainWindow)
			return
		}
		if reader == nil {
			return
		}
		img, err := ourimage.NewFromPath(reader.URI().Path(), reader.URI().Name(),
			ui.label, ui.MainWindow, ui.ROIcallback, ui.closeTabsCallback)
		if err != nil {
			dialog.ShowError(err, ui.MainWindow)
		}
		img, err = currentImage.ImageDiference(img)
		if err != nil {
			dialog.ShowError(err, ui.MainWindow)
		}
		ui.newImage(img)
	}, ui.MainWindow)
	dialog.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpeg", ".jpg", ".tfe", ".tif"}))
	dialog.Show()
}

// TODO refactor open file DRY
func (ui *UI) imgChangeMap() {
	currentImage, err := ui.getCurrentImage()
	if err != nil {
		dialog.ShowError(err, ui.MainWindow)
		return
	}
	entry := widget.NewEntry()
	entry.Validator = func(value string) error {
		if value == "" {
			return fmt.Errorf("please give us a value")
		}
		valueInt, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		if valueInt < 0 || valueInt > 255 {
			return fmt.Errorf("the values must be integers in the range [0, 255]")
		}
		return nil
	}
	var colorPicked color.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	colorPreview := canvas.NewRectangle(color.RGBA{R: 255, G: 0, B: 0, A: 255})
	colorSelectionButton := widget.NewButton("Pick Color", func() {
		colorSelectionDialog := dialog.NewColorPicker("Select a color", "Prueba", func(c color.Color) {
			colorPicked = c
			colorPreview.FillColor = colorPicked
			colorPreview.Refresh()
		}, ui.MainWindow)
		colorSelectionDialog.Advanced = true
		colorSelectionDialog.Show()
	})
	content := container.NewGridWithRows(2, container.NewGridWithColumns(2, widget.NewLabel("T: "), entry), container.NewGridWithColumns(2, colorSelectionButton, colorPreview))
	dialog.ShowCustomConfirm("Select T value: ", "Ok", "Cancel", content,
		func(choice bool) {
			if !choice {
				return
			}
			if err := entry.Validate(); err != nil {
				dialog.ShowError(err, ui.MainWindow)
				return
			}
			tValue, _ := strconv.Atoi(entry.Text) // No need to check thanks to validator
			dialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
				if err != nil {
					dialog.ShowError(err, ui.MainWindow)
					return
				}
				if reader == nil {
					return
				}
				img, err := ourimage.NewFromPath(reader.URI().Path(), reader.URI().Name(),
					ui.label, ui.MainWindow, ui.ROIcallback, ui.closeTabsCallback)
				if err != nil {
					dialog.ShowError(err, ui.MainWindow)
				}
				img, err = currentImage.ChangeMap(img, colorPicked, tValue)
				if err != nil {
					dialog.ShowError(err, ui.MainWindow)
					return
				}
				ui.newImage(img)
			}, ui.MainWindow)
			dialog.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpeg", ".jpg", ".tfe", ".tfi"}))
			dialog.Show()
		},
		ui.MainWindow)
}

func convertToFloat(f32 []int) []float64 {
	f64 := make([]float64, len(f32))
	for i, f := range f32 {
		f64[i] = float64(f)
	}
	return f64
}

func createGraph(points []histogram.Point) (image.Image, error) {
	Xs := make([]float64, len(points))
	Ys := make([]float64, len(points))
	for i, point := range points {
		Xs[i] = float64(point.X)
		Ys[i] = float64(point.Y)
	}
	graph := chart.Chart{
		Series: []chart.Series{
			chart.ContinuousSeries{
				Style: chart.Style{
					StrokeColor: drawing.ColorRed,
				},
				XValues: Xs,
				YValues: Ys,
			},
		},
	}
	graph.YAxis.Range = &chart.ContinuousRange{Min: 0, Max: 255} // Fix flat graph not being displayed
	collector := &chart.ImageWriter{}
	graph.Render(chart.PNG, collector)
	image, err := collector.Image()
	if err != nil {
		return nil, err
	}
	return image, nil
}

func (ui *UI) horizontal() {
	currentImage, err := ui.getCurrentImage()
	if err != nil {
		dialog.ShowError(err, ui.MainWindow)
		return
	}
	ui.newImage(currentImage.HorizontalMirror())
}

func (ui *UI) vertical() {
	currentImage, err := ui.getCurrentImage()
	if err != nil {
		dialog.ShowError(err, ui.MainWindow)
		return
	}
	ui.newImage(currentImage.VerticalMirror())
}

func (ui *UI) rotateRight() {
	currentImage, err := ui.getCurrentImage()
	if err != nil {
		dialog.ShowError(err, ui.MainWindow)
		return
	}
	ui.newImage(currentImage.RotateRight())
}

func (ui *UI) rotateLeft() {
	currentImage, err := ui.getCurrentImage()
	if err != nil {
		dialog.ShowError(err, ui.MainWindow)
		return
	}
	ui.newImage(currentImage.RotateLeft())
}

func (ui *UI) transpose() {
	currentImage, err := ui.getCurrentImage()
	if err != nil {
		dialog.ShowError(err, ui.MainWindow)
		return
	}
	ui.newImage(currentImage.Transpose())
}

func (ui *UI) rescaling() {
	currentImage, err := ui.getCurrentImage()
	if err != nil {
		dialog.ShowError(err, ui.MainWindow)
		return
	}
	entry := widget.NewEntry()
	entry.Validator = func(value string) error {
		valueFloat, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		if valueFloat < 1 || valueFloat > 500 {
			return fmt.Errorf("rescalingfactor must be between values 1 and 500")
		}
		return nil
	}
	var typeSelect bool
	radio := widget.NewRadioGroup([]string{"VMP", "Bilineal"}, func(value string) {
		if value == "VMP" {
			typeSelect = true
		} else {
			typeSelect = false
		}
	})
	form := []*widget.FormItem{
		widget.NewFormItem("Scale(in %)", entry),
		widget.NewFormItem("Type", radio),
	}
	dialog.ShowForm("Select Scale", "Ok", "Cancel", form,
		func(choice bool) {
			if !choice {
				return
			}
			rescalingFactor, _ := strconv.ParseFloat(entry.Text, 64) // No need to check thanks to validator
			ui.newImage(currentImage.Rescaling(rescalingFactor/100, typeSelect))
		},
		ui.MainWindow)
}

func (ui *UI) rotateAndPrint() {
	currentImage, err := ui.getCurrentImage()
	if err != nil {
		dialog.ShowError(err, ui.MainWindow)
		return
	}
	entry := widget.NewEntry()
	form := []*widget.FormItem{
		widget.NewFormItem("Angular grades", entry),
	}
	dialog.ShowForm("Select angular", "Ok", "Cancel", form,
		func(choice bool) {
			if !choice {
				return
			}
			angle, _ := strconv.ParseFloat(entry.Text, 64)
			ui.newImage(currentImage.RotateAndPrint(angle))
		},
		ui.MainWindow)
}

func (ui *UI) rotate() {
	currentImage, err := ui.getCurrentImage()
	if err != nil {
		dialog.ShowError(err, ui.MainWindow)
		return
	}
	entry := widget.NewEntry()
	selection := widget.NewSelect([]string{"Bilineal", "NearestNeighbor"}, func(string) { return })
	selection.SetSelectedIndex(0)
	form := []*widget.FormItem{
		widget.NewFormItem("Angular grades", entry),
		widget.NewFormItem("Strategy", selection),
	}
	dialog.ShowForm("Select angular", "Ok", "Cancel", form,
		func(choice bool) {
			if !choice {
				return
			}
			angle, _ := strconv.ParseFloat(entry.Text, 64)
			ui.newImage(currentImage.Rotate(angle, selection.SelectedIndex()))
		},
		ui.MainWindow)
}
